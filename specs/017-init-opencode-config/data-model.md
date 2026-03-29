# Data Model: Init OpenCode Config

## opencode.json Structure

The `opencode.json` file at repo root is a JSON object
with three top-level keys:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "<server-name>": {
      "type": "local",
      "command": ["<binary>", "<arg1>", ...],
      "enabled": true
    }
  },
  "plugin": [
    "<plugin-name>"
  ]
}
```

### Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `$schema` | string | Yes | OpenCode config schema URL |
| `mcp` | object | No | Map of MCP server definitions |
| `mcp.<name>` | object | No | Individual MCP server config |
| `mcp.<name>.type` | string | Yes | Transport type (`"local"` for stdio) |
| `mcp.<name>.command` | string[] | Yes | Command and args to launch server |
| `mcp.<name>.enabled` | boolean | No | Whether server is active (default true) |
| `plugin` | string[] | No | List of installed OpenCode plugins |

### Managed Entries

Init manages exactly two entries:

**Dewey MCP Server** (`mcp.dewey`):
```json
{
  "type": "local",
  "command": ["dewey", "serve", "--vault", "."],
  "enabled": true
}
```
- Added when: `dewey` binary is in PATH
- Idempotent: skipped if `mcp.dewey` already exists
- Force: overwritten when `--force` flag is set

**Swarm Plugin** (`plugin` array):
```json
["opencode-swarm-plugin"]
```
- Added when: `.hive/` directory exists
- Idempotent: skipped if array already contains the
  entry
- Force: not affected (array membership check only)

### Preservation Rules

- Unknown top-level keys are preserved (e.g., custom
  user config)
- Unknown MCP server entries are preserved (e.g.,
  `mcp.my-custom-server`)
- Unknown plugin entries are preserved
- Key ordering is deterministic (Go's
  `json.MarshalIndent` on maps sorts keys
  alphabetically). Byte-identical output is required
  on no-change re-runs (FR-016).

## subToolResult

The `configureOpencodeJSON()` function returns a
`subToolResult` with these possible action values:

| Action | Meaning | Summary Symbol |
|--------|---------|:--------------:|
| `"created"` | File did not exist, created with entries | `✓` |
| `"configured"` | File existed, entries added | `✓` |
| `"already configured"` | Both entries already present | `✓` |
| `"overwritten"` | Force mode, entries replaced | `✓` |
| `"skipped"` | Nothing to configure (no tools available) | `—` |
| `"dry-run"` | DryRun mode, no file written | `—` |
| `"error"` | Malformed JSON or read failure (non-fatal) | `✗` |
| `"failed"` | Write error (non-fatal) | `✗` |
