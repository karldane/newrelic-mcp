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

# Clean build artifacts
.PHONY: clean
clean:
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
	rm -f $(BUILD_DIR)/$(BINARY_NAME)-*

# Install locally
.PHONY: install
install: build
	go install $(LDFLAGS) .

# Build .mcpb release artifact
.PHONY: mcpb
mcpb: build
	@echo "Packaging $(BINARY_NAME).mcpb..."
	cp $(BUILD_DIR)/$(BINARY_NAME) $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64.mcpb
	@echo "SHA256:"
	@openssl dgst -sha256 $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64.mcpb

# Publish to MCP Registry (requires mcp-publisher)
# Steps:
#   1. make build
#   2. make mcpb     (prints SHA256 for server.json)
#   3. gh release create v1.0.0 newrelic-mcp-linux-amd64.mcpb
#   4. Update fileSha256 in server.json
#   5. mcp-publisher publish
.PHONY: publish
publish: mcpb
	@echo ""
	@echo "=== Next Steps ==="
	@echo "1. Create GitHub Release: gh release create v1.0.0 $(BINARY_NAME)-linux-amd64.mcpb"
	@echo "2. Update fileSha256 in server.json with the SHA256 above"
	@echo "3. Run: mcp-publisher publish"
	@echo ""

# Show help
.PHONY: help
help:
	@echo "New Relic MCP Server"
	@echo ""
	@echo "Usage:"
	@echo "  make              - Download dependencies and build binary"
	@echo "  make deps         - Download and verify dependencies"
	@echo "  make build        - Build the binary"
	@echo "  make mcpb         - Build binary and create .mcpb release artifact"
	@echo "  make build-all    - Build for all platforms (Linux, macOS, Windows)"
	@echo "  make test         - Run tests"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make install      - Install binary to GOPATH/bin"
	@echo "  make publish      - Build, create .mcpb, and print publish instructions"
	@echo "  make help         - Show this help message"
