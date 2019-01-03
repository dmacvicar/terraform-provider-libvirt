LDFLAGS += -X main.version=$$(git describe --always --abbrev=40 --dirty)

# default  args for tests
TEST_ARGS_DEF := -covermode=count -coverprofile=profile.cov

default: build

build: gofmtcheck golint vet
	go build -ldflags "${LDFLAGS}"

install:
	go install -ldflags "${LDFLAGS}"

# unit tests
# usage:
# - run all the unit tests: make test
# - run some particular test: make test TEST_ARGS="-run TestAccLibvirtDomain_Cpu"
test:
	go test -v $(TEST_ARGS_DEF) $(TEST_ARGS) ./libvirt
	go test -v .

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
	go test -v .
	./travis/run-tests-acceptance $(TEST_ARGS)

vet:
	@echo "go vet ."
	@go vet $$(go list ./... | grep -v vendor/) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

golint:
	golint ./libvirt

gofmtcheck:
	bash travis/run-gofmt

clean:
	./travis/cleanup.sh

.PHONY: build install test vet fmt golint
