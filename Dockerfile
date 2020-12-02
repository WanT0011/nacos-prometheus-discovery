FROM golang:1.14.1-alpine as builder

ENV GOPROXY=https://goproxy.cn,https://goproxy.io,direct \
    GO111MODULE=on \
    CGO_ENABLED=1

WORKDIR /go/src/nacos-prometheus-discovery
RUN go env -w GOPROXY=https://goproxy.cn,https://goproxy.io,direct
COPY . /go/src/nacos-prometheus-discovery
RUN go build .

FROM lo-harbor.yyjzt.com/shengyi/busybox:v1
LABEL MAINTAINER="shengyi"

WORKDIR /app
# copy go apps
COPY --from=builder /go/src/nacos-prometheus-discovery/nacos-prometheus-discovery .
COPY --from=builder /go/src/nacos-prometheus-discovery/conf/config.json .

EXPOSE 8080
ENTRYPOINT ["/app/nacos-prometheus-discovery","/app/config.json"]
