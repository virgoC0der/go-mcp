.PHONY: build test vet tidy clean

BIN := bin/oceanengine-mcp
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

build:
	go build -ldflags "-X main.version=$(VERSION)" -o $(BIN) ./cmd/oceanengine-mcp

test:
	go test ./...

vet:
	go vet ./...

tidy:
	go mod tidy

clean:
	rm -rf bin
