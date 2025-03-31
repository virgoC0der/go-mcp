.PHONY: test lint coverage clean

# Default target
all: lint test

# Run all tests
test:
	go test -v -race ./...

# Run tests with coverage
coverage:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -html=coverage.txt -o coverage.html

# Run linters
lint:
	golangci-lint run ./...

# Clean generated files
clean:
	rm -f coverage.txt
	rm -f coverage.html

# Run example
run-echo:
	go run ./examples/echo/main.go

# Run advanced example
run-advanced:
	go run ./examples/advanced/main.go

# Help
help:
	@echo "Available targets:"
	@echo "  all        - Run lint and tests"
	@echo "  test       - Run tests"
	@echo "  coverage   - Run tests with coverage and generate HTML report"
	@echo "  lint       - Run linters"
	@echo "  clean      - Clean generated files"
	@echo "  run-echo   - Run echo example"
	@echo "  run-advanced - Run advanced example" 