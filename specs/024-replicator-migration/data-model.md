# Data Model: Replicator Migration

**Branch**: `024-replicator-migration` | **Date**: 2026-04-06

## Overview

This feature modifies no persistent data models. All changes are
to runtime configuration (`opencode.json`) and CLI behavior. No
new Go types, schemas, or artifact formats are introduced.

## opencode.json Schema Changes

### Before (Swarm Plugin)

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "dewey": {
      "type": "local",
      "command": ["dewey", "serve", "--vault", "."],
      "enabled": true
    }
  },
  "plugin": [
    "opencode-swarm-plugin"
  ]
}
```

### After (Replicator MCP)

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "dewey": {
      "type": "local",
      "command": ["dewey", "serve", "--vault", "."],
      "enabled": true
    },
    "replicator": {
      "type": "local",
      "command": ["replicator", "serve"],
      "enabled": true
    }
  }
}
```

### Migration Rules

| Condition | Action |
|-----------|--------|
| `plugin` array contains `"opencode-swarm-plugin"` | Remove the entry |
| `plugin` array is now empty | Remove the `plugin` key |
| `mcp.replicator` does not exist | Add the entry |
| `mcp.replicator` already exists | Preserve (idempotent) |
| Both legacy plugin AND `mcp.replicator` exist | Remove plugin, keep MCP |

## Function Signature Changes

### setup.go — Functions Removed

```go
// REMOVED: ensureBun, installSwarmPlugin, runSwarmSetup, initializeHive
// REMOVED: const swarmForkSource
```

### setup.go — Functions Added

```go
// installReplicator installs Replicator via Homebrew.
// Follows installGaze() pattern.
func installReplicator(opts *Options, env doctor.DetectedEnvironment) stepResult

// runReplicatorSetup runs `replicator setup` for per-machine init.
func runReplicatorSetup(opts *Options) stepResult
```

### setup.go — Functions Modified

```go
// installOpenSpec — remove bun preference, npm-only.
func installOpenSpec(opts *Options, env doctor.DetectedEnvironment) stepResult

// Run — update step count 15→12, restructure step flow.
func Run(opts Options) error
```

### scaffold.go — Functions Modified

```go
// configureOpencodeJSON — replace plugin array logic with
// mcp.replicator entry. Add legacy plugin migration.
func configureOpencodeJSON(opts *Options) []subToolResult

// initSubTools — add replicator init delegation.
func initSubTools(opts *Options) []subToolResult
```

### checks.go — Functions Removed

```go
// REMOVED: checkSwarmPlugin
```

### checks.go — Functions Added

```go
// checkReplicator checks the Replicator installation:
// binary, doctor delegation, .hive/, MCP config.
func checkReplicator(opts *Options) CheckGroup
```

### doctor.go — Functions Modified

```go
// Run — replace checkSwarmPlugin with checkReplicator in
// groups slice.
func Run(opts Options) (*Report, error)
```

### environ.go — Functions Modified

```go
// managerInstallCmd — remove ManagerBun "swarm" case.
// homebrewInstallCmd — replace "swarm" with "replicator".
// genericInstallCmd — replace "swarm" with "replicator".
// coreToolSpecs — replace "swarm" with "replicator".
```

## Existing Types Unchanged

The following types are used but not modified:

- `doctor.Options` — all existing fields sufficient
- `doctor.CheckGroup` / `doctor.CheckResult` — reused as-is
- `scaffold.Options` — all existing fields sufficient
- `scaffold.subToolResult` — reused as-is
- `setup.Options` — all existing fields sufficient
- `setup.stepResult` — reused as-is
