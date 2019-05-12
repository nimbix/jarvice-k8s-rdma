FROM golang:1.12 as build

WORKDIR /go/jarvice-k8s-rdma-device-plugin
COPY . .

RUN go build -o jarvice-ibrdma-plugin .

FROM ubuntu:bionic

COPY --from=build /go/jarvice-k8s-rdma-device-plugin /usr/local/bin
RUN apt-get -y update && \
    apt-get -y install libibverbs1 libmlx4-1 libmlx5-1 ibutils ibverbs-utils perftest && \
    apt-get clean

ENTRYPOINT ["jarvice-ibrdma-plugin"]