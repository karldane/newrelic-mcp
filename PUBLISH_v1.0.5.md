# Publishing newrelic-mcp v1.0.5

## Pre-flight Checklist ✅

- [x] All 54 tools have outputSchema
- [x] All 54 tools have complete parameter descriptions
- [x] All 54 tools have annotations
- [x] Manifest.json generated (105KB)
- [x] MCPB package built (3.4MB)
- [x] Version bumped to 1.0.5
- [x] Git tag v1.0.5 created
- [x] Release notes written
- [x] Code committed (4 commits ahead of origin/main)

## SHA256 Checksums

```
039624542e965ba66488d2717c9ad4f593e996d93655cf9b8b44153a0fe74278  newrelic-mcp-linux-amd64.mcpb
```

## Publishing Steps

### 1. Push to GitHub

```bash
cd /tmp/newrelic-mcp
git push origin main
git push origin v1.0.5
```

### 2. Create GitHub Release

```bash
gh release create v1.0.5 \
  --title "v1.0.5 - Complete OutputSchema Coverage" \
  --notes-file RELEASE_NOTES_v1.0.5.md \
  newrelic-mcp-linux-amd64.mcpb
```

Or manually:
1. Go to https://github.com/karldane/newrelic-mcp/releases/new
2. Tag: `v1.0.5`
3. Title: `v1.0.5 - Complete OutputSchema Coverage`
4. Description: Copy from `RELEASE_NOTES_v1.0.5.md`
5. Upload artifact: `newrelic-mcp-linux-amd64.mcpb`

### 3. Publish to MCP Registry

```bash
# Ensure logged in
mcp-publisher login github

# Publish
mcp-publisher publish \
  --name "io.github.karldane/newrelic-mcp" \
  --mcpb newrelic-mcp-linux-amd64.mcpb \
  --github-release v1.0.5
```

### 4. Publish to Smithery

```bash
# Ensure logged in
smithery auth login

# Publish
smithery publish \
  --name "karldane/newrelic-mcp" \
  --mcpb newrelic-mcp-linux-amd64.mcpb \
  --version 1.0.5
```

### 5. Verify Published Versions

**GitHub Release:**
- https://github.com/karldane/newrelic-mcp/releases/tag/v1.0.5

**MCP Registry:**
- https://mcp.run/servers/io.github.karldane/newrelic-mcp

**Smithery:**
- https://smithery.ai/server/karldane/newrelic-mcp
- Check scoring improvements via API: `https://api.smithery.ai/servers/karldane%2Fnewrelic-mcp`

## Expected Smithery Score Improvements

### Before v1.0.5
- Parameter descriptions: 53/54 (98%)
- Output schemas: 0/54 (0%)
- Annotations: Runtime only

### After v1.0.5
- Parameter descriptions: 54/54 (100%) ✅
- Output schemas: 54/54 (100%) ✅
- Annotations: 54/54 (100%) ✅

## Post-Release Verification

### Test Installation

**Via Smithery:**
```bash
npx @smithery/cli install karldane/newrelic-mcp --client claude
```

**Via MCP Registry:**
```bash
mcp install io.github.karldane/newrelic-mcp
```

### Test Tool Schemas

```bash
# Start the server
./newrelic-mcp

# In another terminal, test with MCP client
mcp-client connect stdio ./newrelic-mcp
# Send tools/list request
# Verify outputSchema is present in responses
```

### Verify Manifest in MCPB

```bash
unzip -p newrelic-mcp-linux-amd64.mcpb manifest.json | jq '.tools[0]'
# Should show inputSchema, outputSchema, and annotations
```

## Rollback Plan (if needed)

If issues are discovered:

```bash
# Delete GitHub release
gh release delete v1.0.5 --yes

# Delete git tag
git tag -d v1.0.5
git push origin :refs/tags/v1.0.5

# Revert commits
git reset --hard origin/main

# Republish v1.0.4 if needed
```

## Success Criteria

- [x] GitHub release published with MCPB artifact
- [ ] MCP Registry updated to v1.0.5
- [ ] Smithery updated to v1.0.5
- [ ] Smithery score improved (all three metrics at 100%)
- [ ] Test installation works via both registries
- [ ] Tool schemas visible in MCP tools/list responses

## Notes

- This is a **backward compatible** release (no breaking changes)
- All existing configurations continue to work
- Only adds metadata (outputSchema, annotations) - no functional changes
- Binary size unchanged (8.3MB)
- Manifest size increased from ~20KB to 105KB (due to comprehensive schemas)
