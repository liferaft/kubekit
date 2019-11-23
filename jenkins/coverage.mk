.PHONY: coverage
coverage:
	@GO111MODULE=off go get github.com/axw/gocov/gocov github.com/AlekSi/gocov-xml
	@$(ECHO) "$(C_GREEN)Testing & Coverage$(C_STD)"
	@$(GOPATH)/bin/gocov test $$(go list ./...) | $(GOPATH)/bin/gocov-xml > cobertura.xml
