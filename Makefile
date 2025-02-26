# Project information
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))

## protoc files
THIRD_PARTY_DIR=$(PROJECT_DIR)/third_party
PROTO_FILES=$(shell find $(PROJECT_DIR) -path $(THIRD_PARTY_DIR) -prune -o -name *.proto -print)

# Build binary
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)
GOPATH=$(shell go env GOPATH)
BIN_EXT ?=

ifeq ($(GOOS),windows)
	BIN_EXT=.exe
	GOPATH := $(subst \,/,$(GOPATH))
endif


# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
GOBIN=$(GOPATH)/bin

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make \033[36m<target>\033[0m\n"} /^[%\/a-zA-Z_._0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
.DEFAULT_GOAL := help

##@ Development
pb: protoc-gen-go pb.clean ## generate /*/*.proto
	@protoc --proto_path=$(PROJECT_DIR) \
		--proto_path=$(THIRD_PARTY_DIR) \
		--go_out=paths=source_relative:$(PROJECT_DIR) \
		$(PROTO_FILES)
	@$(MAKE) pb.publish

pb.publish: ## copy submodule proto file to third_party
	@rm -rf $(THIRD_PARTY_DIR)/corex
	@for file in $(PROTO_FILES); do \
  		m=$${file##*/corex/}; \
  		dir=$${m%/*.proto}; \
  		mkdir -p $(THIRD_PARTY_DIR)/corex/$$dir; \
		echo " * $$file -> $(THIRD_PARTY_DIR)/corex/$$m"; \
		cp $$file $(THIRD_PARTY_DIR)/corex/$$m; \
  	done

pb.clean: ## clean *.pb.go
	@find $(PROJECT_DIR) -name "*.pb.go" -exec rm -rf {} \;

##@ Install
.PHONY: protoc-gen-go
protoc-gen-go: ## Download protoc-gen-go-* plugin to global '$GOPATH/bin' if necessary.
	$(call go-install-tool-global,protoc-gen-go,google.golang.org/protobuf/cmd/protoc-gen-go@latest)


define go-install-tool-global
@[ -f $(GOBIN)/$(1)$(BIN_EXT) ] || { \
set -e ;\
echo "Downloading $(2) to $(GOBIN)" ;\
go install $(2) ;\
}
endef
