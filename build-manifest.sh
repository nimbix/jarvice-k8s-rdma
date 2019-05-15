#!/usr/bin/env bash

JARVICE_RDMA_VERSION=1.0.0
[[ -n "$1" ]] && JARVICE_RDMA_VERSION="$1"

docker manifest create \
    docker.io/nimbix/lxcfs:${JARVICE_RDMA_VERSION} \
    docker.io/nimbix/lxcfs:${JARVICE_RDMA_VERSION}-amd64 \
    docker.io/nimbix/lxcfs:${JARVICE_RDMA_VERSION}-ppc64le

docker manifest push docker.io/nimbix/lxcfs:${JARVICE_RDMA_VERSION}
