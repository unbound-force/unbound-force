## Context

`uf setup` installs 7 tools but misses 3 (mxf, gh,
openspec CLI). It installs Dewey but doesn't initialize
it. `uf init` scaffolds files but doesn't initialize
sub-tools or guide the user on next steps.

## Goals / Non-Goals

### Goals
- `uf setup` installs ALL ecosystem tools
- `uf setup` runs `dewey init` + `dewey index`
- `uf init` runs `dewey init` + `dewey index` when
  Dewey is available (idempotent -- skips if already
  done by `uf setup`)
- `printSummary` shows context-aware next steps
- All changes follow existing patterns exactly
- Full test coverage for new steps

### Non-Goals
- Starting Ollama or Dewey servers (user manages)
- Running `/speckit.constitution` (interactive,
  requires OpenCode agent context)
- Installing Go itself (assumed pre-existing)
- Adding new dependencies to go.mod

## Coverage Strategy

All new functions use injected `LookPath`/`ExecCmd`
for testability. Tests are unit tests using the
existing `cmdRecorder` and `stubLookPath` patterns.

- **Coverage target**: 100% of new function branches
  (success, already-installed, skip, failure, dry-run)
- **Test classification**: All unit tests. No
  integration or e2e tests needed.
- **Coverage ratchet**: New code must not reduce
  overall package coverage (currently 86.6% for setup,
  89.2% for scaffold).

## Decisions

### New Install Functions

Three new functions following existing patterns.
All include DryRun handling matching their pattern
source.

**`installMxF()`** -- follows `installGaze()` exactly
(setup.go:307-332):
```go
func installMxF(opts *Options, env doctor.DetectedEnvironment) stepResult {
    if _, err := opts.LookPath("mxf"); err == nil {
        return stepResult{name: "Mx F", action: "already installed"}
    }
    if opts.DryRun {
        if doctor.HasManager(env, doctor.ManagerHomebrew) {
            return stepResult{name: "Mx F", action: "dry-run",
                detail: "Would install: brew install unbound-force/tap/mxf"}
        }
        return stepResult{name: "Mx F", action: "dry-run",
            detail: "Would install: download from GitHub releases"}
    }
    if !doctor.HasManager(env, doctor.ManagerHomebrew) {
        return stepResult{name: "Mx F", action: "skipped",
            detail: "Homebrew not available. Download from https://github.com/unbound-force/unbound-force/releases"}
    }
    if _, err := opts.ExecCmd("brew", "install", "unbound-force/tap/mxf"); err != nil {
        return stepResult{name: "Mx F", action: "failed",
            detail: "brew install failed", err: err}
    }
    return stepResult{name: "Mx F", action: "installed",
        detail: "via Homebrew"}
}
```

**`installGH()`** -- follows `installGaze()` pattern:
```go
func installGH(opts *Options, env doctor.DetectedEnvironment) stepResult {
    if _, err := opts.LookPath("gh"); err == nil {
        return stepResult{name: "GitHub CLI", action: "already installed"}
    }
    if opts.DryRun {
        if doctor.HasManager(env, doctor.ManagerHomebrew) {
            return stepResult{name: "GitHub CLI", action: "dry-run",
                detail: "Would install: brew install gh"}
        }
        return stepResult{name: "GitHub CLI", action: "dry-run",
            detail: "Would install: download from https://cli.github.com"}
    }
    if !doctor.HasManager(env, doctor.ManagerHomebrew) {
        return stepResult{name: "GitHub CLI", action: "skipped",
            detail: "Homebrew not available. Download from https://cli.github.com"}
    }
    if _, err := opts.ExecCmd("brew", "install", "gh"); err != nil {
        return stepResult{name: "GitHub CLI", action: "failed",
            detail: "brew install failed", err: err}
    }
    return stepResult{name: "GitHub CLI", action: "installed",
        detail: "via Homebrew"}
}
```

**`installOpenSpec()`** -- follows `installSwarmPlugin()`
pattern (setup.go:395-418) with bun preference:
```go
func installOpenSpec(opts *Options, env doctor.DetectedEnvironment) stepResult {
    if _, err := opts.LookPath("openspec"); err == nil {
        return stepResult{name: "OpenSpec CLI", action: "already installed"}
    }
    if opts.DryRun {
        if doctor.HasManager(env, doctor.ManagerBun) {
            return stepResult{name: "OpenSpec CLI", action: "dry-run",
                detail: "Would install: bun add -g @fission-ai/openspec@latest"}
        }
        return stepResult{name: "OpenSpec CLI", action: "dry-run",
            detail: "Would install: npm install -g @fission-ai/openspec@latest"}
    }
    // Prefer bun, fall back to npm (matching installSwarmPlugin pattern)
    if doctor.HasManager(env, doctor.ManagerBun) {
        if _, err := opts.ExecCmd("bun", "add", "-g", "@fission-ai/openspec@latest"); err == nil {
            return stepResult{name: "OpenSpec CLI", action: "installed",
                detail: "via bun"}
        }
    }
    if _, err := opts.ExecCmd("npm", "install", "-g", "@fission-ai/openspec@latest"); err != nil {
        return stepResult{name: "OpenSpec CLI", action: "failed",
            detail: "npm install failed — try: sudo npm install -g @fission-ai/openspec@latest",
            err: err}
    }
    return stepResult{name: "OpenSpec CLI", action: "installed",
        detail: "via npm"}
}
```

### Dewey Init Steps

Two new steps after Dewey binary + model install.
Both include DryRun handling and use `opts.TargetDir`
for correct working directory context. Note: `env` is
not needed (no Homebrew/version manager logic) --
follows the `runSwarmSetup()` precedent.

**`initDewey()`** -- runs `dewey init` if `.dewey/`
doesn't exist:
```go
func initDewey(opts *Options) stepResult {
    deweyDir := filepath.Join(opts.TargetDir, ".dewey")
    if info, err := os.Stat(deweyDir); err == nil && info.IsDir() {
        return stepResult{name: ".dewey/", action: "already initialized"}
    }
    if _, err := opts.LookPath("dewey"); err != nil {
        return stepResult{name: ".dewey/", action: "skipped",
            detail: "dewey not installed"}
    }
    if opts.DryRun {
        return stepResult{name: ".dewey/", action: "dry-run",
            detail: "Would run: dewey init"}
    }
    if _, err := opts.ExecCmd("dewey", "init"); err != nil {
        return stepResult{name: ".dewey/", action: "failed",
            detail: "dewey init failed", err: err}
    }
    return stepResult{name: ".dewey/", action: "initialized"}
}
```

**`indexDewey()`** -- runs `dewey index` if `.dewey/`
exists:
```go
func indexDewey(opts *Options) stepResult {
    deweyDir := filepath.Join(opts.TargetDir, ".dewey")
    if _, err := os.Stat(deweyDir); os.IsNotExist(err) {
        return stepResult{name: "dewey index", action: "skipped",
            detail: "no .dewey/ workspace"}
    }
    if _, err := opts.LookPath("dewey"); err != nil {
        return stepResult{name: "dewey index", action: "skipped",
            detail: "dewey not installed"}
    }
    if opts.DryRun {
        return stepResult{name: "dewey index", action: "dry-run",
            detail: "Would run: dewey index"}
    }
    if _, err := opts.ExecCmd("dewey", "index"); err != nil {
        return stepResult{name: "dewey index", action: "failed",
            detail: "dewey index failed — ensure Ollama server is running (ollama serve)",
            err: err}
    }
    return stepResult{name: "dewey index", action: "completed"}
}
```

**Working directory note**: `ExecCmd` uses
`CombinedOutput()` which inherits the process CWD.
Since `uf setup` is run from the project root and
`opts.TargetDir` defaults to CWD, the commands
execute in the correct directory. The `.dewey/`
existence check uses `filepath.Join(opts.TargetDir, ...)`
for explicit path resolution. This matches the
existing `initializeHive()` pattern which also uses
`opts.TargetDir` for the existence check but CWD for
the subprocess.

### Revised Step Order

Steps 5-11 are conditional on Node.js availability
(inside the `nodeAvailable` block).

```
 1. OpenCode         (brew)
 2. Gaze             (brew)
 3. Mx F             (brew)            NEW
 4. GitHub CLI       (brew)            NEW
 ── if Node.js available ──
 5. Node.js          (version managers)
 6. Bun              (npm)
 7. OpenSpec CLI     (bun/npm)         NEW
 8. Swarm plugin     (bun/npm)
 9. swarm setup
10. opencode.json
11. .hive/           (swarm init)
 ── end conditional ──
12. Ollama           (brew)
13. Dewey            (brew + model pull)
14. .dewey/          (dewey init)      NEW
15. dewey index                        NEW
16. uf init          (scaffold)
```

### Scaffold Options Expansion

Add two optional fields to `scaffold.Options`:

```go
type Options struct {
    TargetDir   string
    Force       bool
    DivisorOnly bool
    Lang        string
    Version     string
    Stdout      io.Writer
    LookPath    func(string) (string, error)              // NEW
    ExecCmd     func(string, ...string) ([]byte, error)   // NEW
}
```

In `Run()`, default these to `exec.LookPath` and
a simple `exec.Command(...).CombinedOutput()` wrapper
if nil. **This MUST happen at the top of `Run()`
before any code path can call `initSubTools()`.**

### Deduplication: setup vs scaffold dewey init

**Setup owns init, scaffold is idempotent.** When
`uf setup` calls `runUnboundInit()` → `scaffold.Run()`
→ `initSubTools()`, the `.dewey/` directory already
exists from setup step 14. The existence check in
`initSubTools` causes it to skip. No double execution.

When `uf init` is called standalone (not via setup),
`initSubTools` runs dewey init for the first time.

This is safe because:
1. `dewey init` is idempotent (creates `.dewey/` only
   if absent)
2. The `.dewey/` existence check prevents re-init
3. No TOCTOU race -- single-threaded execution

### `runUnboundInit` forwarding

`runUnboundInit()` in setup.go MUST forward
`opts.LookPath` and `opts.ExecCmd` to the
`scaffold.Options` struct to maintain the testability
injection chain:

```go
result, err := scaffold.Run(scaffold.Options{
    TargetDir: opts.TargetDir,
    Stdout:    opts.Stdout,
    LookPath:  opts.LookPath,   // Forward
    ExecCmd:   opts.ExecCmd,     // Forward
})
```

### initSubTools in scaffold.go

After scaffolding files and before `printSummary()`,
call `initSubTools()`. Errors are captured and reported
as warnings in `printSummary`, not hard failures (per
Constitution Principle II -- Composability First).

Skip in `DivisorOnly` mode (deploying reviewer assets
to an external repo should not initialize Dewey).

```go
func initSubTools(opts *Options) []subToolResult {
    if opts.DivisorOnly {
        return nil
    }

    var results []subToolResult

    // Dewey: init + index if binary available
    if _, err := opts.LookPath("dewey"); err == nil {
        deweyDir := filepath.Join(opts.TargetDir, ".dewey")
        if _, err := os.Stat(deweyDir); os.IsNotExist(err) {
            if _, initErr := opts.ExecCmd("dewey", "init"); initErr != nil {
                results = append(results, subToolResult{
                    name: ".dewey/", action: "failed",
                    detail: "dewey init failed"})
                return results // skip index if init failed
            }
            results = append(results, subToolResult{
                name: ".dewey/", action: "initialized"})

            if _, idxErr := opts.ExecCmd("dewey", "index"); idxErr != nil {
                results = append(results, subToolResult{
                    name: "dewey index", action: "failed",
                    detail: "dewey index failed"})
            } else {
                results = append(results, subToolResult{
                    name: "dewey index", action: "completed"})
            }
        }
    }

    return results
}
```

### Updated printSummary

`printSummary` receives sub-tool results as a parameter
and uses `opts.LookPath` (passed via a `toolsAvailable`
struct or boolean flags pre-computed in `Run()`).

```
uf init: 49 files processed (N created, N updated, N skipped)

Sub-tool initialization:
  ✅ Dewey workspace initialized (.dewey/)
  ✅ Dewey index built

Next steps:
  1. Run /speckit.constitution to create your project constitution
  2. Run uf doctor to verify your environment
  3. Run /speckit.specify to start a strategic spec
  4. Run /opsx:propose to start a tactical change
```

When Dewey is not available:

```
Next steps:
  1. Run uf setup to install the full toolchain
  2. Run /speckit.constitution to create your project constitution
  3. Run uf doctor to verify your environment
```

The `printSummary` signature expands to accept
sub-tool results. All existing call sites (2
production, 4 test) must be updated.

## Risks / Trade-offs

**Low risk**: All changes follow established patterns.
No new architectures, no new packages, no new
dependencies.

**`dewey index` may take a few seconds**: For large
repos (200+ files), `dewey index` takes 2-5 seconds.
This is acceptable for a one-time init operation. For
repos with no Markdown files, `dewey index` completes
instantly (empty index is a success, not a failure).

**`npm install -g` may require sudo on some systems**:
The error message includes a permissions hint. The bun
fallback avoids this issue when bun is available.

**`gh` may already be installed via system package
manager**: The `LookPath` check handles this -- if `gh`
is already in PATH from any source, we skip.

**Ollama server dependency**: `dewey index` requires
the Ollama server to be running for embedding
generation. If Ollama was just installed by `uf setup`
but not started, `dewey index` will fail. The error
message tells the user to run `ollama serve`. This
matches the existing `pullEmbeddingModel()` pattern.

**`openspec` not in `uf doctor`**: This change installs
`openspec` but does not add a doctor check for it. A
follow-up change should add `openspec` to
`coreToolSpecs` in `internal/doctor/checks.go` to
maintain the setup↔doctor symmetry. Tracked as a
known gap.
