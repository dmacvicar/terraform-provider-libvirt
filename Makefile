default: build

build: gofmtcheck golint vet
	go build

install: 
	go install

test: 
	bash travis/run_acceptance_test.sh
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
	bash travis/gofmtcheck.sh

.PHONY: build install test vet fmt golint
