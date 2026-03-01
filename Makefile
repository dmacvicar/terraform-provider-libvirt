.PHONY: help build install test testacc testacc-tofu sweep lint fmt clean generate testdeps-acc testdeps-acc-tinylinux

# Default target
.DEFAULT_GOAL := help

# Version can be overridden
VERSION ?= dev
CIRROS_VERSION ?= 0.6.2
TEST_IMAGE_DIR ?= .cache/test-images
CIRROS_IMAGE ?= $(TEST_IMAGE_DIR)/cirros-$(CIRROS_VERSION)-x86_64-disk.img
CIRROS_URL ?= https://download.cirros-cloud.net/$(CIRROS_VERSION)/cirros-$(CIRROS_VERSION)-x86_64-disk.img
ALPINE_TINY_VERSION ?= 3.23.3
ALPINE_TINY_RELEASE ?= r0
ALPINE_TINY_VHD ?= $(TEST_IMAGE_DIR)/aws_alpine-$(ALPINE_TINY_VERSION)-x86_64-bios-tiny-$(ALPINE_TINY_RELEASE).vhd
ALPINE_TINY_QCOW2 ?= $(TEST_IMAGE_DIR)/alpine-$(ALPINE_TINY_VERSION)-x86_64-bios-tiny-$(ALPINE_TINY_RELEASE).qcow2
ALPINE_TINY_URL ?= https://dl-cdn.alpinelinux.org/alpine/latest-stable/releases/cloud/aws_alpine-$(ALPINE_TINY_VERSION)-x86_64-bios-tiny-$(ALPINE_TINY_RELEASE).vhd

# Build flags
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

help: ## Display this help message
	@echo "Terraform Provider Libvirt - Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

generate: ## Run code generation
	@echo "Running code generator..."
	@go run ./internal/codegen

build: generate ## Build the provider binary
	@echo "Building terraform-provider-libvirt..."
	@go build $(LDFLAGS) -o terraform-provider-libvirt

install: build ## Install the provider to local Terraform plugin directory
	@echo "Installing to local Terraform plugin directory..."
	@mkdir -p ~/.terraform.d/plugins/registry.terraform.io/dmacvicar/libvirt/$(VERSION)/linux_amd64/
	@cp terraform-provider-libvirt ~/.terraform.d/plugins/registry.terraform.io/dmacvicar/libvirt/$(VERSION)/linux_amd64/
	@echo "Installed to ~/.terraform.d/plugins/registry.terraform.io/dmacvicar/libvirt/$(VERSION)/linux_amd64/"

test: generate ## Run unit tests
	@echo "Running unit tests..."
	@go test ./... -v

testacc: generate ## Run acceptance tests (requires running libvirt)
	@echo "Running acceptance tests..."
	@TF_ACC=1 go test -count=1 -v -timeout 10m ./internal/provider

testdeps-acc: ## Download/cache acceptance test image dependencies (CirrOS)
	@echo "Preparing acceptance test image dependencies..."
	@mkdir -p $(TEST_IMAGE_DIR)
	@if [ ! -f "$(CIRROS_IMAGE)" ]; then \
		echo "Downloading $(CIRROS_URL) -> $(CIRROS_IMAGE)"; \
		curl -fL --retry 3 --retry-delay 2 -o "$(CIRROS_IMAGE)" "$(CIRROS_URL)"; \
	else \
		echo "Using cached image: $(CIRROS_IMAGE)"; \
	fi
	@echo "Set LIBVIRT_TEST_ACPI_IMAGE=$(CIRROS_IMAGE) to run image-based shutdown ACC test."

testdeps-acc-tinylinux: ## Download/cache Tiny Linux acceptance test image (Alpine BIOS tiny)
	@echo "Preparing Tiny Linux acceptance test image..."
	@mkdir -p $(TEST_IMAGE_DIR)
	@if [ ! -f "$(ALPINE_TINY_VHD)" ]; then \
		echo "Downloading $(ALPINE_TINY_URL) -> $(ALPINE_TINY_VHD)"; \
		curl -fL --retry 3 --retry-delay 2 -o "$(ALPINE_TINY_VHD)" "$(ALPINE_TINY_URL)"; \
	else \
		echo "Using cached VHD image: $(ALPINE_TINY_VHD)"; \
	fi
	@if [ ! -f "$(ALPINE_TINY_QCOW2)" ]; then \
		if ! command -v qemu-img >/dev/null 2>&1; then \
			echo "qemu-img is required to convert Tiny Linux VHD to QCOW2"; \
			exit 1; \
		fi; \
		echo "Converting $(ALPINE_TINY_VHD) -> $(ALPINE_TINY_QCOW2)"; \
		qemu-img convert -f vpc -O qcow2 "$(ALPINE_TINY_VHD)" "$(ALPINE_TINY_QCOW2)"; \
	else \
		echo "Using cached QCOW2 image: $(ALPINE_TINY_QCOW2)"; \
	fi
	@echo "Set LIBVIRT_TEST_ACPI_IMAGE=$(ALPINE_TINY_QCOW2) to run image-based shutdown ACC test."

testacc-tofu: ## Run acceptance tests with OpenTofu
	@TF_ACC_TERRAFORM_PATH=$$(which tofu) TF_ACC_PROVIDER_NAMESPACE=dmacvicar TF_ACC_PROVIDER_HOST=registry.terraform.io $(MAKE) testacc

sweep: ## Clean up leaked test resources from failed tests
	@echo "Running test sweepers..."
	@cd internal/provider && go test -sweep=$(shell if [ -n "$$LIBVIRT_TEST_URI" ]; then echo "$$LIBVIRT_TEST_URI"; else echo "qemu:///system"; fi) -timeout 10m .

lint: generate ## Run golangci-lint
	@echo "Verifying golangci-lint config..."
	@golangci-lint config verify
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

check: lint vet test ## Run all checks (lint, vet, test)
	@echo "All checks passed!"

docs: ## Generate provider documentation
	@echo "Installing tfplugindocs..."
	@go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest
	@echo "Generating documentation..."
	@PATH="$(PATH):$(shell go env GOPATH)/bin" tfplugindocs generate

.PHONY: check docs
