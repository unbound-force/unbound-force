# Quickstart: Replicator Migration

**Branch**: `024-replicator-migration` | **Date**: 2026-04-06

## Implementation Order

This migration touches 6 production files and their tests.
The recommended implementation order minimizes broken
intermediate states:

### Phase 1: Setup (US-1, US-4)

1. **setup.go** — Core migration
   - Delete `ensureBun()`, `installSwarmPlugin()`,
     `runSwarmSetup()`, `initializeHive()`, `swarmForkSource`
   - Add `installReplicator()` (copy `installGaze` pattern,
     change binary name and tap formula)
   - Add `runReplicatorSetup()` (copy `runSwarmSetup` pattern,
     change binary name)
   - Modify `installOpenSpec()` — remove bun preference block,
     keep npm-only
   - Modify `Run()` — restructure step flow, update step count
     15→12, remove `runUnboundInit` step
   - Update embedding model note (remove "Swarm" reference)
   - Update package doc comment

2. **setup_test.go** — Update all tests
   - Remove tests for deleted functions
   - Add tests for `installReplicator`, `runReplicatorSetup`
   - Update `installOpenSpec` tests (npm-only)
   - Update integration tests for new step count

### Phase 2: Init (US-2)

3. **scaffold.go** — opencode.json + init delegation
   - Modify `configureOpencodeJSON()`:
     - Replace `hasHive`/plugin-array block with
       `hasReplicator`/MCP-entry block
     - Add legacy plugin migration (remove
       `opencode-swarm-plugin` from `plugin` array)
     - Remove empty `plugin` key
   - Modify `initSubTools()`:
     - Add Replicator init delegation (after Dewey, before
       `configureOpencodeJSON`)
   - Update `configureOpencodeJSON` doc comment

4. **scaffold_test.go** — Update opencode.json tests
   - Update all tests that assert on `plugin` array
   - Add tests for legacy plugin migration
   - Add tests for `mcp.replicator` entry
   - Add test for `replicator init` delegation

### Phase 3: Doctor (US-3)

5. **checks.go** — Replace check group
   - Delete `checkSwarmPlugin()`
   - Add `checkReplicator()` (follow `checkDewey` structure)
   - Update `coreToolSpecs` — replace `"swarm"` with
     `"replicator"`

6. **doctor.go** — Update check group list
   - Replace `checkSwarmPlugin(&opts)` with
     `checkReplicator(&opts)` in `groups` slice

7. **environ.go** — Update install hints
   - `managerInstallCmd()` — remove `ManagerBun` "swarm" case
   - `homebrewInstallCmd()` — replace "swarm" with "replicator"
   - `genericInstallCmd()` — replace "swarm" with "replicator"
   - `installURL()` — add "replicator" entry

8. **doctor_test.go** — Update all tests
   - Remove `checkSwarmPlugin` tests
   - Add `checkReplicator` tests
   - Update install hint assertion strings

### Phase 4: Config + Docs (US-2, FR-014, FR-016)

9. **opencode.json** (repo root) — Update live config
   - Remove `plugin` key
   - Add `mcp.replicator` entry

10. **Agent/command files** — Update install hint text
    - `.opencode/command/unleash.md` — "Swarm plugin" →
      "Replicator"
    - `internal/scaffold/assets/opencode/command/unleash.md` —
      same change (scaffold asset sync)

### Phase 5: Verification

11. Run `make check` — all tests pass
12. Run `go test -race -count=1 ./...` — verify
13. Verify SC-004: grep for stale references
    (`opencode-swarm-plugin`, `ensureBun`,
    `installSwarmPlugin`, `bun add -g`)

## Key Patterns to Copy

### installReplicator (from installGaze)

```go
func installReplicator(opts *Options, env doctor.DetectedEnvironment) stepResult {
    if _, err := opts.LookPath("replicator"); err == nil {
        return stepResult{name: "Replicator", action: "already installed"}
    }
    // ... (same pattern as installGaze with "unbound-force/tap/replicator")
}
```

### checkReplicator (from checkDewey + checkSwarmPlugin)

```go
func checkReplicator(opts *Options) CheckGroup {
    group := CheckGroup{Name: "Replicator", Results: []CheckResult{}}
    // Check 1: binary (LookPath)
    // Check 2: replicator doctor (ExecCmdTimeout, 10s)
    // Check 3: .hive/ (os.Stat)
    // Check 4: mcp.replicator in opencode.json (ReadFile + JSON parse)
    return group
}
```

### MCP entry in configureOpencodeJSON (from Dewey entry)

```go
replicatorEntry := json.RawMessage(`{
    "type": "local",
    "command": ["replicator", "serve"],
    "enabled": true
}`)
```

## Risk Areas

1. **Test count**: ~50 test assertions reference swarm/bun/plugin.
   Systematic find-and-replace is needed but each assertion must
   be verified for semantic correctness (not just string replacement).

2. **Step numbering**: All step numbers in setup output change.
   Tests that assert on `[N/15]` format strings must be updated
   to `[N/12]`.

3. **configureOpencodeJSON complexity**: The function handles
   Dewey, Replicator, and legacy migration in one pass. Keep the
   logic sections clearly separated with comments.

4. **Scaffold asset sync**: The `unleash.md` file exists in both
   `.opencode/command/` (live) and
   `internal/scaffold/assets/opencode/command/` (embedded).
   Both must be updated identically.
