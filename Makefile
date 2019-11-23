#===============================================================================
# Author: Johandry Amador <johandry@gmail.com>
# Title:  Makefile to automate the builds, tests and deployments.
#
# Usage: make [<rule>]
#
# Main rules:
#  <none>      If no rule is specified will do the 'default' rule.
#  build       Build just the binary.
#  clean       Remove all the created binaries, containers and images.
#  helpDisplay all the existing rules and description of what they do.
#  version     Shows the application version.
#  all         Execute the end to end default process.
#
# Description: This Makefile is to automate every action used for development,
# testing or release of the KubeKit CLI.
# Use 'make help' to view all the options or go to
# https://github.com/liferaft/kubekit
#
# Report Issues or create Pull Requests in https://github.com/liferaft/kubekit
#===============================================================================

## Variables:
## -----------------------------------------------------------------------------

# SHELL need to be defined at the top of the Makefile. Do not change its value.
SHELL			:= /bin/bash

# all the defined variables in the Makefile are exported as environment variables
# to be used by every rule
.EXPORT_ALL_VARIABLES:

BIN       = bin
PKG       = binaries
TMPBIN    = tmp-bin
BIN_NAME  = kubekit
CTL_NAME  = kubekitctl
FROM      ?= alpine

# Macros to set the application version, needed for the build:
GIT_COMMIT		=  $(shell git rev-parse --short HEAD  2>/dev/null || echo 'unknown')
BUILD_ID			?= 0
VERSION				=  $(shell grep 'const Version ' ./pkg/manifest/version.go | cut -f2 -d= | tr -d ' ' | tr -d '"')
PKG_NAME			:= kubekit
PKG_BASE			:= github.com/kubekit
TAR_COMMENT		?= ''
BINARY				=  $(BIN)/$(BIN_NAME)
LDFLAGS				=  -ldflags '\
	-X $(PKG_BASE)/$(PKG_NAME)/version.GitCommit=$(GIT_COMMIT) \
	-X $(PKG_BASE)/$(PKG_NAME)/version.Build=$(BUILD_ID) -s -w'


GOLANG_VER		?= 1.13.1
GO111MODULE		:= on
GOMOD					?= on

ifeq ($(GOMOD),on)
	GOVENDOR 			=
	GOMODVAR			= GO111MODULE=on
else
	GOVENDOR 			= -mod=vendor
	GOMODVAR			= GO111MODULE=off
endif

ifeq (,$(shell which upx))
#failed
	UPXCMD=ls
else
	UPXCMD=upx
endif

include jenkins/*.mk
## To find rules without .PHONY:
# diff <(grep '^.PHONY:' Makefile | sed 's/.PHONY: //' | tr ' ' '\n' | sort) <(grep '^[^# ]*:' Makefile | grep -v '.PHONY:' | sed 's/:.*//' | sort) | grep '[>|<]'

## Default Rules:
## -----------------------------------------------------------------------------

# default is the rule that is executed when no rule is specified in make
.PHONY: default
default: fmt mod generate test build

# all is to execute the entire process to build the KubeKit binaries for every
# OS and Architecture, and an image to ship it.
.PHONY: all
all: build-all

# help to print all the commands and description
.PHONY: help
help:
	@content=""; grep -v '.PHONY:' Makefile | grep -v '^##' | grep -v '^\s' | grep -v '^$$' | grep '^[^# ]*:' -B 5 | grep -E '^#|^[^# ]*:' | \
	while read line; do if [[ $${line:0:2} == "# " ]]; \
		then l=$$($(ECHO) $$line | sed 's/^# /  /'); content="$${content}\n$$l"; \
		else header=$$($(ECHO) $$line | sed 's/^\([^ ]*\):.*/\1/'); [[ $${content} == "" ]] && content="\n  $(C_YELLOW)No help information for $${header}$(C_STD)"; $(ECHO) "$(C_BLUE)$${header}:$(C_STD)$$content\n"; content=""; fi; \
	done

# display the version of this project
.PHONY: version
version:
	@[[ -z "$(SHORT)" ]] || $(ECHO) "$(VERSION)"
	@[[ -z "$(LONG)" ]]  || $(ECHO) "$(VERSION)+build.$(BUILD_ID).$(GIT_COMMIT)"
	@[[ -n "$(SHORT)" || -n "$(LONG)" ]] || $(ECHO) "$(C_GREEN)Version:$(C_STD) v$(VERSION)+build.$(BUILD_ID).$(GIT_COMMIT)$(C_STD)"

## Build Rules:
## -----------------------------------------------------------------------------


# tests the Go code
.PHONY: test
test:
	@$(ECHO) "$(C_GREEN)Testing$(C_STD)"
	$(GOMODVAR) go test $(GOVENDOR) ./...

# checks the Go code for actual programming errors and style violations.
.PHONY: fmt
fmt:
	@$(ECHO) "$(C_GREEN)Checking fmt$(C_STD)"
	@files=$$(GO111MODULE=off go fmt ./...); if [[ -n $${files} ]]; then $(ECHO) "$(C_RED)$(I_CROSS) Go fmt found errors but the code has been fixed.\n$(C_GREEN)Commit and push the following files:$(C_YELLOW)\n$${files}\n$(C_RED)Next time, execute $(C_YELLOW)make fmt$(C_RED) before commit$(C_STD)"; exit 1; fi
	@$(ECHO) "$(C_GREEN)Checking lint$(C_STD)"
	@[[ -x $(GOPATH)/bin/golint ]] || GO111MODULE=off go get -u golang.org/x/lint/golint
	@pkgs=$$($(GOMODVAR) go list ./...); \
	$(GOMODVAR) $(GOPATH)/bin/golint $${pkgs}; \
	$(ECHO) "$(C_GREEN)Checking vet$(C_STD)"; \
	$(GOMODVAR) go vet $${pkgs}

# build the application without testing or any other validation.
# The binary is located in $(BIN)/$(BIN_NAME)
.PHONY: build
build:
	@$(ECHO) "$(C_GREEN)Building $(C_YELLOW)v$(VERSION) ($(GIT_COMMIT))$(C_GREEN) for $(C_YELLOW)$$(go env GOOS)$(C_STD)"
	@mkdir -p $(BIN)
	@go build $(GOVENDOR) $(LDFLAGS) -o $(BIN)/$(BIN_NAME) ./cmd/kubekit/main.go && \
		$(ECHO) "$(C_GREEN)$(I_CHECK) Build completed at $(C_YELLOW)$(BINARY)$(C_STD)" || \
		$(ECHO) "$(C_RED)$(I_CROSS) Build failed$(C_STD)"

# build the KubeKit-Server client without testing or any other validation.
# The binary is located in $(BIN)/$(CTL_NAME)
.PHONY: build-ctl
build-ctl:
	@$(ECHO) "$(C_GREEN)Building $(CTL_NAME) $(C_YELLOW)v$(VERSION) ($(GIT_COMMIT))$(C_GREEN) for $(C_YELLOW)$$(go env GOOS)$(C_STD)"
	@mkdir -p $(BIN)
	@go build $(GOVENDOR) -o $(BIN)/$(CTL_NAME) ./cmd/kubekitctl/ && \
		$(ECHO) "$(C_GREEN)$(I_CHECK) Build completed at $(C_YELLOW)$(BIN)/$(CTL_NAME)$(C_STD)" || \
		$(ECHO) "$(C_RED)$(I_CROSS) Build failed$(C_STD)"

# all the actions required to have the Go modules in your environment
.PHONY: mod
mod:
	@$(ECHO) "$(C_GREEN)Downloading Go modules$(C_STD)"
	@go mod tidy
	@go mod download

# generates the Go code with the contents of the templates directory
.PHONY: generate
generate:
	$(MAKE) -C pkg/configurator
	$(MAKE) -C pkg/provisioner

# api generates the Go code for the gRPC and REST API from the .proto files,
# generates the swagger files and API documentation
.PHONY: api
api:
	$(MAKE) -C api


API_URL		?= https://localhost:5823/swagger/kubekit.json
# start a swagger server to render the swagger API documentation
.PHONY: show-swagger
show-swagger:
	@$(ECHO) "$(C_GREEN)open a browser to: $(C_YELLOW)http://localhost$(C_STD)"
	docker run -p 80:8080 -e API_URL=$(API_URL) swaggerapi/swagger-ui

# view the GoDoc in a local web server
.PHONY: godoc
godoc:
	@$(ECHO) "$(C_GREEN)Open a browser on $(C_YELLOW)http://localhost:6060/pkg/github.com/liferaft/kubekit/$(C_STD)"
	docker run --rm \
		-v $(CURDIR):/go/src/$(PKG_BASE)/$(PKG_NAME) \
		--expose=6060 \
		-p 6060:6060 \
		golang:$(GOLANG_VER) \
			godoc -http=":6060" -play

# remove the built binaries
.PHONY: clean
clean:
	@$(ECHO) "$(C_GREEN)Cleaning binaries$(C_STD)"
	@$(RM) -r $(PKG)
	@$(RM) -r $(BIN)

# clean all the Go caches
.PHONY: clean-go-env
clean-go-env:
	@$(ECHO) "$(C_GREEN)Cleaning Go environment$(C_STD)"
	@go clean -modcache -testcache -cache -i -r

# clean up everything
.PHONY: clean-all
clean-all: clean clean-go-env