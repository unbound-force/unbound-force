# Contract: Doctor — Replicator Health Checks

**Spec**: 024-replicator-migration | **FR**: FR-011 through FR-013

## Scope

Defines the behavioral contract for `uf doctor` after the
Replicator migration. Covers the Replicator check group,
install hints, and bun reference removal.

## checkReplicator Check Group

Group name: `"Replicator"` (replaces `"Swarm Plugin"`)

### Check 1: replicator binary

| Condition | Severity | Message | InstallHint |
|-----------|----------|---------|-------------|
| `replicator` in PATH | Pass | `"installed"` (with path detail) | — |
| `replicator` not in PATH | Warn | `"not found"` | `"brew install unbound-force/tap/replicator"` |

**Design decision**: Severity is `Warn` (not `Fail`) because
Replicator is optional per Constitution Principle II
(Composability First). This matches the Dewey pattern where
absence is informational.

When `replicator` is not found, remaining checks are skipped
(same early-return pattern as `checkDewey`).

### Check 2: replicator doctor

| Condition | Severity | Message | InstallHint |
|-----------|----------|---------|-------------|
| `replicator doctor` succeeds | Pass | (embedded output) | — |
| `replicator doctor` times out | Warn | `"replicator doctor timed out"` | `"Run replicator doctor manually"` |
| `replicator doctor` fails | Warn | `"replicator doctor reported issues"` | `"Run: uf setup"` |

Uses `opts.ExecCmdTimeout(10*time.Second, ...)` for timeout
protection (same as current `swarm doctor` pattern).

Successful output is stored in `group.Embed` for display.

### Check 3: .hive/ directory

| Condition | Severity | Message | InstallHint |
|-----------|----------|---------|-------------|
| `.hive/` exists and is directory | Pass | `"initialized"` | — |
| `.hive/` does not exist | Warn | `"not initialized"` | `"Run: uf init"` |

Note: Install hint changed from `"Run: swarm init"` to
`"Run: uf init"` because `replicator init` is now delegated
through `uf init`.

### Check 4: MCP config

| Condition | Severity | Message | InstallHint |
|-----------|----------|---------|-------------|
| `opencode.json` not found | Warn | `"opencode.json not found"` | `"Run: uf init"` |
| `opencode.json` unparseable | Warn | `"opencode.json could not be parsed"` | `"Fix JSON syntax in opencode.json"` |
| `mcp.replicator` exists | Pass | `"mcp.replicator in opencode.json"` | — |
| `mcp.replicator` missing | Warn | `"mcp.replicator not in opencode.json"` | `"Run: uf init"` |

**Design decision**: Checks `mcp` key (not `plugin` array).
No fallback to legacy `mcpServers` key for Replicator — only
Dewey has a legacy key history.

## coreToolSpecs Update

Replace the `"swarm"` entry with `"replicator"`:

```go
{
    name: "replicator",
    // Not required, not recommended — optional/informational
}
```

## Install Hint Updates (environ.go)

### managerInstallCmd

Remove the `ManagerBun` case for `"swarm"`:
```go
// REMOVED:
// case ManagerBun:
//     switch toolName {
//     case "swarm":
//         return "bun add -g github:unbound-force/swarm-tools"
//     }
```

### homebrewInstallCmd

Replace `"swarm"` with `"replicator"`:
```go
case "replicator":
    return "brew install unbound-force/tap/replicator"
```

### genericInstallCmd

Replace `"swarm"` with `"replicator"`:
```go
case "replicator":
    return "brew install unbound-force/tap/replicator"
```

Note: Generic fallback also uses Homebrew because Replicator
is a Go binary distributed exclusively via Homebrew (no npm
fallback).

### installURL

Add `"replicator"` entry:
```go
case "replicator":
    return "https://github.com/unbound-force/replicator"
```

## Bun Reference Removal (FR-013)

After migration, zero references to bun MUST remain in install
hints for any tool. The `ManagerBun` detection in `environ.go`
remains (bun is still a valid package manager for user projects),
but no Unbound Force tool uses bun as an install method.

## doctor.go Check Group Order

```go
groups := []CheckGroup{
    checkDetectedEnvironment(env),
    checkCoreTools(&opts, env),
    checkReplicator(&opts),      // was: checkSwarmPlugin(&opts)
    checkDewey(&opts),
    checkScaffoldedFiles(&opts),
    checkHeroAvailability(&opts),
    checkMCPConfig(&opts),
    checkAgentSkillIntegrity(&opts),
}
```

The position is unchanged — Replicator occupies the same slot
as the former Swarm Plugin group.
