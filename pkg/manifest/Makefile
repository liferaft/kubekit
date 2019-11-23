SHELL  			:= /bin/bash

VERSION 		= $(shell grep 'const Version ' version.go | cut -f2 -d= | tr -d ' ' | tr -d '"')
VER_SION 		= $(shell echo $(VERSION) | tr '.' '_')
GOLANG_VER 		= 1.12.7

LASTEST_RELEASE 	?= release$(VER_SION)

NR			?=
NEW_RELEASE 		= $(shell echo $(NR) | tr '.' '_')

.PHONY: release
release:
	@if [[ -z '$(NEW_RELEASE)' ]]; then echo "ERROR: Missing the new release number, the variable NR is required. Example: make NR=1.1.2"; exit 1; fi
	@sed 's/$(VER_SION)/$(NEW_RELEASE)/g' $(LASTEST_RELEASE).go > release$(NEW_RELEASE).go
	@sed -i.bkp -e '/Releases: map\[string\]Release{/a\'$$'\n''		"$(NR)": release$(NEW_RELEASE),' manifest.go
	@sed -i.bkp 's/const Version = ".*"/const Version = "$(NR)"/' version.go
	@$(RM) *.bkp
	@[[ -z $$(command -v go) ]] || go fmt .

.PHONY: rollback-release
rollback-release:
	@rollbackVer=$$(grep ': release' manifest.go| head -1 | cut -f1 -d: | tr -d '"' | tr -d "\t"); \
	rollback_ver=$$(echo $${rollbackVer} | tr '.' '_'); \
	$(RM) release$${rollback_ver}.go; \
	sed -i.bkp '/		"'$${rollbackVer}'": 'release$${rollback_ver}',/d' manifest.go
	@prev_ver=$$(grep ': release' manifest.go| head -1 | cut -f1 -d: | tr -d '"' | tr -d "\t"); \
	sed -i.bkp 's/const Version = ".*"/const Version = "'$${prev_ver}'"/' version.go
	@$(RM) *.bkp
	@[[ -z $$(command -v go) ]] || go fmt .

$(GOPATH)/bin/gocov:
	GO111MODULES=off go get github.com/axw/gocov/...

$(GOPATH)/bin/gocov-xml:
	GO111MODULES=off go get github.com/AlekSi/gocov-xml

# execute unit testing to the Go code
.PHONY: test
test: $(GOPATH)/bin/gocov $(GOPATH)/bin/gocov-xml
	$${GOPATH}/bin/gocov test ./... | $${GOPATH}/bin/gocov-xml > cobertura.xml

.PHONY: test-in-docker
test-in-docker:
	docker run --rm \
	        --net=host --dns 153.64.180.100 --dns 153.64.251.200 --dns-opt attempts:5 --dns-opt timeout:15 \
		-v $(CURDIR):/go/src/github.com/kubekit/manifest \
		-w /go/src/github.com/kubekit/manifest \
		golang:$(GOLANG_VER) make vendor-mod test

.PHONY: vendor
vendor:
	go get gopkg.in/yaml.v2

# vendor-mod vendors all dependencies based on go.mod file
.PHONY: vendor-mod
vendor-mod: go.mod
	GO111MODULE=on go mod vendor

# install will compile and install all packages/executables while
# skipping any unused vendor packages
.PHONY: install
install: vendor
	go install `go list ./... | grep -v vendor`

# configure git credentials for use in CI
.PHONY: gitconfig
gitconfig:
	git config --global url."https://$(GIT_USERNAME):$(GIT_PASSWORD)@github.com".insteadOf "https://github.com"

.PHONY: all
all: fmt install test

# verifies the format of the Go code, analyze the Go code to find suspiciou
# constructs (common mistakes) and checks the Go code for actual programming
# errors and style violations.
.PHONY: fmt
fmt:
	go fmt ./...
	go vet ./...
	go list ./... | xargs -n1 golint

# generates the Go code with the Terraform code located in <platform>/templates
# directory. The Go code will be in the file <platform>/terraform.go
.PHONY: generate
generate:
	go generate .
