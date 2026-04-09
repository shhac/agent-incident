BINARY := agent-incident
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/agent-incident

test:
	go test ./... -count=1

test-short:
	go test ./... -count=1 -short

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .
	goimports -w .

clean:
	rm -f $(BINARY)
	rm -rf dist/

dev:
	go run ./cmd/agent-incident $(ARGS)

vet:
	go vet ./...

.PHONY: build test test-short lint fmt clean dev vet
