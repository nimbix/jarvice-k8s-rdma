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

// NB: comments from api.pb.go in the k8s source included to clarify API calls

package main

import (
	"fmt"
	"github.com/google/uuid"
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
)

// RDMADevicePlugin implements the Kubernetes device plugin API
type RDMADevicePlugin struct {
	plugindev *pluginapi.Device
	socket    string
	devices   map[string]rdma.IBDevice

	stop   chan interface{}
	health chan *pluginapi.Device

	server *grpc.Server
}

// NewRDMADevicePlugin returns an initialized RDMADevicePlugin
func NewRDMADevicePlugin() *RDMADevicePlugin {
	devices := rdma.GetDevices()
	if devices == nil {
		log.Print("Error getting RDMA devices")
		return nil
	}

	// generate a UUID as the ID for the plugin device
	pluginuuid, err := uuid.NewUUID()
	if err != nil {
		log.Println("Unable to generate a UUID for the plugin device")
		return nil
	}
	log.Println("New plugin UUID: ", pluginuuid.String())

	// Make a map of IB device files to an ID, a UUID
	//var devs []*pluginapi.Device
	//ibdevmap := make(map[string]rdma.IBDevice)
	//for _, device := range devices {
	//	id := device.Name
	//	devs = append(devs, &pluginapi.Device{
	//		ID:     id,
	//		Health: pluginapi.Healthy,
	//	})
	//	ibdevmap[id] = device
	//}
	var plugindev *pluginapi.Device
	ibdevmap := make(map[string]rdma.IBDevice)
	for _, device := range devices {
		id := device.Name
		ibdevmap[id] = device
	}
	log.Printf("device map for new RDMA plugin object: %v", ibdevmap)

	plugindev = &pluginapi.Device{
		ID:     pluginuuid.String(),
		Health: pluginapi.Healthy,
	}

	return &RDMADevicePlugin{
		plugindev: plugindev,
		socket:    serverSock,
		devices:   ibdevmap,

		stop:   make(chan interface{}),
		health: make(chan *pluginapi.Device),
	}
}

func (rcvr *RDMADevicePlugin) cleanup() error {
	if err := os.Remove(rcvr.socket); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

// Needs to be implemented using some check with ibstat/ibstatus call to card(s)
func (rcvr *RDMADevicePlugin) healthcheck() {

	//ctx, cancel := context.WithCancel(context.Background())
	_, cancel := context.WithCancel(context.Background())

	//var xids chan *pluginapi.Device
	//if !strings.Contains(disableHealthChecks, "xids") {
	//	xids = make(chan *pluginapi.Device)
	//	go watchXIDs(ctx, rcvr.devs, xids)
	//}

	for {
		select {
		case <-rcvr.stop:
			cancel()
			return
			//case dev := <-xids:
			//	rcvr.unhealthy(dev)
		}
	}
}

func (rcvr *RDMADevicePlugin) unhealthy(dev *pluginapi.Device) {
	rcvr.health <- dev
}

// AllocateResponse includes the artifacts that needs to be injected into
// a container for accessing 'deviceIDs' that were mentioned as part of
// 'AllocateRequest'.
// Failure Handling:
// if Kubelet sends an allocation request for dev1 and dev2.
// Allocation on dev1 succeeds but allocation on dev2 fails.
// The Device plugin should send a ListAndWatch update and fail the
// Allocation request

// Allocate returns the list of devices to expose in the container, ie AllocateOnce...
// NB: must NOT allocate if devices have already been allocated on the node: ConfigMap?
// only return 1 k8s plugin api device: allocate ALL IB devices at once, all cards, all device files
//  look for rdma_cm presence
//  grab all the uverbs*
//  optionally find /dev/knem
func (rcvr *RDMADevicePlugin) Allocate(ctx context.Context, reqs *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	//dev := rcvr.plugindev
	responses := pluginapi.AllocateResponse{}
	var devicesList []*pluginapi.DeviceSpec

	for _, req := range reqs.ContainerRequests {
		response := pluginapi.ContainerAllocateResponse{}

		log.Printf("Allocate() called: Request IDs: %v", req.DevicesIDs)

		if _, err := os.Stat(rdma.IBCMDevicePath); err != nil {
			log.Println("No rdma_cm device found, failing Allocate")
			devicesList = nil
			return nil, err
		}

		// kubelet requests the devices (1) it was told of at registration, now build the device file paths
		// and DeviceSpec for mounting the device file paths into the pod container
		for _, id := range req.DevicesIDs {
			log.Print("DeviceID: ", id)
			//if !rdma.DeviceExists(dev, id) {
			//	return nil, fmt.Errorf("invalid allocation request: unknown device: %s", id)
			//} else {
			//	log.Printf("device: %s", id)
			//}

			var devPath string
			for key := range rcvr.devices {
				if dev, ok := rcvr.devices[key]; ok {
					devPath = fmt.Sprintf("/dev/infiniband/%s", dev.Name)
					log.Printf("IB device path found: %v", devPath)
				}

				// DeviceSpec has the paths for mounting the files into a container
				ds := &pluginapi.DeviceSpec{
					ContainerPath: devPath,
					HostPath:      devPath,
					Permissions:   "rw",
				}
				devicesList = append(devicesList, ds)
			}
		}
		log.Printf("Devices list from DevicesIDs: %v", devicesList)

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

// Start creates the gRPC server of the device plugin and starts the server
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

	// needs an error channel
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

	go rcvr.healthcheck()

	return nil
}

// Serve runs the gRPC server and register the device plugin to Kubelet
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
	return &pluginapi.DevicePluginOptions{}, nil
}

// ListAndWatch lists devices and update that list according to the health status
// ListAndWatch returns a stream of List of Devices
// Whenever a Device state change or a Device disappears, ListAndWatch
// returns the new list
func (rcvr *RDMADevicePlugin) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	var plugindevlist []*pluginapi.Device
	plugindevlist = append(plugindevlist, rcvr.plugindev)

	s.Send(&pluginapi.ListAndWatchResponse{Devices: plugindevlist})

	for {
		select {
		case <-rcvr.stop:
			return nil
		case d := <-rcvr.health:
			// FIXME: there is no way to recover from the Unhealthy state.
			d.Health = pluginapi.Unhealthy
			s.Send(&pluginapi.ListAndWatchResponse{Devices: plugindevlist})
		}
	}
}

// PreStartContainer is called, if indicated by Device Plugin during registration phase,
// before each container start. Device plugin can run device specific operations
// such as resetting the device before making devices available to the container
func (rcvr *RDMADevicePlugin) PreStartContainer(context.Context, *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}
