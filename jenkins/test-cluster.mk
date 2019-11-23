
C 		?= kkdemo
CLUSTER 	?= $(C)
P		?= aws vsphere eks
PLATFORMS  	?= $(P)
LOG_FILE   	?= $(BIN)/$(BIN_NAME)_$(CLUSTER).log
ARGS       	?=
L 		?= --log $(LOG_FILE) --debug $(ARGS)
TEST_LOG   	?= $(L)
E 		?= no
EDIT 		?= $(E)

# perform a full integration test creating a cluster in every platform
.PHONY: test-platform-all
test-platform-all:
	@for p in $(PLATFORMS); do \
		make test-platform-$${p} C=$(CLUSTER)-$${p}; \
	done

test-check:
	@if ! env | grep -q KUBEKIT_VAR_; then \
		$(ECHO) "$(C_YELLOW)[ WARN ]$(C_STD) Not found any custom variable, using the defaults."; \
		$(ECHO) "export variables with $(C_YELLOW)export KUBEKIT_VAR_name=value$(C_STD)"; \
	else \
		$(ECHO) "$(C_YELLOW)[ WARN ]$(C_STD) Only found the following variables, make sure they have the correct values and it's all needed to create the cluster:"; \
		env | grep KUBEKIT_VAR_; \
	fi

# perform a full integration test creating a cluster in the given platform
test-platform-%: test-check
	@p=$@; platform=$${p##*-}; $(ECHO) "$(C_GREEN)Testing cluster $(C_YELLOW)$(CLUSTER)$(C_GREEN) on $(C_YELLOW)$${platform}$(C_STD)"
	@$(ECHO) "$(C_YELLOW) You may want to run in another terminal $(C_GREEN)make log-test-platform C=$(C)$(C_STD)";
	@if [[ -e $(BIN)/variables.sh ]]; then \
		$(ECHO) " $(C_BLUE)$(I_BULLET)$(C_GREEN)  Loading cluster configuration variables$(C_STD)"; \
		source $(BIN)/variables.sh; \
	fi; \
	$(ECHO) " $(C_BLUE)$(I_BULLET)$(C_GREEN)  Initializing cluster configuration$(C_STD)"; \
	p=$@; platform=$${p##*-}; $(BIN)/$(BIN_NAME) init $(CLUSTER) --platform $${platform}  $(TEST_LOG)
	@cc=$$($(BIN)/$(BIN_NAME) describe $(CLUSTER) $(TEST_LOG) | grep path | cut -f2 -d: | tr -d ' ')/cluster.yaml; \
	$(ECHO) " $(C_BLUE)$(I_BULLET)$(C_GREEN)  Auto-editing cluster configuration$(C_STD)"; \
	sed -i.bkp "s/ja186051/$${USER}/" $${cc}; \
	sed -i.bkp "s/'\# Required value\. Example: \(.*\)'/\1/" $${cc}
	@if $(BIN)/$(BIN_NAME) login $(CLUSTER) --list $(TEST_LOG) | grep -q '(none)'; then \
		$(ECHO) " $(C_BLUE)$(I_BULLET)$(C_GREEN)  Login cluster into platform manager$(C_STD)"; \
		$(BIN)/$(BIN_NAME) login $(CLUSTER) $(TEST_LOG); \
	fi
	@if [[ "$(EDIT)" == "yes" || "$(EDIT)" == "y" ]]; then \
		$(ECHO) " $(C_BLUE)$(I_BULLET)$(C_GREEN)  Edit cluster configuration file, press any key when ready$(C_STD)"; \
		$(BIN)/$(BIN_NAME) edit $(CLUSTER) $(TEST_LOG); \
		read -s -n 1; \
	fi
	@$(ECHO) " $(C_BLUE)$(I_BULLET)$(C_GREEN)  Creating the cluster $(C_STD)(provisioning and configuration)"
	$(BIN)/$(BIN_NAME) apply --package-file $(BIN)/$(BIN_NAME)-$(VERSION).rpm $(CLUSTER) $(TEST_LOG)
	@echo
	@$(ECHO) " $(C_BLUE)$(I_BULLET)$(C_GREEN)  Cluster information$(C_STD)"
	$(BIN)/$(BIN_NAME) describe $(CLUSTER) $(TEST_LOG)

# test a running cluster
.PHONY: test-cluster-all
test-cluster-all:
	@for p in $(PLATFORMS); do \
		make test-cluster C=$(CLUSTER)-$${p}; \
	done

# test a running cluster
.PHONY: test-cluster
test-cluster:
	@$(ECHO) "$(C_GREEN)Testing running cluster $(C_YELLOW)$(CLUSTER)$(C_STD)"
	@eval $$($(BIN)/$(BIN_NAME) get env $(CLUSTER)); \
	$(ECHO) " $(C_BLUE)$(I_BULLET)$(C_GREEN)  Getting nodes:$(C_STD)\n"; \
	kubectl get nodes; \
	echo "================================================================================"; \
	$(ECHO) " $(C_BLUE)$(I_BULLET)$(C_GREEN)  Getting pods:$(C_STD)\n"; \
	kubectl get pods --all-namespaces; \
	echo "================================================================================"

# destroy the clusters created for a full integration test in every platform
.PHONY: destroy-test-platform-all
destroy-test-platform-all:
	@for p in $(PLATFORMS); do \
		make destroy-test-platform-$${p} C=$(CLUSTER)-$${p}; \
	done

# destroy the clusters created for a full integration test in the given platform
destroy-test-platform-%:
	@$(ECHO) " $(C_YELLOW) You may want to run in another terminal $(C_GREEN)make log-test-platform C=$(C)$(C_STD)"
	@if $(BIN)/$(BIN_NAME) login $(CLUSTER) --list $(TEST_LOG) | grep -q '(none)'; then \
		$(ECHO) " $(C_BLUE)$(I_BULLET)$(C_GREEN)  Login cluster into platform manager$(C_STD)"; \
		$(BIN)/$(BIN_NAME) login $(CLUSTER) $(TEST_LOG); \
	fi
	@$(ECHO) " $(C_BLUE)$(I_BULLET)$(C_GREEN)  Destroying cluster$(C_STD)"
	$(BIN)/$(BIN_NAME) destroy $(CLUSTER) $(TEST_LOG)
	@$(ECHO) " $(C_BLUE)$(I_BULLET)$(C_GREEN)  Cluster information$(C_STD)"
	$(BIN)/$(BIN_NAME) describe $(CLUSTER) $(TEST_LOG)
	@$(ECHO) " $(C_BLUE)$(I_BULLET)$(C_GREEN)  Deleting cluster configuration$(C_STD)"
	$(BIN)/$(BIN_NAME) delete clusters-config $(CLUSTER) $(TEST_LOG)
	@$(ECHO) " $(C_BLUE)$(I_BULLET)$(C_GREEN)  Available clusters configuration$(C_STD)"
	$(BIN)/$(BIN_NAME) get clusters $(CLUSTER) $(TEST_LOG)
	@$(ECHO) "\n$(C_YELLOW) You may want to delete the log file $(C_GREEN)make log-test-platform-delete C=$(C)$(C_STD)"

# prints the log file during tests. Run this in a different terminal
.PHONY: log-test-platform
log-test-platform:
	@[[ -z "$(CLEAN)" ]] || echo > $(LOG_FILE)
	tail -f $(LOG_FILE)

# deletes the log file generated during tests
.PHONY: log-test-platform
log-test-platform-delete:
	$(RM) $(LOG_FILE)

