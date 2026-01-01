.PHONY: build fmt lint test tools

GOBIN ?= $(shell go env GOPATH)/bin
GOLANGCI_LINT ?= $(GOBIN)/golangci-lint

build:
	go build -o terraformat ./cmd/terraformat

fmt:
	gofumpt -w $$(go list -f '{{.Dir}}' ./...)

lint:
	$(GOLANGCI_LINT) run ./...

test:
	go test ./...

tools:
	GOBIN=$(GOBIN) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	GOBIN=$(GOBIN) go install mvdan.cc/gofumpt@latest
