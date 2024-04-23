TOOLS_BIN_DIR ?= $(shell pwd)/tmp/bin

export PATH := $(TOOLS_BIN_DIR):$(PATH)

MDOX_BINARY=$(TOOLS_BIN_DIR)/mdox
MDOX_VALIDATE_CONFIG?=.mdox.validate.yaml

TOOLING=$(MDOX_BINARY)

MD_FILES_TO_FORMAT=$(shell ls *.md)

docs: $(MDOX_BINARY)
	@echo ">> formatting and local/remote link check"
	$(MDOX_BINARY) fmt --soft-wraps -l --links.localize.address-regex="https://prometheus-operator.dev/.*" --links.validate.config-file=$(MDOX_VALIDATE_CONFIG) $(MD_FILES_TO_FORMAT)

$(TOOLS_BIN_DIR):
	mkdir -p $(TOOLS_BIN_DIR)

$(TOOLING): $(TOOLS_BIN_DIR)
	@echo Installing tools from scripts/tools.go
	@cat scripts/tools.go | grep _ | awk -F'"' '{print $$2}' | GOBIN=$(TOOLS_BIN_DIR) xargs -tI % go install -mod=readonly -modfile=scripts/go.mod %