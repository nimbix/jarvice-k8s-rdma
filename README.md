# Infiniband RDMA device plugin for Kubernetes on JARVICE

A **Kubernetes** [device plugin](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/) 
with a DaemonSet to detect and pass _Infiniband_ (IB) devices to requesting pods from the kubelet, 
used to enable RDMA in containers by presenting the needed IB devices from the host.

Based on work from the [Nimbix fork of k8s-rdma-device-plugin](https://github.com/nimbix/k8s-rdma-device-plugin) 
and the [NVIDIA reference device plugin](https://github.com/NVIDIA/k8s-device-plugin), this device plugin removes the use
of the ibverbs library which appears to have issues with mixed software/firmware revisions.

## Infiniband Devices
The known required devices for RDMA are:
* /dev/infiniband/uverbs*
  
Optionally:
* /dev/infiniband/rdma_cm

Optionally, if present, the [KNEM](http://knem.gforge.inria.fr/) device for Open/IBM MPI:
* /dev/knem

## Contents
* Device plugin code
* DaemonSet YAML file for plugin deployment
* 2-pod example YAML file to test RDMA between two pods
* Dockerfile and build scripts

## Kubernetes Deployment
Deploying the device plugin to each node requires applying the DaemonSet to the chosen namespace:
```
$ kubectl -n jarvice-daemonsets apply -f rdma-device.yml
```

### Notes
Developed with Kubernetes 1.12 and Go modules, using semantic versioning, not tied to a specific release, 1.10+

Build command (local build):
```bash
go build -o jarvice-rdma-plugin .
```

Build command for a static Linux binary:
```bash
CGO_ENABLED=0 GOOS=linux go build -o jarvice-rdma-plugin -a -ldflags '-extldflags "-static"' .
```

Build Docker image locally:
```bash
docker build -t jarvice/k8s-rdma-device:1.0.0 .
```

Run (privileged) Docker image locally:
(create mocked */var/lib/kubelet/device-plugins* and */tmp/infiniband* directories, touch a /tmp/infiniband/uverbs0 file)
```bash
docker run --security-opt=no-new-privileges --cap-drop=ALL --network=none -it -v /var/lib/kubelet/device-plugins:/var/lib/kubelet/device-plugins -v /tmp/infiniband:/dev/infiniband  jarvice/k8s-rdma-device:1.0.0
```

Run unit tests for RDMA source:
```bash
 go test -v ./rdma/...
```

Run unit tests for the sysutl source:
```bash
 go test -v ./sysutl/...
```

The plugin is built for x86_64 and ppc64le architectures with Docker manifests.

## Testing

#### Example pods, test RDMA bandwidth using *ib_read_bw*
This setup will run two Xenial pods, each allocating the RDMA devices for the node, 
then run the Infiniband diagnostic tools to test RDMA between the nodes/pods.

* Find the IP for the first pod:
  * `kubectl get pods -o wide`
    * `ibpod  1/1 Running  0 13m 10.40.0.27`

* Get a shell in the first pod:
  * `kubectl exec -ti ibpod -c ibpod-ctr bash`

* Add some diagnostic tools:
  * `apt-get -y update && 
apt-get -y install infiniband-diags libibverbs1 libmlx4-1 libmlx5-1 ibutils ibverbs-utils perftest && 
apt-get clean`

* Run the server command on the first pod:
  * `ib_read_bw -d mlx4_0 -i 1 -F --report_gbits`

* Get a shell in the second pod:
  * `kubectl exec -ti ibpod2 -c ibpod-ctr2 bash`

* Add the diagnostic tools to the second pod:
  * `apt-get -y update && 
apt-get -y install infiniband-diags libibverbs1 libmlx4-1 libmlx5-1 ibutils ibverbs-utils perftest && 
apt-get clean`

* run the client command:
  * `ib_read_bw -d mlx4_0 -i 1 -F --report_gbits 10.40.0.27`
  
* output should be similar to:
  *      ---------------------------------------------------------------------------------------
          #bytes     #iterations    BW peak[Gb/sec]    BW average[Gb/sec]   MsgRate[Mpps]
          65536      1000             49.72              49.72  		   0.094832
         ---------------------------------------------------------------------------------------
     
## TODO
* health checks, seemingly mostly for hot plug devices
* update the dial function for newer API
