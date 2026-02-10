.PHONY: build install test clean lint help

# Build the Go binary
build:
	@echo "Building ghost-tab-tui..."
	go build -o bin/ghost-tab-tui ./cmd/ghost-tab-tui
	@echo "✓ Built bin/ghost-tab-tui"

# Install to local bin
install: build
	@echo "Installing to ~/.local/bin..."
	mkdir -p $(HOME)/.local/bin
	cp bin/ghost-tab-tui $(HOME)/.local/bin/
	@echo "✓ Installed ghost-tab-tui"

# Run tests
test:
	@echo "Running Go tests..."
	go test ./...
	@echo "Running bash tests..."
	./run-tests.sh

# Run Go tests only
test-go:
	go test -v ./...

# Run bash tests only
test-bash:
	./run-tests.sh

# Clean build artifacts
clean:
	rm -f bin/ghost-tab-tui
	go clean

# Lint Go code
lint:
	@echo "Running golangci-lint..."
	golangci-lint run ./...
	@echo "Running shellcheck..."
	find lib bin ghostty -name '*.sh' -exec shellcheck {} +

# Show help
help:
	@echo "Ghost Tab Build Targets:"
	@echo "  make build   - Build the Go binary"
	@echo "  make install - Install to ~/.local/bin"
	@echo "  make test    - Run all tests (Go + bash)"
	@echo "  make clean   - Remove build artifacts"
	@echo "  make lint    - Run linters"
