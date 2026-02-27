.PHONY: build install test clean lint release sync-version help

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

# Create a new release (tag, GitHub release with binaries)
release:
	@bash scripts/release.sh

# Sync package.json version with VERSION file
sync-version:
	@node -e "const p=require('./package.json');const v=require('fs').readFileSync('VERSION','utf8').trim();p.version=v;require('fs').writeFileSync('package.json',JSON.stringify(p,null,2)+'\n');console.log('Synced package.json to '+v)"

# Show help
help:
	@echo "Ghost Tab Build Targets:"
	@echo "  make build   - Build the Go binary"
	@echo "  make install - Install to ~/.local/bin"
	@echo "  make test    - Run all tests (Go + bash)"
	@echo "  make clean   - Remove build artifacts"
	@echo "  make lint    - Run linters"
	@echo "  make release - Create a new release"
