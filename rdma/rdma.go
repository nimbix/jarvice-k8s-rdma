// Copyright (c) 2021, Nimbix, Inc.
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

// Package rdma Separate package for rdma access code, help with tests
package rdma

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const (
	IBDevicePath = "/dev/infiniband/"
	//IBDevicePath = "/tmp/" //TESTING local only
	IBCMDevicePrefix   = "rdma_cm"
	IBVerbDevicePrefix = "uverb"
	IBCMDevicePath     = IBDevicePath + IBCMDevicePrefix
)

// IBDevice Simple type using the name as the ID, possibly have to pull the GUID out for uniqueness
type IBDevice struct {
	Name string
	Path string
}

// Match the device name from a device path
// e.g. /dev/infiniband/uverbs1 with a substring of uverb
func validDevicePrefix(path string) bool {
	prefixes := []string{IBCMDevicePrefix, IBVerbDevicePrefix}
	match := false
	for _, sub := range prefixes {
		if strings.Contains(path, sub) {
			match = true
		}
	}
	return match
}

// GetIBFileList Return all the device files
//
// $ ll /dev/infiniband/
// drwxr-xr-x  2 root root      140 Feb 25 13:21 ./
// drwxr-xr-x 22 root root     4540 Apr 13 11:46 ../
// crw-------  1 root root 231,  64 Feb 25 13:21 issm0
// crw-rw-rw-  1 root root  10,  54 Feb 25 13:21 rdma_cm
// crw-rw-rw-  1 root root 231, 224 Feb 25 13:21 ucm0
// crw-------  1 root root 231,   0 Feb 25 13:21 umad0
// crw-rw-rw-  1 root root 231, 192 Feb 25 13:21 uverbs0
func GetIBFileList() ([]os.FileInfo, error) {
	log.Print("Getting list of device files from: ", IBDevicePath)

	// Call ls on the /dev/infiniband/ directory
	files, err := ioutil.ReadDir(IBDevicePath)
	if err != nil {
		log.Printf("failed getting Infiniband device files: %v", err)
	}

	return files, err
}

// GetDevices Get all the Infiniband devices from the files
func GetDevices() []IBDevice {
	var devs []IBDevice

	// Get the list of device files
	files, err := GetIBFileList()
	if err != nil {
		log.Fatal("No RDMA devices found on node")
		return nil
	}

	// for each IB device file, make a local device type
	//   only append devices we want: uverbs and rdma_cm
	for _, file := range files {
		if validDevicePrefix(file.Name()) {
			devs = append(devs, IBDevice{
				Name: file.Name(),
				Path: IBDevicePath + file.Name(),
			})
		}
	}

	return devs
}

//func DeviceExists(devray []*pluginapi.Device, id string) bool {
// for the plugin device, pull the rdma_cm file and verify it exists
//func DeviceExists(devray *pluginapi.Device, id string) bool {
//
//	devray.
//	//for _, dev := range devray.ID {
//	//	if dev.ID == id {
//	//		return true
//	//	}
//	//}
//	return false
//}

//func watchXIDs(ctx context.Context, devs []*pluginapi.Device, xids chan<- *pluginapi.Device) {
//	for {
//		select {
//		case <-ctx.Done():
//			return
//		}
//
//		// TODO: check RDMA device healthy status
//	}
//}
