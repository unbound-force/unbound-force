# Contract: Setup Changes

**Branch**: `027-externalize-tool-init` | **Date**: 2026-04-11

## New Functions

### installUV(opts *Options, env doctor.DetectedEnvironment) stepResult

Installs the `uv` Python package manager. Follows the
Homebrew-first pattern with curl fallback.

```go
func installUV(opts *Options, env doctor.DetectedEnvironment) stepResult {
    // 1. Check if already installed
    if _, err := opts.LookPath("uv"); err == nil {
        return stepResult{name: "uv", action: "already installed"}
    }

    // 2. Dry-run
    if opts.DryRun {
        if doctor.HasManager(env, doctor.ManagerHomebrew) {
            return stepResult{name: "uv", action: "dry-run",
                detail: "Would install: brew install uv"}
        }
        return stepResult{name: "uv", action: "dry-run",
            detail: "Would install: curl -LsSf https://astral.sh/uv/install.sh | sh"}
    }

    // 3. Try Homebrew first
    if doctor.HasManager(env, doctor.ManagerHomebrew) {
        if _, err := opts.ExecCmd("brew", "install", "uv"); err != nil {
            return stepResult{name: "uv", action: "failed",
                detail: "brew install failed", err: err}
        }
        return stepResult{name: "uv", action: "installed",
            detail: "via Homebrew"}
    }

    // 4. Fallback to curl|bash (requires --yes or TTY)
    if !opts.YesFlag && !opts.IsTTY() {
        return stepResult{
            name:   "uv",
            action: "skipped",
            detail: "curl|bash install requires --yes flag or interactive terminal",
        }
    }

    if _, err := opts.ExecCmd("bash", "-c",
        "curl -LsSf https://astral.sh/uv/install.sh | sh"); err != nil {
        return stepResult{name: "uv", action: "failed",
            detail: "curl install failed", err: err}
    }
    return stepResult{name: "uv", action: "installed",
        detail: "via curl"}
}
```

**Pattern**: Follows `installOpenCode()` (Homebrew + curl
fallback with interactive guard).

**Error handling**: Homebrew failure is a hard failure. Curl
failure is a hard failure. No Homebrew + no TTY/--yes is a
skip (not a failure).

### installSpecify(opts *Options, env doctor.DetectedEnvironment) stepResult

Installs the `specify` CLI via `uv tool install`. Gated by
`uv` availability.

```go
func installSpecify(opts *Options, env doctor.DetectedEnvironment) stepResult {
    // 1. Check if already installed
    if _, err := opts.LookPath("specify"); err == nil {
        return stepResult{name: "Specify CLI", action: "already installed"}
    }

    // 2. Dry-run
    if opts.DryRun {
        return stepResult{name: "Specify CLI", action: "dry-run",
            detail: "Would install: uv tool install specify-cli"}
    }

    // 3. Check uv availability
    if _, err := opts.LookPath("uv"); err != nil {
        return stepResult{
            name:   "Specify CLI",
            action: "skipped",
            detail: "uv not available — install uv first",
        }
    }

    // 4. Install via uv
    if _, err := opts.ExecCmd("uv", "tool", "install", "specify-cli"); err != nil {
        return stepResult{
            name:   "Specify CLI",
            action: "failed",
            detail: "uv tool install failed — try: uv tool install specify-cli",
            err:    err,
        }
    }
    return stepResult{name: "Specify CLI", action: "installed",
        detail: "via uv"}
}
```

**Pattern**: Follows `installOpenSpec()` (single install
method, gated by package manager availability).

**Dependency chain**: `uv` must be installed before `specify`.
In the step flow, `installUV()` runs at step 7 and
`installSpecify()` runs at step 8. If `installUV()` fails or
is skipped, `installSpecify()` will skip with "uv not
available" message.

## Step Flow Changes

### Before (12 steps)

```go
fmt.Fprintf(opts.Stdout, "  [1/12] OpenCode...\n")
fmt.Fprintf(opts.Stdout, "  [2/12] Gaze...\n")
fmt.Fprintf(opts.Stdout, "  [3/12] Mx F...\n")
fmt.Fprintf(opts.Stdout, "  [4/12] GitHub CLI...\n")
fmt.Fprintf(opts.Stdout, "  [5/12] Node.js...\n")
fmt.Fprintf(opts.Stdout, "  [6/12] OpenSpec CLI...\n")
fmt.Fprintf(opts.Stdout, "  [7/12] Replicator...\n")
fmt.Fprintf(opts.Stdout, "  [8/12] Replicator setup...\n")
fmt.Fprintf(opts.Stdout, "  [9/12] Ollama...\n")
fmt.Fprintf(opts.Stdout, "  [10/12] Dewey...\n")
fmt.Fprintf(opts.Stdout, "  [11/12] golangci-lint...\n")
fmt.Fprintf(opts.Stdout, "  [12/12] govulncheck...\n")
```

### After (14 steps)

```go
fmt.Fprintf(opts.Stdout, "  [1/14] OpenCode...\n")
fmt.Fprintf(opts.Stdout, "  [2/14] Gaze...\n")
fmt.Fprintf(opts.Stdout, "  [3/14] Mx F...\n")
fmt.Fprintf(opts.Stdout, "  [4/14] GitHub CLI...\n")
fmt.Fprintf(opts.Stdout, "  [5/14] Node.js...\n")
fmt.Fprintf(opts.Stdout, "  [6/14] OpenSpec CLI...\n")
fmt.Fprintf(opts.Stdout, "  [7/14] uv...\n")
fmt.Fprintf(opts.Stdout, "  [8/14] Specify CLI...\n")
fmt.Fprintf(opts.Stdout, "  [9/14] Replicator...\n")
fmt.Fprintf(opts.Stdout, "  [10/14] Replicator setup...\n")
fmt.Fprintf(opts.Stdout, "  [11/14] Ollama...\n")
fmt.Fprintf(opts.Stdout, "  [12/14] Dewey...\n")
fmt.Fprintf(opts.Stdout, "  [13/14] golangci-lint...\n")
fmt.Fprintf(opts.Stdout, "  [14/14] govulncheck...\n")
```

### Dependency Gating

The `installSpecify()` call is NOT gated by `uvResult.err`
the way `installOpenSpec()` is gated by `nodeAvailable`. This
is because `installSpecify()` internally checks for `uv` via
`LookPath` — if `uv` is not available (either because
`installUV()` failed or `uv` was already installed but not in
the current PATH), it returns a skip result. This is simpler
than threading the uv result through.

However, if we want to match the Node.js → OpenSpec pattern
exactly, we could gate it:

```go
// Step 7: Install uv.
fmt.Fprintf(opts.Stdout, "  [7/14] uv...\n")
uvResult := installUV(&opts, env)
results = append(results, uvResult)
uvAvailable := uvResult.err == nil && uvResult.action != "failed"

// Step 8: Install Specify CLI (uv-dependent).
if uvAvailable {
    fmt.Fprintf(opts.Stdout, "  [8/14] Specify CLI...\n")
    results = append(results, installSpecify(&opts, env))
} else {
    results = append(results, stepResult{
        name: "Specify CLI", action: "skipped",
        detail: "no uv"})
}
```

**Decision**: Use the gated pattern (matching Node.js →
OpenSpec) for consistency. The step number is still printed
even when skipped, matching the existing behavior.

Wait — looking at the existing code more carefully (lines
194–199), the Node.js → OpenSpec gating does NOT print the
step number when skipped. It silently appends a skip result.
But the step counter still shows `[6/12]` for the next step.
This means the step numbers in the output are sequential
regardless of skips.

Actually, re-reading the code: when `nodeAvailable` is false,
the code does NOT print `[6/12] OpenSpec CLI...` — it just
appends the skip result silently. This means the output jumps
from `[5/12]` to `[7/12]`. This is the existing behavior.

**Decision**: Match the existing pattern exactly. When `uv` is
not available, skip the `Specify CLI` step silently (no step
number printed).

## Test Changes

### New Test Functions

1. `TestInstallUV_AlreadyInstalled` — LookPath succeeds
2. `TestInstallUV_DryRun_Homebrew` — dry-run with Homebrew
3. `TestInstallUV_DryRun_Curl` — dry-run without Homebrew
4. `TestInstallUV_Homebrew` — successful Homebrew install
5. `TestInstallUV_Curl` — successful curl install
6. `TestInstallUV_CurlSkipped` — no Homebrew, no TTY, no --yes
7. `TestInstallSpecify_AlreadyInstalled` — LookPath succeeds
8. `TestInstallSpecify_DryRun` — dry-run mode
9. `TestInstallSpecify_NoUV` — uv not in PATH
10. `TestInstallSpecify_Success` — successful uv tool install
11. `TestInstallSpecify_Failed` — uv tool install fails

### Step Count Assertions

Update all test assertions that reference the step count
(`[N/12]` → `[N/14]`). Search for `"/12]"` in test strings.

## Doctor Changes

No doctor changes needed. The existing `.specify/` existence
check in `checkScaffoldedFiles()` remains valid regardless of
whether the directory was created by embedded assets or
`specify init`.

If desired in a future spec, a `specify` binary check could be
added to the "Core Tools" group, but this is out of scope for
this change.
