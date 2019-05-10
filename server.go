// Copyright (c) 2019, Nimbix, Inc.
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice,
//    this list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright notice,
//    this list of conditions and the following disclaimer in the documentation
//    and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.
//
// The views and conclusions contained in the software and documentation are
// those of the authors and should not be interpreted as representing official
// policies, either expressed or implied, of Nimbix, Inc.

package main

import (
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
	"log"
	"main/rdma"
	"net"
	"os"
	"path"
	"time"
)

const (
	resourceName   = "jarvice.com/rdma"
	serverSock     = pluginapi.DevicePluginPath + "rdma.sock"
	knemDevicePath = "/dev/knem"
	//envDisableHealthChecks = "DP_DISABLE_HEALTHCHECKS"
	//allHealthChecks        = "xids"
)

// RDMADevicePlugin implements the Kubernetes device plugin API
type RDMADevicePlugin struct {
	devs   []*pluginapi.Device
	socket string

	stop   chan interface{}
	health chan *pluginapi.Device

	server *grpc.Server
}

// NewRDMADevicePlugin returns an initialized RDMADevicePlugin
func NewRDMADevicePlugin() *RDMADevicePlugin {
	return &RDMADevicePlugin{
		devs:   rdma.GetDevices(),
		socket: serverSock,

		stop:   make(chan interface{}),
		health: make(chan *pluginapi.Device),
	}
}

func deviceExists(devray []*pluginapi.Device, id string) bool {
	for _, dev := range devray {
		if dev.ID == id {
			return true
		}
	}
	return false
}

// Allocate returns the list of devices to expose in the container, ie AllocateOnce...
// NB: must NOT allocate if devices have already been allocated on the node: TODO ConfigMap?
//  look for rdma_cm presence
//  grab all the uverbs*
//  optionally find /dev/knem
// TODO: list of devices to allow: uverbs, rdma_cm, knem
func (rcvr *RDMADevicePlugin) Allocate(ctx context.Context, r *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	devs := rcvr.devs
	responses := pluginapi.AllocateResponse{}
	var devicesList []*pluginapi.DeviceSpec

	for _, req := range r.ContainerRequests {
		response := pluginapi.ContainerAllocateResponse{}

		log.Printf("Allocate() called: Request IDs: %v", req.DevicesIDs)

		for _, id := range req.DevicesIDs {
			if !deviceExists(devs, id) {
				return nil, fmt.Errorf("invalid allocation request: unknown device: %s", id)
			} else {
				log.Printf("device: %s", id)
			}

			//var devPath string
			//if dev, ok := rcvr.devices[id]; ok {
			//	// TODO: to function
			//	devPath = fmt.Sprintf("/dev/infiniband/%s", dev.RdmaDevice.DevName)
			//	log.Printf("device path found: %v", devPath)
			//} else {
			//	continue
			//}

			//ds := &pluginapi.DeviceSpec{
			//	ContainerPath: devPath,
			//	HostPath:      devPath,
			//	Permissions:   "rw",
			//}
			//devicesList = append(devicesList, ds)
		}
		log.Printf("Devices list from DevicesIDs: %v", devicesList)

		// for /dev/infiniband/rdma_cm
		if _, err := os.Stat(rdma.IBCMDevicePath); err == nil {
			devicesList = append(devicesList, &pluginapi.DeviceSpec{
				ContainerPath: rdma.IBCMDevicePath,
				HostPath:      rdma.IBCMDevicePath,
				Permissions:   "rw",
			})
		}

		// MPI (Intel at least) also requires the use of /dev/knem, add if present
		if _, err := os.Stat(knemDevicePath); err == nil {
			// Add the device to the list to mount in the container
			devicesList = append(devicesList, &pluginapi.DeviceSpec{
				ContainerPath: knemDevicePath,
				HostPath:      knemDevicePath,
				Permissions:   "rw",
			})
		}
		log.Printf("Devices list after manual additions: %v", devicesList)

		response.Devices = devicesList

		responses.ContainerResponses = append(responses.ContainerResponses, &response)
	}

	return &responses, nil
}

// dial establishes the gRPC communication with the registered device plugin.
func dial(unixSocketPath string, timeout time.Duration) (*grpc.ClientConn, error) {
	c, err := grpc.Dial(unixSocketPath,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(timeout),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}),
	)

	if err != nil {
		return nil, err
	}

	return c, nil
}

// Start starts the gRPC server of the device plugin
func (rcvr *RDMADevicePlugin) Start() error {
	err := rcvr.cleanup()
	if err != nil {
		return err
	}

	sock, err := net.Listen("unix", rcvr.socket)
	if err != nil {
		return err
	}

	rcvr.server = grpc.NewServer([]grpc.ServerOption{}...)
	pluginapi.RegisterDevicePluginServer(rcvr.server, rcvr)

	go rcvr.server.Serve(sock)

	// Wait for server to start by launching a blocking connexion
	conn, err := dial(rcvr.socket, 5*time.Second)
	if err != nil {
		return err
	}
	err = conn.Close()
	if err != nil {
		log.Fatalln("Failed connecting to Kubelet, error closing connection")
		return err
	}

	//go rcvr.healthcheck()

	return nil
}

// Serve starts the gRPC server and register the device plugin to Kubelet
func (rcvr *RDMADevicePlugin) Serve() error {
	err := rcvr.Start()
	if err != nil {
		log.Printf("Could not start device plugin: %s", err)
		return err
	}
	log.Println("Starting to serve on", rcvr.socket)

	err = rcvr.Register(pluginapi.KubeletSocket, resourceName)
	if err != nil {
		log.Printf("Could not register device plugin: %s", err)
		rcvr.Stop()
		return err
	}
	log.Println("Registered device plugin with Kubelet")

	return nil
}

func (rcvr *RDMADevicePlugin) cleanup() error {
	if err := os.Remove(rcvr.socket); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

// Stop stops the gRPC server
func (rcvr *RDMADevicePlugin) Stop() error {
	if rcvr.server == nil {
		return nil
	}

	rcvr.server.Stop()
	rcvr.server = nil
	close(rcvr.stop)

	return rcvr.cleanup()
}

// Register registers the device plugin for the given resourceName with Kubelet.
func (rcvr *RDMADevicePlugin) Register(kubeletEndpoint, resourceName string) error {
	conn, err := dial(kubeletEndpoint, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pluginapi.NewRegistrationClient(conn)
	reqt := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     path.Base(rcvr.socket),
		ResourceName: resourceName,
	}

	_, err = client.Register(context.Background(), reqt)
	if err != nil {
		return err
	}
	return nil
}

func (rcvr *RDMADevicePlugin) GetDevicePluginOptions(context.Context, *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	panic("implement me")
}

func (rcvr *RDMADevicePlugin) ListAndWatch(*pluginapi.Empty, pluginapi.DevicePlugin_ListAndWatchServer) error {
	panic("implement me")
}

func (rcvr *RDMADevicePlugin) PreStartContainer(context.Context, *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	panic("implement me")
}
