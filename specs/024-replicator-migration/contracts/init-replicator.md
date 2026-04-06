# Contract: Init — Replicator Configuration

**Spec**: 024-replicator-migration | **FR**: FR-007 through FR-010

## Scope

Defines the behavioral contract for `uf init` after the
Replicator migration. Covers `opencode.json` configuration,
legacy plugin migration, and `replicator init` delegation.

## configureOpencodeJSON Behavior (Modified)

### Replicator MCP Entry

| Condition | Action | Result |
|-----------|--------|--------|
| `replicator` not in PATH | Skip replicator entry | No `mcp.replicator` added |
| `replicator` in PATH + no existing entry | Add entry | `mcp.replicator` created |
| `replicator` in PATH + entry exists | Skip | Preserve existing entry |
| `replicator` in PATH + Force | Overwrite | Replace `mcp.replicator` |

The MCP entry structure:
```json
{
  "type": "local",
  "command": ["replicator", "serve"],
  "enabled": true
}
```

### Legacy Plugin Migration

| Condition | Action |
|-----------|--------|
| `plugin` array contains `"opencode-swarm-plugin"` | Remove the entry |
| `plugin` array is now empty after removal | Remove the `plugin` key entirely |
| `plugin` array has other entries besides swarm | Keep array with remaining entries |
| No `plugin` key exists | No action needed |

Migration runs regardless of whether `replicator` is in PATH.
This ensures legacy cleanup happens even before Replicator is
installed.

### Combined Decision Matrix

| hasDewey | hasReplicator | hasLegacyPlugin | Action |
|----------|---------------|-----------------|--------|
| false | false | false | Skip (nothing to configure) |
| true | false | false | Add dewey MCP only |
| false | true | false | Add replicator MCP only |
| true | true | false | Add both MCP entries |
| false | false | true | Remove legacy plugin only |
| true | false | true | Add dewey MCP + remove legacy |
| false | true | true | Add replicator MCP + remove legacy |
| true | true | true | Add both MCP + remove legacy |

**Design decision**: The `hasHive` condition that previously
gated plugin array addition is replaced by `hasReplicator`
(binary in PATH). The `.hive/` directory is no longer a
prerequisite for `opencode.json` configuration — Replicator
creates `.hive/` via `replicator init`, which runs after
`configureOpencodeJSON`.

## initSubTools — Replicator Init Delegation

Added after Dewey init, before `configureOpencodeJSON`:

| Condition | Action | Result |
|-----------|--------|--------|
| `replicator` not in PATH | Skip | No action |
| `.hive/` already exists | Skip | No action |
| DryRun | Report | `{action: "dry-run"}` |
| `replicator init` succeeds | Report | `{action: "initialized"}` |
| `replicator init` fails | Report | `{action: "failed", detail: "replicator init failed"}` |

**Error handling**: Failure is non-blocking (Constitution
Principle II — Composability First). The error is captured
and reported in the summary but does not prevent remaining
sub-tool initialization.

## Plugin Array Removal

The entire `// --- Swarm plugin entry ---` block (lines
560–584 in scaffold.go) MUST be replaced with:

1. Legacy plugin migration logic (remove
   `opencode-swarm-plugin` from `plugin` array)
2. Replicator MCP entry logic (add `mcp.replicator`)

The `plugin` key MUST NOT be added to `opencode.json` by
`uf init` under any circumstances (FR-010).
