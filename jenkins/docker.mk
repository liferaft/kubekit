# all-in-docker will do the same as 'all' but using a container to build instead
# of the Go at your system.
.PHONY: all-in-docker
all-in-docker: build-in-docker

#Without specifying GOOS=darwin GOARCH=amd64 it will build for linux
.PHONY: build-in-docker
build-in-docker:
	$(RM) -r $(PKG)/$(BIN_NAME)*
	docker run --rm \
		-v $(CURDIR):/go/src/$(PKG_BASE)/$(PKG_NAME) \
		-w /go/src/$(PKG_BASE)/$(PKG_NAME) \
		golang:$(GOLANG_VER) make build GOOS=darwin GOARCH=amd64

.PHONY: build-all-in-docker
build-all-in-docker:
	$(RM) -r $(PKG)/$(BIN_NAME)_$(VERSION)*
	docker run --rm \
	  -v $(CURDIR):/go/src/$(PKG_BASE)/$(PKG_NAME) \
		-w /go/src/$(PKG_BASE)/$(PKG_NAME) \
		golang:$(GOLANG_VER) make build-all

# Build KubeKit or KubeKitCtl for production or development using the following 
# rule names: docker-build-{kubekit, kubekitctl, kubekit-dev, kubekitctl-dev}
# .PHONY: docker-build-kubekit docker-build-kubekitctl docker-build-kubekit-dev docker-build-kubekitctl-dev
docker-build-%:
	@name=$@; app=$${name##*-}; \
	$(ECHO) "$(C_GREEN)Building Docker image $(C_YELLOW)$${app%%-*}$(C_GREEN) with/for $(C_YELLOW)KubeKit v$(VERSION) ($(GIT_COMMIT)-$(BUILD_ID))$(C_STD)"; \
	docker build \
		--tag $${app} \
		--target=$${app} \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_ID=$(BUILD_ID) \
		.