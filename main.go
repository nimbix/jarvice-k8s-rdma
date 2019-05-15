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
//
//   This package implements the RDMA device plugin for JARVICE Kubernetes

package main

import (
	"github.com/fsnotify/fsnotify"
	"log"
	"main/rdma"
	"main/sysutl"
	"os"
	"syscall"

	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
)

func main() {
	log.Println("Starting RDMA plugin")

	log.Println("look for Infiniband device files")
	ibfiles, err := rdma.GetIBFileList()
	if err != nil {
		log.Fatalf("no IB files found: %v", err)
	} else {
		for _, file := range ibfiles {
			log.Println(file.Name())
		}
		log.Println()
	}

	log.Println("Fetching devices")
	if len(rdma.GetDevices()) == 0 {
		log.Println("No devices found...waiting indefinitely")
		//select {} TODO: uncomment for release
		select {}
	}

	log.Printf("Starting FS watcher for: %v", pluginapi.DevicePluginPath)
	watcher, err := sysutl.FSWatcher(pluginapi.DevicePluginPath)
	if err != nil {
		log.Println("Failed to created FS watcher")
		os.Exit(1)
	}
	defer watcher.Close()

	log.Println("Starting signal handler")
	sigs := sysutl.SignalWatcher(syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	restart := true
	var devicePlugin *RDMADevicePlugin

LOOP:
	for {
		if restart {
			devicePlugin = NewRDMADevicePlugin()
			log.Print("new devicePlugin ID: ", devicePlugin.plugindev.ID)
			log.Print("devices in devicePlugin: ", devicePlugin.devices)
			if err := devicePlugin.Serve(); err != nil {
				log.Println("Could not contact kubelet, retrying...")
			} else {
				restart = false
			}
		}

		// Respond to events
		select {
		case event := <-watcher.Events:
			if event.Name == pluginapi.KubeletSocket && event.Op&fsnotify.Create == fsnotify.Create {
				log.Printf("notify: %s created, restarting", pluginapi.KubeletSocket)
				restart = true
			}

		case err := <-watcher.Errors:
			log.Printf("notify: %s", err)

		case s := <-sigs:
			switch s {
			case syscall.SIGHUP:
				log.Println("Received SIGHUP, restarting")
				restart = true
			default:
				log.Printf("Received signal \"%v\", shutting down", s)
				_ = devicePlugin.Stop()
				break LOOP
			}
		}
	}
}
