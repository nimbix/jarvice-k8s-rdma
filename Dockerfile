FROM golang:1.13 as build

WORKDIR /go/jarvice-k8s-rdma-device-plugin
COPY . .

# statically compile the plugin
RUN CGO_ENABLED=0 GOOS=linux go build -o jarvice-rdma-plugin -a -ldflags '-extldflags "-static"' .

FROM scratch

COPY --from=build /go/jarvice-k8s-rdma-device-plugin /usr/local/bin

ENTRYPOINT ["jarvice-rdma-plugin"]