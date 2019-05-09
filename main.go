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
	"log"
	"main/rdma"
	"main/sysutl"
	"syscall"
)

func main() {

	log.Println("Starting RDMA plugin")

	//log.Println("Try a command")
	//out, err := sysutl.ExecCommand("ls", "/tmp")
	//if err != nil {
	//	log.Fatal("failed command")
	//}
	//log.Printf("ls /tmp output: \n\n%v\n", out.String())

	log.Println("look for /dev/infiniband files")
	ibfiles, err := rdma.GetIBFileList()
	if err != nil {
		log.Fatalf("no IB files found: %v", err)
	} else {
		for _, file := range ibfiles {
			log.Println(file.Name())
		}
	}

	log.Println("Fetching devices")
	if len(rdma.GetDevices()) == 0 {
		log.Println("No devices found...waiting indefinitely")
		//select {}
	}

	log.Println("Starting signal handler")
	sigs := sysutl.SignalWatcher(syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	restart := true
	//var devicePlugin *RdmaDevicePlugin

LOOP:
	for {
		if restart {
			//devicePlugin = NewRdmaDevicePlugin()
			//if err := devicePlugin.Serve(); err != nil {
			//	log.Println("Could not contact Kubelet, retrying. Did you enable the device plugin feature gate?")
			//} else {
			//	restart = false
			//}
		}

		// Respond to events
		select {
		case s := <-sigs:
			switch s {
			case syscall.SIGHUP:
				log.Println("Received SIGHUP, restarting")
				restart = true
			default:
				log.Printf("Received signal \"%v\", shutting down", s)
				//devicePlugin.Stop()
				break LOOP
			}
		}
	}
}
