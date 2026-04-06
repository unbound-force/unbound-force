# Research: Replicator Migration

**Branch**: `024-replicator-migration` | **Date**: 2026-04-06
**Purpose**: Resolve all technical unknowns before implementation.

## R1: Homebrew Install Pattern (installGaze)

**Question**: What is the canonical pattern for installing a Go
binary via Homebrew in `uf setup`?

**Finding**: The `installGaze()` function in `setup.go` (lines
426–450) is the reference pattern:

1. Check if binary is in PATH via `opts.LookPath("gaze")`
2. If found, return `stepResult{action: "already installed"}`
3. If `opts.DryRun`, return dry-run result with Homebrew command
4. If Homebrew not available, return `skipped` with GitHub
   releases download link
5. Run `brew install unbound-force/tap/<binary>`
6. Return `installed` or `failed` result

**Replicator application**: `installReplicator()` follows this
exact pattern with `brew install unbound-force/tap/replicator`.
No fallback to npm/bun needed (Go binary, not Node.js package).

**Status**: RESOLVED

---

## R2: MCP Binary + Config Check Pattern (checkDewey)

**Question**: What is the canonical pattern for checking an MCP
server binary and its configuration in `uf doctor`?

**Finding**: The `checkDewey()` function in `checks.go` (lines
1003–1098) is the reference pattern:

1. Create a `CheckGroup` with a descriptive name
2. Check binary via `opts.LookPath("dewey")`
3. If not found, return informational Pass results for all
   sub-checks with "skipped: not installed" messages
4. If found, check additional components:
   - Embedding model (via `ollama list`)
   - Embedding capability (via `opts.EmbedCheck`)
   - Workspace directory (via `os.Stat`)
5. Each sub-check returns its own `CheckResult`

**Replicator application**: `checkReplicator()` follows this
pattern with:
- Binary check: `opts.LookPath("replicator")`
- Doctor delegation: `opts.ExecCmdTimeout(10*time.Second, "replicator", "doctor")`
- `.hive/` check: `os.Stat(filepath.Join(opts.TargetDir, ".hive"))`
- MCP config check: parse `opencode.json` for `mcp.replicator`

The `checkSwarmPlugin()` function (lines 330–461) provides the
existing swarm-specific pattern that will be replaced. Key
differences from Dewey:
- Uses `ExecCmdTimeout` for `swarm doctor` (10s timeout)
- Checks `.hive/` directory
- Checks `plugin` array in `opencode.json`

**Status**: RESOLVED

---

## R3: MCP Server Entry Pattern (Dewey in scaffold.go)

**Question**: How does `uf init` add an MCP server entry to
`opencode.json`?

**Finding**: The `configureOpencodeJSON()` function in
`scaffold.go` (lines 435–628) handles this:

1. Detect if binary is in PATH: `opts.LookPath("dewey")`
2. Read existing `opencode.json` or create new map
3. Check for existing entry in both `mcp` and legacy
   `mcpServers` keys
4. If not present (or Force), add entry to `mcp` map:
   ```json
   {
     "type": "local",
     "command": ["dewey", "serve", "--vault", "."],
     "enabled": true
   }
   ```
5. Write file with 2-space indent + trailing newline

**Replicator application**: Add a `replicator` entry to the
`mcp` map following the same structure:
```json
{
  "type": "local",
  "command": ["replicator", "serve"],
  "enabled": true
}
```

**Migration**: When a legacy `opencode-swarm-plugin` entry exists
in the `plugin` array, remove it. If the `plugin` array becomes
empty, remove the `plugin` key entirely. This is safe because
OpenCode does not require a `plugin` key to exist (per spec
assumption).

**Status**: RESOLVED

---

## R4: Sub-Tool Init Delegation Pattern (dewey init)

**Question**: How does `uf init` delegate initialization to a
sub-tool?

**Finding**: The `initSubTools()` function in `scaffold.go`
(lines 653–749) handles this:

1. Check if binary is in PATH: `opts.LookPath("dewey")`
2. Check if workspace already exists: `os.Stat(deweyDir)`
3. If workspace absent, run init: `opts.ExecCmd("dewey", "init")`
4. Report result as `subToolResult`
5. Errors are captured as warnings, not hard failures
   (Constitution Principle II — Composability First)

**Replicator application**: Add Replicator init delegation:
1. Check `opts.LookPath("replicator")`
2. Check `os.Stat(".hive/")`
3. If `.hive/` absent, run `opts.ExecCmd("replicator", "init")`
4. Report as `subToolResult`

**Status**: RESOLVED

---

## R5: Setup Step Count and Flow

**Question**: What is the current step flow and how does it
change?

**Finding**: Current 15-step flow:

| Step | Name | Function | Keep? |
|------|------|----------|-------|
| 1 | OpenCode | `installOpenCode` | ✓ |
| 2 | Gaze | `installGaze` | ✓ |
| 3 | Mx F | `installMxF` | ✓ |
| 4 | GitHub CLI | `installGH` | ✓ |
| 5 | Node.js | `ensureNodeJS` | ✓ |
| 6 | Bun | `ensureBun` | ✗ REMOVE |
| 7 | OpenSpec CLI | `installOpenSpec` | ✓ (modify: npm-only) |
| 8 | Swarm plugin | `installSwarmPlugin` | ✗ REMOVE |
| 9 | Swarm setup | `runSwarmSetup` | ✗ REMOVE |
| 10 | .hive/ | `initializeHive` | ✗ REMOVE |
| 11 | Ollama | `installOllama` | ✓ |
| 12 | Dewey | `installDewey` | ✓ |
| 13 | golangci-lint | `installGolangciLint` | ✓ |
| 14 | govulncheck | `installGovulncheck` | ✓ |
| 15 | uf init | `runUnboundInit` | ✓ |

New 12-step flow:

| Step | Name | Function |
|------|------|----------|
| 1 | OpenCode | `installOpenCode` |
| 2 | Gaze | `installGaze` |
| 3 | Mx F | `installMxF` |
| 4 | GitHub CLI | `installGH` |
| 5 | Node.js | `ensureNodeJS` |
| 6 | OpenSpec CLI | `installOpenSpec` (npm-only) |
| 7 | Replicator | `installReplicator` (NEW) |
| 8 | Replicator setup | `runReplicatorSetup` (NEW) |
| 9 | Ollama | `installOllama` |
| 10 | Dewey | `installDewey` |
| 11 | golangci-lint | `installGolangciLint` |
| 12 | govulncheck | `installGovulncheck` |

Key changes:
- Steps 6 (Bun), 8 (Swarm plugin), 9 (Swarm setup), 10 (.hive/)
  are removed (4 steps removed)
- Steps 7 (Replicator) and 8 (Replicator setup) are added
  (2 steps added)
- Step 15 (uf init) is removed from setup — `uf init` handles
  per-repo configuration including `opencode.json` and
  `replicator init` (1 step removed)
- Net: 15 - 4 + 2 - 1 = 12 steps
- OpenSpec CLI moves from step 7 to step 6 and drops bun
  preference (npm-only)
- Replicator does NOT depend on Node.js, so it can be installed
  independently of the Node.js step

**Design decision**: Remove `uf init` from setup. Setup is
per-machine; init is per-repo. Replicator init (creating `.hive/`)
is a per-repo operation that belongs in `uf init`, not `uf setup`.
This matches the Dewey pattern where `dewey init` runs inside
`initSubTools()` (called by `uf init`), not during `uf setup`.

**Status**: RESOLVED

---

## R6: OpenSpec CLI npm-only Migration

**Question**: Can OpenSpec CLI be installed via npm without bun?

**Finding**: The current `installOpenSpec()` function (lines
392–423) tries bun first, falls back to npm. The spec assumption
states: "OpenSpec CLI (`@fission-ai/openspec`) works correctly
when installed via npm without bun."

The npm install command is: `npm install -g @fission-ai/openspec@latest`

This is already the fallback path in the current code. The change
is to remove the bun preference and use npm directly.

**Status**: RESOLVED

---

## R7: Install Hint Updates

**Question**: Which install hints reference swarm/bun and need
updating?

**Finding**: Searched `environ.go` for all swarm/bun references:

1. `managerInstallCmd()` line 248: `ManagerBun` case for "swarm"
   returns `"bun add -g github:unbound-force/swarm-tools"` →
   REMOVE this case entirely
2. `homebrewInstallCmd()` line 272: "swarm" case returns
   `"npm install -g github:unbound-force/swarm-tools"` →
   REPLACE with "replicator" case returning
   `"brew install unbound-force/tap/replicator"`
3. `genericInstallCmd()` line 293: "swarm" case returns
   `"npm install -g github:unbound-force/swarm-tools"` →
   REPLACE with "replicator" case returning
   `"brew install unbound-force/tap/replicator"`

Also in `checks.go`:
4. `checkSwarmPlugin()` line 343: install hint
   `"npm install -g github:unbound-force/swarm-tools"` →
   Entire function replaced by `checkReplicator()`

**Status**: RESOLVED

---

## R8: Agent/Command File References

**Question**: Which agent and command files reference "Swarm
plugin" and need text updates?

**Finding**: Searched scaffold assets and live `.opencode/` files:

1. `.opencode/command/unleash.md` (lines 42, 393): References
   "Swarm plugin" in graceful degradation context. These refer
   to the Swarm MCP tools (like `swarm_worktree_list`), not the
   npm plugin package. The tool names don't change — Replicator
   provides the same MCP tools. However, the install hint text
   should change from "Install the Swarm plugin" to "Install
   Replicator".

2. Corresponding scaffold asset at
   `internal/scaffold/assets/opencode/command/unleash.md` must
   be updated in sync.

**Status**: RESOLVED

---

## R9: opencode.json Live File Update

**Question**: What changes to the live `opencode.json` at repo
root?

**Finding**: Current content:
```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "dewey": { ... }
  },
  "plugin": ["opencode-swarm-plugin"]
}
```

Target content:
```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "dewey": { ... },
    "replicator": {
      "type": "local",
      "command": ["replicator", "serve"],
      "enabled": true
    }
  }
}
```

The `plugin` key is removed entirely.

**Status**: RESOLVED

---

## R10: Test Impact Analysis

**Question**: Which existing tests will break and need updating?

**Finding**:

**setup_test.go** (high impact — ~20 test assertions):
- Tests referencing `"npm install -g github:unbound-force/swarm-tools"`
- Tests referencing `"bun add -g github:unbound-force/swarm-tools"`
- Tests referencing `"opencode-swarm-plugin"` in opencode.json
- Tests referencing step count `[N/15]`
- Tests for `ensureBun`, `installSwarmPlugin`, `runSwarmSetup`,
  `initializeHive` functions
- Tests for `installOpenSpec` bun preference

**doctor_test.go** (medium impact — ~10 test assertions):
- Tests referencing `"npm install -g github:unbound-force/swarm-tools"`
- Tests referencing `"bun add -g github:unbound-force/swarm-tools"`
- Tests referencing `"opencode-swarm-plugin"` in opencode.json
- Tests for `checkSwarmPlugin` function
- Tests for install hint strings

**scaffold_test.go** (medium impact — ~15 test assertions):
- Tests referencing `"opencode-swarm-plugin"` in plugin array
- Tests for `configureOpencodeJSON` function
- `TestScaffoldOutput_*` regression tests (may need new
  regression test for no swarm references)

**Status**: RESOLVED

---

## R11: Core Tools List Update

**Question**: Does the `coreToolSpecs` list in `checks.go` need
updating?

**Finding**: The current `coreToolSpecs` (lines 88–125) includes
`"swarm"` as an optional tool (not required, not recommended).
This should be replaced with `"replicator"` with the same
classification (optional/informational). The `"swarm"` entry
currently has no version check — just binary presence.

**Status**: RESOLVED

---

## R12: Replicator Binary Assumptions

**Question**: What subcommands does the `replicator` binary
expose?

**Finding**: Per spec assumptions:
- `replicator serve` — starts MCP server on stdio
- `replicator setup` — per-machine initialization (idempotent)
- `replicator doctor` — health check (text output)
- `replicator init` — per-repo setup, creates `.hive/`

These match the patterns used by Dewey (`dewey serve`,
`dewey init`) and the old Swarm (`swarm setup`, `swarm init`,
`swarm doctor`).

**Status**: RESOLVED
