#!/usr/bin/env bash

JARVICE_RDMA_VERSION=1.0.0
[[ -n "$1" ]] && JARVICE_RDMA_VERSION="$1"

ARCH=$(uname -m)
[[ "$ARCH" = "x86_64" ]] && ARCH=amd64

echo "Building version: ${JARVICE_RDMA_VERSION} for arch: ${ARCH}"
docker build --rm -t jarvice/k8s-rdma-device:${JARVICE_RDMA_VERSION}-${ARCH} .

docker push jarvice/k8s-rdma-device:${JARVICE_RDMA_VERSION}-${ARCH}
