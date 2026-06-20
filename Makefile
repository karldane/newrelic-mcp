# New Relic MCP Server Makefile

BINARY_NAME=newrelic-mcp
BUILD_DIR=.
LDFLAGS=-ldflags="-s -w" -trimpath

# Default target - downloads dependencies and builds
.PHONY: all
all: deps build

# Download and verify dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	@GOPROXY=direct GOSUMDB=off go mod tidy
	@echo "Dependencies ready"

# Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"
	@du -h $(BUILD_DIR)/$(BINARY_NAME) | cut -f1

# Build for multiple platforms
.PHONY: build-all
build-all: deps build-linux build-darwin build-windows

.PHONY: build-linux
build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .

.PHONY: build-darwin
build-darwin:
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .

.PHONY: build-windows
build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .

# Run tests
.PHONY: test
test:
	go test ./newrelic -v

# Build MCP Bundle (.mcpb ZIP with manifest.json for Smithery)
.PHONY: mcpb
mcpb: build-linux
	@echo "Creating $(BINARY_NAME)-linux-amd64.mcpb..."
	@mkdir -p /tmp/mcpb-build/server
	cp $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 /tmp/mcpb-build/server/$(BINARY_NAME)
	printf '{"manifestVersion":"0.3","server":{"type":"binary","entryPoint":"server/$(BINARY_NAME)"}}\n' > /tmp/mcpb-build/manifest.json
	cd /tmp/mcpb-build && zip -q -X ../$(BINARY_NAME)-linux-amd64.mcpb manifest.json server/$(BINARY_NAME)
	rm -rf /tmp/mcpb-build
	@echo "SHA256:"
	@openssl dgst -sha256 $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64.mcpb

# Publish a release to GitHub, the MCP Registry, and Smithery.
# Auto-bumps the patch version from the latest tag.
.PHONY: publish
publish: clean mcpb
	@set -e; \
	LATEST=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	MAJOR=$$(echo "$$LATEST" | sed 's/^v//' | cut -d. -f1); \
	MINOR=$$(echo "$$LATEST" | sed 's/^v//' | cut -d. -f2); \
	PATCH=$$(echo "$$LATEST" | sed 's/^v//' | cut -d. -f3); \
	PATCH=$$((PATCH + 1)); \
	VERSION="v$$MAJOR.$$MINOR.$$PATCH"; \
	VNUM=$${VERSION#v}; \
	echo "Auto-bumped $$LATEST -> $$VERSION"; \
	git tag "$$VERSION"; \
	git push origin "$$VERSION"; \
	gh release create "$$VERSION" "$(BUILD_DIR)/$(BINARY_NAME)-linux-amd64.mcpb" --title "$$VERSION" --generate-notes; \
	smithery mcp publish "$(BUILD_DIR)/$(BINARY_NAME)-linux-amd64.mcpb" -n karldane/newrelic-mcp; \
	echo "=== Done: $$VERSION published ==="

# Clean build artifacts
.PHONY: clean
clean:
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
	rm -f $(BUILD_DIR)/$(BINARY_NAME)-*

# Install locally
.PHONY: install
install: build
	go install $(LDFLAGS) .

# Show help
.PHONY: help
help:
	@echo "New Relic MCP Server"
	@echo ""
	@echo "Usage:"
	@echo "  make              - Download dependencies and build binary"
	@echo "  make deps         - Download and verify dependencies"
	@echo "  make build        - Build the binary"
	@echo "  make mcpb         - Build .mcpb release artifact (ZIP + manifest.json)"
	@echo "  make build-all    - Build for all platforms (Linux, macOS, Windows)"
	@echo "  make test         - Run tests"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make install      - Install binary to GOPATH/bin"
	@echo "  make publish      - Bump tag, release to GitHub, publish to Smithery"
	@echo "  make help         - Show this help message"
