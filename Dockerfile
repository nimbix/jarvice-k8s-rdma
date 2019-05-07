FROM golang:1.11-stretch as build

RUN apt-get -y update && apt-get -y install libibverbs-dev

WORKDIR /go/src/github.com/nimbix/k8s-rdma-device-plugin
COPY . .

#RUN go get -d -v ./...
RUN go install -v ./...

FROM ubuntu:xenial

COPY --from=build /go/bin/k8s-rdma-device-plugin /usr/local/bin
RUN apt-get -y update && apt-get -y install libibverbs1 libmlx4-1 libmlx5-1 ibutils ibverbs-utils perftest && apt-get clean

ENTRYPOINT ["k8s-rdma-device-plugin"]