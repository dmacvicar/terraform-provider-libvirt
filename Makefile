default: build

build: gofmtcheck golint vet
	go build

install:
	go install

test:
	go test -v -covermode=count -coverprofile=profile.cov ./libvirt

testacc:
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

.PHONY: build install test vet fmt golint
