SRC := ../../templates/resources
DST := ../../resources/code.go
EXCLUDE := --exclude "*.new,*.not-ready"

TMP_SRC := ./templates/resources
TMP_DST := code.go

.PHONY: generate
generate:
	go run . --src $(SRC) --dst $(DST) $(EXCLUDE)

.PHONY: reverse
reverse:
	# $(RM) -r $(TMP_DST) $(TMP_SRC)
	$(RM) -r $(TMP_SRC)
	mkdir -p $(TMP_SRC)
	@echo "Genereating the code.go for internal use (if it isn't there)"
	go run . --src $(SRC) --dst $(TMP_DST) --internal
	@echo "Genereating the manifests from given code.go"
	go run . --src $(TMP_SRC) --dst $(TMP_DST) --reverse
