# Infiniband RDMA device plugin for Kubernetes on JARVICE

Simple **Kubernetes** [device plugin](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/) 
with DaemonSet to detect and pass _Infiniband_ devices, 
used to enable RDMA in containers by presenting the needed IB devices from the host.

## Inifiniband Devices
The known needed devices for RDMA are:
* /dev/infiniband/rdma_cm
* /dev/infiniband/uverbs*

Optionally present the [KNEM](http://knem.gforge.inria.fr/) device for Open/IBM MPI if available:
* /dev/infiniband/knem

## Contents
* Device plugin code
* DaemonSet YAML file

### Notes
Developed with Kubernetes 1.12 but using sematic versioning, not tied to a specific release, 1.10+

## TODO
*
