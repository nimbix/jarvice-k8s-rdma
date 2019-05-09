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

// Separate package for rdma access code, help with tests
package rdma

import (
	"bytes"
	"log"
	"main/sysutl"

	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
)

const IBRdmaDevicePath = "/dev/infiniband"

func check(err error) {
	if err != nil {
		log.Panicln("Fatal:", err)
	}
}

func GetIBFileList() bytes.Buffer {
	log.Print("Grabbing simple list of device files")

	var devlist bytes.Buffer

	// Call ls on the /dev/infiniband/ directory
	devlist, err := sysutl.ExecCommand("ls", IBRdmaDevicePath)
	if err != nil {
		log.Printf("failed fetching Infiniband device files: %v", err)
	}

	return devlist
}

// Get all the Infiniband devices
//
// $ ll /dev/infiniband/
// drwxr-xr-x  2 root root      140 Feb 25 13:21 ./
// drwxr-xr-x 22 root root     4540 Apr 13 11:46 ../
// crw-------  1 root root 231,  64 Feb 25 13:21 issm0
// crw-rw-rw-  1 root root  10,  54 Feb 25 13:21 rdma_cm
// crw-rw-rw-  1 root root 231, 224 Feb 25 13:21 ucm0
// crw-------  1 root root 231,   0 Feb 25 13:21 umad0
// crw-rw-rw-  1 root root 231, 192 Feb 25 13:21 uverbs0
func GetDevices() []*pluginapi.Device {
	//n, err := nvml.GetDeviceCount()
	//check(err)

	//var n uint = 10
	var devs []*pluginapi.Device
	//for i := uint(0); i < n; i++ {
	//	d, err := nvml.NewDeviceLite(i)
	//	check(err)
	//	devs = append(devs, &pluginapi.Device{
	//		ID:     d.UUID,
	//		Health: pluginapi.Healthy,
	//	})
	//}

	return devs
}

//func GetDevices() ([]Device, error) {
//	var devs []Device
//	// Get all RDMA device list
//	ibvDevList, err := ibverbs.IbvGetDeviceList()
//	if err != nil {
//		return nil, err
//	}
//
//	netDevList, err := GetAllNetDevice()
//	if err != nil {
//		return nil, err
//	}
//	for _, d := range ibvDevList {
//		for _, n := range netDevList {
//			dResource, err := getRdmaDeviceResoure(d.Name)
//			if err != nil {
//				continue
//			}
//			nResource, err := getNetDeviceResoure(n)
//			if err != nil {
//				continue
//			}
//
//			// the same device
//			if bytes.Compare(dResource, nResource) == 0 {
//				devs = append(devs, Device{
//					RdmaDevice: d,
//					NetDevice:  n,
//				})
//			}
//		}
//	}
//	return devs, nil
//}
