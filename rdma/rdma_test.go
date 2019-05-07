package rdma

import (
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
	"testing"
)

//func TestNewRdmaDevicePlugin(t *testing.T) {
//	println("testing RDMA new")
//}

func TestGetDevices(t *testing.T) {
	println("testing device poking with the command line")

	var devs []*pluginapi.Device
	devs = GetDevices()
}
