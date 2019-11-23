SHELL  				:= /bin/bash

################################# Build Rules #################################

.PHONY: install 
install: fmt test
	go install .

.PHONY: test
test:
	go test -v -cover ./...

.PHONY: fmt
fmt:
	go fmt ./...
	go vet ./...
	go list ./... | xargs -n1 golint

