LDFLAGS += -X main.version=$$(git describe --always --abbrev=40 --dirty)
GO := GO111MODULE=on GO15VENDOREXPERIMENT=1 go

SOURCES_DIRS = . libvirt/
TERRAFORM_LIBVIRT_SRCS = $(shell find $(SOURCES_DIRS) -type f -name '*.go')

default: build

build: gofmtcheck golint vet terraform-fmt
	$(GO) build -ldflags "${LDFLAGS}"

install:
	$(GO) install -ldflags "${LDFLAGS}"

test:
	$(GO) test -v -covermode=count -coverprofile=profile.cov . ./libvirt

# this will ensure that the example we provide are linted and not syntax broken
terraform-fmt:
	terraform fmt
testacc:
	$(GO) test -v .
	./travis/run-tests-acceptance

vet:
	@$(GO) tool vet ${TERRAFORM_LIBVIRT_SRCS}
golint:
	golint ./libvirt

gofmtcheck:
	bash travis/run-gofmt


clean:
	./travis/cleanup.sh

.PHONY: build install test vet fmt golint
