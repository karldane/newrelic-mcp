#!/bin/bash
set -e

# Generate manifest.json from template + scan output
# Usage: ./generate-manifest.sh

BINARY="./newrelic-mcp-linux-amd64"
TEMPLATE="manifest-template.json"
OUTPUT="manifest.json"

if [ ! -f "$BINARY" ]; then
    echo "Error: $BINARY not found. Run 'make build-linux' first."
    exit 1
fi

if [ ! -f "$TEMPLATE" ]; then
    echo "Error: $TEMPLATE not found"
    exit 1
fi

if ! command -v jq &> /dev/null; then
    echo "Error: jq is required but not installed"
    exit 1
fi

echo "Scanning tools with --scan-format=manifest..."
SCAN_OUTPUT=$(NEWRELIC_API_KEY=dummy $BINARY --scan --scan-format=manifest 2>&1)

if [ $? -ne 0 ]; then
    echo "Error: Failed to scan tools"
    echo "$SCAN_OUTPUT"
    exit 1
fi

echo "Extracting tools array..."
TOOLS=$(echo "$SCAN_OUTPUT" | jq '.tools')

if [ "$TOOLS" == "null" ] || [ -z "$TOOLS" ]; then
    echo "Error: No tools found in scan output"
    exit 1
fi

echo "Merging with template..."
jq --argjson tools "$TOOLS" '. + {tools: $tools}' "$TEMPLATE" > "$OUTPUT"

TOOL_COUNT=$(echo "$TOOLS" | jq 'length')
echo "Generated $OUTPUT with $TOOL_COUNT tools"
