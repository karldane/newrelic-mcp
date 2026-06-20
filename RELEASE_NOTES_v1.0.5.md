# Release Notes: newrelic-mcp v1.0.5

## 🎯 Major Enhancement: Complete Tool Schema Coverage

This release adds comprehensive `outputSchema` definitions to all 54 tools, enabling maximum Smithery scoring and better tooling integration.

## ✨ What's New

### Complete Schema Coverage (54/54 tools)
- ✅ All 54 tools now have **inputSchema** with complete parameter descriptions
- ✅ All 54 tools now have **outputSchema** with detailed response structures
- ✅ All 54 tools now have **annotations** (readOnlyHint, idempotentHint, openWorldHint)

### Accurate API Response Schemas
Each `outputSchema` accurately reflects the New Relic NerdGraph/GraphQL API response structure:
- Matches actual field names and types returned by the API
- Includes nested object and array structures
- Provides detailed field descriptions for better discoverability

### Enhanced Files (12 tool files, +1541 lines)
1. **alert_policy_tools.go** - 6 tools with output schemas
2. **dashboard_crud_tools.go** - 4 tools with output schemas
3. **notifications_tools.go** - 5 tools with output schemas
4. **synthetics_tools.go** - 5 tools with output schemas
5. **entity_ops_tools.go** - 2 tools with output schemas
6. **tags_tools.go** - 4 tools with output schemas
7. **workflows_tools.go** - 5 tools with output schemas
8. **workloads_tools.go** - 2 tools with output schemas
9. **service_levels_tools.go** - 3 tools with output schemas
10. **cross_account_tools.go** - 1 tool with output schema
11. **newrelic_tools.go** - 13 tools with output schemas
12. **newrelic.go** - 4 tools with output schemas

### Fixed Issues
- Added missing parameter descriptions for `query_traces` tool (`error_only`, `service_name`)

## 📊 Smithery Scoring Improvements

**Before v1.0.5:**
- Parameter descriptions: 53/54 (98%)
- Output schemas: 0/54 (0%)
- Annotations: Emitted at runtime but not in manifest

**After v1.0.5:**
- Parameter descriptions: 54/54 (100%) ✅
- Output schemas: 54/54 (100%) ✅
- Annotations: 54/54 (100%) ✅

## 🔧 Technical Details

### Manifest Generation
```bash
make generate-manifest  # Auto-generates manifest.json with all schemas
make mcpb              # Creates MCPB package with manifest
```

### Manifest Size
- **manifest.json**: 105KB (includes all tool schemas)
- **MCPB package**: 3.4MB (binary + manifest)

### Schema Pattern
All outputSchemas follow this consistent pattern:
```go
func (t *ToolName) OutputSchema() *mcp.ToolOutputSchema {
	schema := mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"field": map[string]interface{}{
				"type":        "array",
				"description": "Detailed description",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{...},
				},
			},
		},
	}
	return &schema
}
```

## 📦 Downloads

- **MCPB Package**: `newrelic-mcp-linux-amd64.mcpb` (3.4MB)
- **SHA256**: `039624542e965ba66488d2717c9ad4f593e996d93655cf9b8b44153a0fe74278`

## 🚀 Installation

### Via Smithery
```bash
npx @smithery/cli install karldane/newrelic-mcp --client claude
```

### Via MCP Registry
```bash
mcp install io.github.karldane/newrelic-mcp
```

### Direct Download
Download `newrelic-mcp-linux-amd64.mcpb` from this release and extract:
```bash
unzip newrelic-mcp-linux-amd64.mcpb
chmod +x server/newrelic-mcp
```

## 🔗 Links

- **GitHub**: https://github.com/karldane/newrelic-mcp
- **MCP Registry**: https://mcp.run/servers/io.github.karldane/newrelic-mcp
- **Smithery**: https://smithery.ai/server/karldane/newrelic-mcp

## 📝 Full Changelog

**Commits:**
- `d01c842` - Bump version to 1.0.5
- `aff7a5e` - Add OutputSchema to all 54 tools for comprehensive Smithery scoring
- `1982941` - Add automatic manifest generation from --scan-format=manifest

**Dependencies:**
- Updated `mcp-framework` to v0.2.10 (with annotations support)

---

**Contributors**: @karldane
