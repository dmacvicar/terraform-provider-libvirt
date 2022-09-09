HOSTNAME=registry.terraform.io
NAMESPACE=dmacvicar
NAME=libvirt
BINARY=terraform-provider-${NAME}
VERSION?=$(shell git describe --tags | sed -E 's,^v,,g')
OS_ARCH=$(shell go env GOOS)_$(shell go env GOARCH)

LDFLAGS += -X main.version=${VERSION}

# default  args for tests
TEST_ARGS_DEF := -covermode=count -coverprofile=profile.cov

default: build

terraform-provider-libvirt:
	go build -ldflags "${LDFLAGS}"

build: terraform-provider-libvirt

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

uninstall:
	rm -rf ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}

# unit tests
# usage:
# - run all the unit tests: make test
# - run some particular test: make test TEST_ARGS="-run TestAccLibvirtDomain_Cpu"
test:
	go test -v $(TEST_ARGS_DEF) $(TEST_ARGS) ./libvirt

# acceptance tests
# usage:
#
# - run all the acceptance tests:
#   make testacc
#
# - run some particular test:
#   make testacc TEST_ARGS="-run TestAccLibvirtDomain_Cpu"
#
# - run all the network test with a verbose loglevel:
#   TF_LOG=DEBUG make testacc TEST_ARGS="-run TestAccLibvirtNet*"
#
testacc:
	./travis/run-tests-acceptance $(TEST_ARGS)

golangcilint:
	golangci-lint run

tflint:
	terraform fmt -write=false -check=true -diff=true examples/

lint: golangcilint tflint

clean:
	rm -f terraform-provider-libvirt

cleanup:
	./travis/cleanup.sh

.PHONY: build install test testacc tflint golangcilint lint terraform-provider-libvirt
