# Quickstart: Externalize Tool Initialization

**Branch**: `027-externalize-tool-init` | **Date**: 2026-04-11

## Implementation Order

1. Remove embedded assets (scaffold engine)
2. Add tool delegations to `initSubTools()` (scaffold engine)
3. Update tests (scaffold tests)
4. Add `installUV()` and `installSpecify()` (setup)
5. Update step count and flow (setup)
6. Update setup tests
7. Update AGENTS.md and documentation

## Key Patterns to Copy

### initSubTools Delegation (from Replicator)

The simplest existing delegation to copy is Replicator
(lines 876–889 of `scaffold.go`):

```go
// Replicator: init if binary available and .uf/replicator/ absent.
if _, err := opts.LookPath("replicator"); err == nil {
    replicatorDir := filepath.Join(opts.TargetDir, ".uf", "replicator")
    if _, statErr := os.Stat(replicatorDir); os.IsNotExist(statErr) {
        _, _ = fmt.Fprintf(opts.Stdout, "  Initializing Replicator workspace...\n")
        if _, initErr := opts.ExecCmd("replicator", "init"); initErr != nil {
            results = append(results, subToolResult{
                name: ".uf/replicator/", action: "failed",
                detail: "replicator init failed"})
        } else {
            results = append(results, subToolResult{
                name: ".uf/replicator/", action: "initialized"})
        }
    }
}
```

### installGaze (from setup.go)

The simplest Homebrew install pattern to copy:

```go
func installGaze(opts *Options, env doctor.DetectedEnvironment) stepResult {
    if _, err := opts.LookPath("gaze"); err == nil {
        return stepResult{name: "Gaze", action: "already installed"}
    }
    if opts.DryRun {
        if doctor.HasManager(env, doctor.ManagerHomebrew) {
            return stepResult{name: "Gaze", action: "dry-run",
                detail: "Would install: brew install unbound-force/tap/gaze"}
        }
        return stepResult{name: "Gaze", action: "dry-run",
            detail: "Would install: download from GitHub releases"}
    }
    if !doctor.HasManager(env, doctor.ManagerHomebrew) {
        return stepResult{name: "Gaze", action: "skipped",
            detail: "Homebrew not available. Download from ..."}
    }
    if _, err := opts.ExecCmd("brew", "install", "unbound-force/tap/gaze"); err != nil {
        return stepResult{name: "Gaze", action: "failed",
            detail: "brew install failed", err: err}
    }
    return stepResult{name: "Gaze", action: "installed", detail: "via Homebrew"}
}
```

### installOpenCode (Homebrew + curl fallback)

For `installUV()`, copy the `installOpenCode()` pattern which
has both Homebrew and curl fallback with interactive guard:

```go
func installOpenCode(opts *Options, env doctor.DetectedEnvironment) stepResult {
    // ... LookPath check, dry-run ...
    // Try Homebrew first.
    if doctor.HasManager(env, doctor.ManagerHomebrew) {
        if _, err := opts.ExecCmd("brew", "install", "anomalyco/tap/opencode"); err != nil {
            return stepResult{...failed...}
        }
        return stepResult{...installed via Homebrew...}
    }
    // Fallback to curl|bash — requires --yes or TTY.
    if !opts.YesFlag && !opts.IsTTY() {
        return stepResult{...skipped...}
    }
    if _, err := opts.ExecCmd("bash", "-c", "curl ... | bash"); err != nil {
        return stepResult{...failed...}
    }
    return stepResult{...installed via curl...}
}
```

### Node.js → OpenSpec Dependency Chain

For the uv → Specify dependency chain, copy the Node.js →
OpenSpec pattern (lines 188–199 of `setup.go`):

```go
// Step 5: Ensure Node.js.
nodeResult := ensureNodeJS(&opts, env)
results = append(results, nodeResult)
nodeAvailable := nodeResult.err == nil && nodeResult.action != "failed"

// Step 6: Install OpenSpec CLI (Node.js-dependent).
if nodeAvailable {
    fmt.Fprintf(opts.Stdout, "  [6/12] OpenSpec CLI...\n")
    results = append(results, installOpenSpec(&opts, env))
} else {
    results = append(results, stepResult{
        name: "OpenSpec CLI", action: "skipped",
        detail: "no Node.js"})
}
```

## Critical Implementation Notes

### 1. Asset Removal — Delete Files AND Directory

When removing the 12 Speckit files, delete the entire
`internal/scaffold/assets/specify/` directory tree. Don't
leave empty directories — `embed.FS` embeds directory entries
and the `fs.WalkDir` in `Run()` would still iterate over them.

### 2. OpenSpec Gate — Use config.yaml, Not Directory

The OpenSpec delegation MUST gate on `openspec/config.yaml`
existence, NOT `openspec/` directory existence. The embedded
asset walk creates `openspec/schemas/unbound-force/` before
`initSubTools()` runs. If gated on the directory, the
delegation would always be skipped.

### 3. mapAssetPath — Remove specify/ Case

After removing all specify assets, the `specify/` case in
`mapAssetPath()` is dead code. Remove it and remove
`"specify/"` from `knownAssetPrefixes`. The function still
handles `opencode/` and `openspec/` correctly.

### 4. Test Count — 55 → 42

The `expectedAssetPaths` list drops from 55 to 42 entries.
All tests that assert on `len(expectedAssetPaths)` will
automatically use the new count. No hardcoded "55" to find
and replace.

### 5. Step Count — 12 → 14

All `[N/12]` format strings in `setup.go` must change to
`[N/14]`. There are 12 such strings (one per step). The
renumbering starts at step 7 (uv) — steps 1–6 only change
their denominator.

### 6. Specify Install Gating

The `installSpecify()` step is gated by `uvResult` success,
matching the `nodeAvailable` → `installOpenSpec()` pattern.
If `installUV()` fails, `installSpecify()` is silently
skipped with `detail: "no uv"`.

## Verification Checklist

After implementation, verify:

- [ ] `go test -race -count=1 ./internal/scaffold/...` passes
- [ ] `go test -race -count=1 ./internal/setup/...` passes
- [ ] `go build ./...` succeeds
- [ ] `golangci-lint run` passes
- [ ] `internal/scaffold/assets/specify/` directory does not
      exist
- [ ] `internal/scaffold/assets/openspec/config.yaml` does not
      exist
- [ ] `internal/scaffold/assets/openspec/schemas/unbound-force/`
      still exists with 5 files
- [ ] `uf init` in a fresh directory with all tools installed
      creates `.specify/`, `openspec/config.yaml`, and Gaze
      agent files
- [ ] `uf init` in a fresh directory with NO tools installed
      still creates OpenCode agents, packs, and commands
- [ ] `uf setup --dry-run` shows 14 steps with uv and Specify
