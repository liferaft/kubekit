# Image with all the Go modules required to build KubeKit* applications
FROM golang:latest AS base

ARG     GIT_COMMIT=somehash
ARG     BUILD_ID=0

# Dependencies
# tar cannot unzip the .xz file because xz is not installed. The installation of
# xz fail in this base image
# ENV     UPX_VER 3.94
# ADD     https://github.com/upx/upx/releases/download/v${UPX_VER}/upx-${UPX_VER}-amd64_linux.tar.xz /
# RUN     tar -xf /upx-${UPX_VER}-amd64_linux.tar.xz && \
#         mv /upx-${UPX_VER}-amd64_linux/upx /bin/upx && \
#         rm -f /upx-${UPX_VER}-amd64_linux.tar.xz

# These don't work on Jenkins due to the default network config. To make them
# work Docker requires the following network & DNS configuration:
# --net=host --dns 153.64.180.100 --dns 153.64.251.200 --dns-opt attempts:5 --dns-opt timeout:15
# RUN     go get golang.org/x/lint/golint && \
#         go get github.com/axw/gocov/gocov && \
#         go get github.com/AlekSi/gocov-xml

WORKDIR /workspace/liferaft/kubekit

# Get internal packages
COPY    ./staging/src/github.com ./staging/src/github.com
ENV     GO111MODULE=on
COPY    go.mod .
COPY    go.sum .
RUN     go mod download

# Builder image, to build KubeKit and KubeKitCtl
FROM base AS builder

COPY    . .

RUN     CGO_ENABLED=0 \
        GOOS=linux \
        GOARCH=amd64 \
        go build -o /kubekitctl \
          -ldflags="-X github.com/liferaft/kubekit/version.GitCommit=${GIT_COMMIT} -X github.com/liferaft/kubekit/version.Build=${BUILD_ID} -s -w " \
          ./cmd/kubekitctl/main.go

RUN     CGO_ENABLED=0 \
        GOOS=linux \
        GOARCH=amd64 \
        go build -o /kubekit \
          -ldflags="-X github.com/liferaft/kubekit/version.GitCommit=${GIT_COMMIT} -X github.com/liferaft/kubekit/version.Build=${BUILD_ID} -s -w " \
          ./cmd/kubekit/main.go

# KubeKit application image for development
FROM alpine:3.9 AS kubekit-dev

COPY --from=builder /kubekit /app/

ENTRYPOINT [ "ash" ]

# KubeKitCtl application image for development
FROM alpine:3.9 AS kubekitctl-dev

COPY --from=builder /kubekitctl /app/

ENTRYPOINT [ "ash" ]

# KubeKit application image
FROM alpine:3.9 AS kubekit

COPY --from=builder /kubekit /app/

ENTRYPOINT [ "/app/kubekit" ]

# KubeKitCtl application image
FROM alpine:3.9 AS kubekitctl

COPY --from=builder /kubekitctl /app/

ENTRYPOINT [ "/app/kubekitctl" ]