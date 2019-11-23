SHELL  				:= /bin/bash

# rule that is executed when no rule is specified in make
.PHONY: default
default: generate

# generates the Go code with the contents of the templates directory
.PHONY: generate
generate:
	GO111MODULES=off go generate
	GO111MODULES=off cd resources && go generate

# generates the Go code with the contents of the templates directory and verifies the sha
.PHONY: jenkins-generate
jenkins-generate: generate
ifeq ($(shell git status|grep -c code.go), 1)
	@echo "FAILURE: code.go does not match"
	shasum code.go || sha1sum code.go
	git ls-files -s code.go
	exit 1
endif
