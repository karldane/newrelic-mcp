# New Relic MCP Server Makefile
#
# Reusable pattern for Go-based MCP servers. Override the config
# variables below (or via environment) when copying to a new project.

# === Project-specific configuration (override per project) ===
BINARY_NAME      ?= newrelic-mcp
SERVER_NAME      ?= io.github.karldane/newrelic-mcp
SMITHERY_SERVER  ?= karldane/newrelic-mcp
REPO_OWNER       ?= karldane
REPO_NAME        ?= newrelic-mcp
# ============================================================

BUILD_DIR=.
LDFLAGS=-ldflags="-s -w" -trimpath

# Default target
.PHONY: all
all: deps build

# Download and verify dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	@GOPROXY=direct GOSUMDB=off go mod tidy
	@echo "Dependencies ready"

# Build the binary for the host platform
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"
	@du -h $(BUILD_DIR)/$(BINARY_NAME) | cut -f1

# Build for multiple platforms (useful before a release)
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

# Install binary to GOPATH/bin
.PHONY: install
install: build
	go install $(LDFLAGS) .

# Build .mcpb release artifact (always linux/amd64)
.PHONY: mcpb
mcpb:
	@echo "Building $(BINARY_NAME)-linux-amd64.mcpb..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	cp $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64.mcpb
	@echo "SHA256:"
	@openssl dgst -sha256 $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64.mcpb

# Publish a release to GitHub, the MCP Registry, and Smithery.
# Automatically bumps the patch version from the latest tag.
#
# Prerequisites (user must auth separately — no keys in repo):
#   - GitHub CLI:        gh auth login
#   - MCP Publisher:     mcp-publisher login github
#   - Smithery CLI:      smithery auth login
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
	git tag "$$VERSION"; \
	echo "Auto-bumped $$LATEST → $$VERSION"; \
	SHA256=$$(openssl dgst -sha256 "$(BUILD_DIR)/$(BINARY_NAME)-linux-amd64.mcpb" | cut -d' ' -f2); \
	RELEASE_URL="https://github.com/$(REPO_OWNER)/$(REPO_NAME)/releases/download/$$VERSION/$(BINARY_NAME)-linux-amd64.mcpb"; \
	echo "Target: $$VERSION ($$VNUM)"; \
	echo ""; \
	\
	echo "=== Checking prerequisites ==="; \
	command -v gh >/dev/null 2>&1 || { echo "FATAL: GitHub CLI not found — see https://cli.github.com"; exit 1; }; \
	command -v mcp-publisher >/dev/null 2>&1 || { echo "FATAL: mcp-publisher not found — run: brew install mcp-publisher"; exit 1; }; \
	command -v smithery >/dev/null 2>&1 || { echo "FATAL: Smithery CLI not found — run: npm install -g @smithery/cli"; exit 1; }; \
	gh auth status -h github.com >/dev/null 2>&1 || { echo "FATAL: not logged into GitHub CLI — run: gh auth login"; exit 1; }; \
	echo "All tools found."; \
	echo ""; \
	\
	echo "=== 1/4: Creating GitHub Release ==="; \
	gh release create "$$VERSION" "$(BUILD_DIR)/$(BINARY_NAME)-linux-amd64.mcpb" \
		--title "$$VERSION" \
		--generate-notes; \
	echo ""; \
	\
	echo "=== 2/4: Patching server.json ==="; \
	cp server.json server.json.bak; \
	sed -i 's|"version": *"[^"]*"|"version": "'$$VNUM'"|' server.json; \
	sed -i 's|"fileSha256": *"[^"]*"|"fileSha256": "'$$SHA256'"|' server.json; \
	sed -i 's|"identifier": *"[^"]*"|"identifier": "'$$RELEASE_URL'"|' server.json; \
	echo "  version: $$VNUM"; \
	echo "  sha256:  $$SHA256"; \
	echo "  url:     $$RELEASE_URL"; \
	echo ""; \
	\
	echo "=== 3/4: Publishing to MCP Registry ==="; \
	mcp-publisher publish; \
	echo ""; \
	\
	echo "=== 4/4: Publishing to Smithery ==="; \
	smithery mcp publish "$(BUILD_DIR)/$(BINARY_NAME)-linux-amd64.mcpb" -n "$(SMITHERY_SERVER)"; \
	echo ""; \
	\
	echo "=== Done: $$VERSION published to MCP Registry and Smithery ==="; \
	mv server.json.bak server.json

# Show help
.PHONY: help
help:
	@echo "$(REPO_NAME) — MCP Server"
	@echo ""
	@echo "Usage:"
	@echo "  make              - Download dependencies and build binary"
	@echo "  make deps         - Download and verify dependencies"
	@echo "  make build        - Build the binary"
	@echo "  make mcpb         - Build .mcpb release artifact (linux/amd64)"
	@echo "  make build-all    - Build for all platforms (Linux, macOS, Windows)"
	@echo "  make test         - Run tests"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make install      - Install binary to GOPATH/bin"
	@echo "  make publish      - Publish tagged release (gh + mcp-publisher + smithery)"
	@echo "  make help         - Show this help message"
	@echo ""
	@echo "Publish config:"
	@echo "  SERVER_NAME:      $(SERVER_NAME)"
	@echo "  SMITHERY_SERVER:  $(SMITHERY_SERVER)"
	@echo "  REPO:             $(REPO_OWNER)/$(REPO_NAME)"
