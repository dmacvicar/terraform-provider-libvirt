LDFLAGS += -X main.version=$$(git describe --always --abbrev=40 --dirty)

default: build

build: gofmtcheck golint vet
	go build -ldflags "${LDFLAGS}"

install:
	go install -ldflags "${LDFLAGS}"

test:
	go test -v -covermode=count -coverprofile=profile.cov ./libvirt
	go test -v .
testacc:
	go test -v .
	./travis/run-tests-acceptance

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
