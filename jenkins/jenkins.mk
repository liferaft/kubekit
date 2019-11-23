# build the application without testing or any other validation.
# The binary is located in $(PKG)/
.PHONY: jenkins-build
jenkins-build: copy_packages
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(PKG)/$(BIN_NAME)_linux_amd64 ./cmd/kubekit/main.go
	@$(ECHO) "$(C_GREEN)Built $(BIN_NAME)_linux_amd64 $(C_YELLOW)v$(VERSION) ($(GIT_COMMIT))$(C_STD)"

TGZ_PACKAGES=README.md KNOWN_ISSUES.md CONTRIBUTING.md USER_GUIDE.md *_CHANGELOG.md kubekit*.rpm

.PHONY: jenkins-package
jenkins-package:
	$(UPXCMD) $(PKG)/$(BIN_NAME)_linux_amd64
	chmod +x $(PKG)/$(BIN_NAME)_linux_amd64
	tar -C $(PKG) -cvzf $(BIN_NAME)_linux_amd64.tgz $(BIN_NAME)_linux_amd64 $(TGZ_PACKAGES)
