
# build the application without testing or any other validation.
# The binary is located in $(PKG)/
PLATFORMS       ?=darwin/amd64 linux/amd64 linux/386 windows/amd64

.PHONY: release-package
release-package: copy_packages
	@rm -f upx.save
	@rm -fr $(TMPBIN)
	@mkdir $(TMPBIN)
	@mv $(PKG)/$(BIN_NAME)_$(VERSION)* $(TMPBIN)/
	for PLAT in $(PLATFORMS) ; do \
		OS_ARCH=$$(echo $$PLAT|tr / _); \
		printf "$(UPXCMD) $(TMPBIN)/$(BIN_NAME)_$(VERSION)_$$OS_ARCH -o $(PKG)/$(BIN_NAME)_$(VERSION)_$$OS_ARCH\n" >> upx.save; \
	done

	cat upx.save | tr '\n' '\0' | xargs -t -n1 -P8 -0 bash -c
	@rm -fr $(TMPBIN)

	for PLAT in $(PLATFORMS) ; do \
		OS_ARCH=$$(echo $$PLAT|tr / _); \
		chmod +x $(PKG)/$(BIN_NAME)_$(VERSION)_$$OS_ARCH; \
		tar -C $(PKG) -cvzf $(BIN_NAME)_$(VERSION)_$$OS_ARCH.tgz $(BIN_NAME)_$(VERSION)_$$OS_ARCH $(TGZ_PACKAGES);\
	done

# build the application for every OS and Architecture. The binaries are located
# in $(PKG)/
.PHONY: build-all
build-all: $(PLATFORMS)
	@$(ECHO) "$(C_GREEN) All binaries $@ are located at $(C_YELLOW)$(PKG)/$(C_STD)"

$(PLATFORMS):
	GOOS=$(word 1, $(subst /, ,$@)) GOARCH=$(word 2, $(subst /, ,$@)) go build $(LDFLAGS) -o $(PKG)/$(BIN_NAME)_$(VERSION)_$(word 1, $(subst /, ,$@))_$(word 2, $(subst /, ,$@)) ./cmd/kubekit/main.go

.PHONY: copy_packages
copy_packages:
	$(shell mkdir -p $(PKG))
	@cp docs/cli-ux.md ./USER_GUIDE.md
	@cp $(TGZ_PACKAGES) $(PKG)

