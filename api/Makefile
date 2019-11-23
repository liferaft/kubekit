.PHONY: default all build dependencies clean clean-code

PROTOC_VERSION 	= 3.10.1
PROTOC_INC 			= -I/usr/local/include -I. \
									-I$(GOPATH)/src \
									-I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
									-I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway

MY_GOARCH 	= $(shell go env GOARCH | sed -e 's/amd64/x86_64/' -e 's/386/x86_32/')
MY_GOOS 		= $(shell go env GOOS | sed -e 's/darwin/osx/')

VERSIONS		= kubekit/v1

default: clean-code build

all: dependencies default clean

build:
	@for d in $(VERSIONS); do echo "Building proto for version $$d ..."; $(MAKE) -C $$d PROTOC_INC="$(PROTOC_INC)"; done

clean-code:
	@for d in $(VERSIONS); do echo "Cleaning code for version $$d ..."; $(MAKE) -C $$d clean; done

dependencies:
	mkdir -p /tmp/protoc && \
	curl -sLk https://github.com/google/protobuf/releases/download/v$(PROTOC_VERSION)/protoc-$(PROTOC_VERSION)-$(MY_GOOS)-$(MY_GOARCH).zip | \
		tar -xzv -C /tmp/protoc
	mv /tmp/protoc/bin/protoc $(GOPATH)/bin
	chmod +x $(GOPATH)/bin/protoc
	rm -rf /usr/local/include/google
	mv /tmp/protoc/include/google /usr/local/include/
	GO111MODULE=off go get -u github.com/grpc-ecosystem/grpc-gateway/{protoc-gen-grpc-gateway,protoc-gen-swagger}
	GO111MODULE=off go get -u google.golang.org/grpc
	GO111MODULE=off go get -u github.com/golang/protobuf/{proto,protoc-gen-go}

clean:
	$(RM) -rf /tmp/protoc
