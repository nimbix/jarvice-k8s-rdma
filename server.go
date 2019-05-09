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
	"google.golang.org/grpc"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
	"main/rdma"
)

const (
	resourceName = "jarvice.com/rdma"
	serverSock   = pluginapi.DevicePluginPath + "ibrdma.sock"
	//envDisableHealthChecks = "DP_DISABLE_HEALTHCHECKS"
	//allHealthChecks        = "xids"
)

// RdmaDevicePlugin implements the Kubernetes device plugin API
type RdmaDevicePlugin struct {
	devs   []*pluginapi.Device
	socket string

	stop   chan interface{}
	health chan *pluginapi.Device

	server *grpc.Server
}

// NewRdmaDevicePlugin returns an initialized RdmaDevicePlugin
func NewRdmaDevicePlugin() *RdmaDevicePlugin {
	return &RdmaDevicePlugin{
		devs:   rdma.GetDevices(),
		socket: serverSock,

		stop:   make(chan interface{}),
		health: make(chan *pluginapi.Device),
	}
}

// Allocate returns the list of devices to expose in the container
// NB: must NOT allocate if devices have already been allocated on the node: TODO
//func (m *RdmaDevicePlugin) Allocate(ctx context.Context, r *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
//	devs := m.devs
//	responses := pluginapi.AllocateResponse{}
//	var devicesList []*pluginapi.DeviceSpec
//	var knemDeviceName string = "/dev/knem"
//
//	for _, req := range r.ContainerRequests {
//		response := pluginapi.ContainerAllocateResponse{}
//
//		log.Debugf("Request IDs: %v", req.DevicesIDs)
//
//		for _, id := range req.DevicesIDs {
//			if !deviceExists(devs, id) {
//				return nil, fmt.Errorf("invalid allocation request: unknown device: %s", id)
//			}
//
//			var devPath string
//			if dev, ok := m.devices[id]; ok {
//				// TODO: to function
//				devPath = fmt.Sprintf("/dev/infiniband/%s", dev.RdmaDevice.DevName)
//				log.Debugf("device path found: %v", devPath)
//			} else {
//				continue
//			}
//
//			ds := &pluginapi.DeviceSpec{
//				ContainerPath: devPath,
//				HostPath:      devPath,
//				Permissions:   "rw",
//			}
//			devicesList = append(devicesList, ds)
//		}
//		log.Debugf("Devices list from DevicesIDs: %v", devicesList)
//
//		// for /dev/infiniband/rdma_cm
//		rdma_cm_paths := []string{
//			"/dev/infiniband/rdma_cm",
//		}
//		for _, dev := range rdma_cm_paths {
//			devicesList = append(devicesList, &pluginapi.DeviceSpec{
//				ContainerPath: dev,
//				HostPath:      dev,
//				Permissions:   "rw",
//			})
//		}
//
//		// MPI (Intel at least) also requires the use of /dev/knem, add if present
//		if _, err := os.Stat(knemSysfsName); err == nil {
//			// Add the device to the list to mount in the container
//			devicesList = append(devicesList, &pluginapi.DeviceSpec{
//				ContainerPath: knemDeviceName,
//				HostPath:      knemDeviceName,
//				Permissions:   "rw",
//			})
//		}
//		log.Debugf("Devices list after manual additions: %v", devicesList)
//
//		response.Devices = devicesList
//
//		responses.ContainerResponses = append(responses.ContainerResponses, &response)
//	}
//
//	return &responses, nil
//}

// dial establishes the gRPC communication with the registered device plugin.
//func dial(unixSocketPath string, timeout time.Duration) (*grpc.ClientConn, error) {
//	c, err := grpc.Dial(unixSocketPath, grpc.WithInsecure(), grpc.WithBlock(),
//		grpc.WithTimeout(timeout),
//		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
//			return net.DialTimeout("unix", addr, timeout)
//		}),
//	)
//
//	if err != nil {
//		return nil, err
//	}
//
//	return c, nil
//}

// Start starts the gRPC server of the device plugin
//func (m *RdmaDevicePlugin) Start() error {
//	//err := m.cleanup()
//	//if err != nil {
//	//	return err
//	//}
//
//	sock, err := net.Listen("unix", m.socket)
//	if err != nil {
//		return err
//	}
//
//	m.server = grpc.NewServer([]grpc.ServerOption{}...)
//	pluginapi.RegisterDevicePluginServer(m.server, m)
//
//	go m.server.Serve(sock)
//
//	// Wait for server to start by launching a blocking connexion
//	conn, err := dial(m.socket, 5*time.Second)
//	if err != nil {
//		return err
//	}
//	conn.Close()
//
//	//go m.healthcheck()
//
//	return nil
//}

// Serve starts the gRPC server and register the device plugin to Kubelet
//func (m *RdmaDevicePlugin) Serve() error {
//	err := m.Start()
//	if err != nil {
//		log.Printf("Could not start device plugin: %s", err)
//		return err
//	}
//	log.Println("Starting to serve on", m.socket)
//
//	err = m.Register(pluginapi.KubeletSocket, resourceName)
//	if err != nil {
//		log.Printf("Could not register device plugin: %s", err)
//		m.Stop()
//		return err
//	}
//	log.Println("Registered device plugin with Kubelet")
//
//	return nil
//}
//
//func (m *RdmaDevicePlugin) cleanup() error {
//	if err := os.Remove(m.socket); err != nil && !os.IsNotExist(err) {
//		return err
//	}
//
//	return nil
//}

// Stop stops the gRPC server
//func (m *RdmaDevicePlugin) Stop() error {
//	if m.server == nil {
//		return nil
//	}
//
//	m.server.Stop()
//	m.server = nil
//	close(m.stop)
//
//	return m.cleanup()
//}
//
//// Register registers the device plugin for the given resourceName with Kubelet.
//func (m *RdmaDevicePlugin) Register(kubeletEndpoint, resourceName string) error {
//	conn, err := dial(kubeletEndpoint, 5*time.Second)
//	if err != nil {
//		return err
//	}
//	defer conn.Close()
//
//	client := pluginapi.NewRegistrationClient(conn)
//	reqt := &pluginapi.RegisterRequest{
//		Version:      pluginapi.Version,
//		Endpoint:     path.Base(m.socket),
//		ResourceName: resourceName,
//	}
//
//	_, err = client.Register(context.Background(), reqt)
//	if err != nil {
//		return err
//	}
//	return nil
//}
