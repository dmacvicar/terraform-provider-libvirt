.PHONY: help build install test testacc sweep lint fmt clean

# Default target
.DEFAULT_GOAL := help

# Version can be overridden
VERSION ?= dev

# Build flags
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

help: ## Display this help message
	@echo "Terraform Provider Libvirt - Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: lint ## Build the provider binary
	@echo "Building terraform-provider-libvirt..."
	@go build $(LDFLAGS) -o terraform-provider-libvirt

install: build ## Install the provider to local Terraform plugin directory
	@echo "Installing to local Terraform plugin directory..."
	@mkdir -p ~/.terraform.d/plugins/registry.terraform.io/dmacvicar/libvirt/$(VERSION)/linux_amd64/
	@cp terraform-provider-libvirt ~/.terraform.d/plugins/registry.terraform.io/dmacvicar/libvirt/$(VERSION)/linux_amd64/
	@echo "Installed to ~/.terraform.d/plugins/registry.terraform.io/dmacvicar/libvirt/$(VERSION)/linux_amd64/"

test: lint ## Run unit tests
	@echo "Running unit tests..."
	@go test ./... -v

testacc: lint ## Run acceptance tests (requires running libvirt)
	@echo "Running acceptance tests..."
	@TF_ACC=1 go test -v -timeout 10m ./internal/provider

sweep: ## Clean up leaked test resources from failed tests
	@echo "Running test sweepers..."
	@go test -sweep=$(shell if [ -n "$$LIBVIRT_TEST_URI" ]; then echo "$$LIBVIRT_TEST_URI"; else echo "qemu:///system"; fi) -timeout 10m ./internal/provider

lint: ## Run golangci-lint
	@echo "Running golangci-lint..."
	@golangci-lint run ./...

fmt: ## Format code with gofmt
	@echo "Formatting code..."
	@gofmt -w -s .

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -f terraform-provider-libvirt
	@go clean

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

.PHONY: check
check: lint vet test ## Run all checks (lint, vet, test)
	@echo "All checks passed!"

docs: ## Generate provider documentation
	@echo "Installing tfplugindocs..."
	@go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest
	@echo "Generating documentation..."
	@PATH="$(PATH):$(shell go env GOPATH)/bin" tfplugindocs generate
