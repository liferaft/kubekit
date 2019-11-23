SHELL  				:= /bin/bash

# Output:
NO_COLOR 		 ?= false
ifeq ($(NO_COLOR),false)
ECHO 				:= echo -e
C_STD 				= $(shell $(ECHO) -e "\033[0m")
C_RED		 		= $(shell $(ECHO) -e "\033[91m")
C_GREEN 			= $(shell $(ECHO) -e "\033[92m")
C_YELLOW 			= $(shell $(ECHO) -e "\033[93m")
C_BLUE	 			= $(shell $(ECHO) -e "\033[94m")
I_CROSS 			= $(shell $(ECHO) -e "\xe2\x95\xb3")
I_CHECK 			= $(shell $(ECHO) -e "\xe2\x9c\x94")
I_BULLET 			= $(shell $(ECHO) -e "\xe2\x80\xa2")
else
ECHO 				:= echo
C_STD 				=
C_RED		 		=
C_GREEN 			=
C_YELLOW 			=
C_BLUE	 			=
I_CROSS 			= x
I_CHECK 			= .
I_BULLET 			= *
endif


################################## Main Rules #################################

# rule that is executed when no rule is specified in make
.PHONY: default
default: generate

############################### Unit Test Rules ###############################

# initialize the templates directory for unit test
.PHONY: unit-test-init
unit-test-init: 
	@for p in */templates; do \
		$(ECHO) "$(C_GREEN) Initializing Unit Testing environment for $(C_YELLOW)$${p%%/*}$(C_STD)"; \
		cd $${p}; \
		terraform init 2>&1 >/dev/null; \
		if [[ -f terraform.tfvars ]]; then \
			mv terraform.tfvars terraform.tfvars.backup; \
			$(ECHO) "    $(C_YELLOW) You have a backup of your variables on $(C_RED)terraform.tfvars.backup$(C_STD)"; \
		fi; \
		cd ../../; \
	done
	@go run test/main.go
	@$(ECHO) "$(C_YELLOW) Edit the Terraform variables files located in $(C_RED)*/templates/terraform.tfvars$(C_STD)";

# P is the Unit Testing Platform. Could be 'aws' or 'vsphere'
P	?= vsphere

# executes the unit test for an specific platform
.PHONY: unit-test
unit-test: 
	@$(ECHO) "$(C_GREEN) Executing Unit Testing for $(C_YELLOW)$(P)$(C_STD)"
	@cd $(P)/templates; terraform apply --auto-approve

# executes the unit test for all the platform
.PHONY: unit-test-all
unit-test-all: 
	@for p in */templates; do $(MAKE) unit-test P=$${p%%/*}; done

# executes the unit test for an specific platform
.PHONY: unit-test
unit-test-destroy: 
	@$(ECHO) "$(C_GREEN) Destroy the cluster used for Unit Testing on $(C_YELLOW)$(P)$(C_STD)"
	@cd $(P)/templates; terraform destroy -force 

# executes the unit test for all the platform
.PHONY: unit-test-destroy-all
unit-test-destroy-all:
	@for p in */templates; do $(MAKE) unit-test P=$${p%%/*}; done

.PHONY: go-unit-tests
go-unit-tests:
	# Run all non vendor unit tests
	go test -v ./...

################################# Build Rules #################################


############################# Code Generation ################################
# generates the Go code with the Terraform code located in
# <platform>/templates directory. The Go code will be in the file
# <platform>/terraform.go. NOTE: not all platforms have generated code
# (e.g. raw and stacki do not have terraform code)
.PHONY: generate
generate: aks/code.go aws/code.go openstack/code.go vsphere/code.go aks/code.go eks/code.go

# Rule will regenerate a terraform.go file if any of the terraform files
# in the templates subdirectory have been modified.
%/code.go: %/templates/*.tf
	GO111MODULES=off go run codegen/main.go --pkg $(@D) --src $(dir $<) --dst $@

# Go integration tests: run Go code to provision actual clusters
.PHONY: test-provisioner-integration
test-provisioner-integration:
	@go test . -v -run TestProvisionerIntegration -tags integration

# test-dry-run will ensure that all test files compile without error
.PHONY: test-dry-run
test-dry-run:
	@for f in `go list ./... | grep -v vendor`; do \
		echo "Compiling tests in package $$f"; \
		TEST_BIN=`mktemp`; \
		go test -c -o $${TEST_BIN} -tags integration $$f 2>&1; \
		rm $${TEST_BIN}; \
	done;

########################## Test reports for Jenkins ###########################
# The following section contains tools, reports, and processes for creating
# Jenkins reports. Jenkins expects reports to be in a format that the plugins
# understand, which is typically an XML report (e.g. junit and cobertura). Go
# does not produce these reports natively, so some third party conversion tools
# are needed to create these reports.

# # All test reports and artifacts are stored in build/reports. The build folder
# # is ignored by git.
# build/reports:
# 	mkdir -p build/reports

# # CLI tools for converting Go test reports

# $(GOPATH)/bin/gocov:
# 	GO111MODULES=off go get github.com/axw/gocov/...

# $(GOPATH)/bin/gocov-xml:
# 	GO111MODULES=off go get github.com/AlekSi/gocov-xml

# # go-junit-report converts Go test reports into junit XML reports
# $(GOPATH)/bin/go-junit-report:
# 	go get -u github.com/jstemmer/go-junit-report

# # Integration test will produce two artifacts:
# # - build/reports/integration-tests.gotest: output of the go test
# # - build/reports/integration-tests.coverprofile: a test coverage profile program
# build/reports/integration-tests.gotest build/reports/integration-tests.coverprofile: build/reports
# 	# TODO: Running go test with the verbose (-v) option might cause issues with building the junit XML files.
# 	# If issues start to pop up, remove the -v option below.
# 	go test -v -coverprofile=build/reports/integration-tests.coverprofile -run TestProvisionerIntegration -tags integration > build/reports/integration-tests.gotest

# # Unit tests will produce two artifacts:
# # - build/reports/unit-tests.gotest: output of the go test
# # - build/reports/unit-tests.coverprofile: a test coverage profile program
# build/reports/unit-tests.gotest build/reports/unit-tests.coverprofile:  build/reports
# 	# Run all non vendor unit tests and store report as a jenkins compatible junit report
# 	# TODO: Running go test with the verbose (-v) option might cause issues with building the junit XML files.
# 	# If issues start to pop up, remove the -v option below.
# 	go test -v -coverprofile=build/reports/unit-tests.coverprofile ./... > build/reports/unit-tests.gotest

# # Example targets: build/reports/go_unit_tests.xml and build/reports/go_integration_tests.xml
# build/reports/%_tests.junit.xml: build/reports/%-tests.gotest $(GOPATH)/bin/go-junit-report
# 	cat build/reports/$*-tests.gotest | go-junit-report > build/reports/$*_tests.junit.xml

# # Example targets: build/reports/unit-tests.cobertura.xml and build/reports/integration-tests.cobertura.xml
# build/reports/%-tests.cobertura.xml: build/reports/%-tests.coverprofile $(GOPATH)/bin/gocov $(GOPATH)/bin/gocov-xml
# 	gocov convert build/reports/$*-tests.coverprofile | $(GOPATH)/bin/gocov-xml > build/reports/$*-tests.cobertura.xml

# # Clean is helpful for getting back to a clean state when troubleshooting issues locally
# .PHONY: clean
# clean:
# 	rm -Rf build vendor
# 	# Cleaning Go caches requires admin rights:
# 	sudo go clean -cache -testcache -modcache -r -i
