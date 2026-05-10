## Why

The DevPod sandbox backend hardcodes `--ide none` in both
`Create()` and `Start()`, preventing users from using VS Code,
Cursor, or other IDEs alongside the OpenCode TUI. DevPod
natively supports multiple IDEs via its `--ide` flag, and
the OpenCode TUI continues to work regardless of which IDE
is selected because it runs as a server inside the container
on port 4096 — the IDE choice only controls what DevPod
launches *after* the container is healthy.

Users who want to browse code visually in VS Code while using
the OpenCode TUI for AI-assisted coding currently have no way
to do this through `uf sandbox`.

## What Changes

Add an `--ide` flag to `uf sandbox create` and
`uf sandbox start` that passes through to `devpod up`.
Default remains `none` for backward compatibility.

## Capabilities

### New Capabilities

- `sandbox create --ide vscode`: Creates a DevPod workspace
  and opens VS Code connected to it. OpenCode TUI remains
  accessible via `uf sandbox attach`.
- `sandbox start --ide vscode`: Resumes a DevPod workspace
  and opens VS Code. Works alongside `--detach`.
- `IDE` field on `Options` struct: Configurable via CLI flag,
  environment variable (`UF_SANDBOX_IDE`), or
  `.uf/sandbox.yaml` (`ide` field).

### Modified Capabilities

- `DevPodBackend.Create()`: Uses `opts.IDE` instead of
  hardcoded `"none"` for the `--ide` argument.
- `DevPodBackend.Start()`: Passes `--ide` to `devpod up`
  when resuming a workspace. Waits for OpenCode server
  health check before TUI attach.
- `Attach()`: Now detects persistent workspaces (DevPod
  and Podman named volumes) before falling back to
  ephemeral container check.
- `Destroy()`: Now handles ephemeral mode directly
  instead of incorrectly routing through
  `ResolveBackend()`.

### Removed Capabilities

- None.

## Impact

- **`internal/sandbox/sandbox.go`**: Add `IDE string` field
  to `Options` struct. Fix `Attach()` to detect persistent
  workspaces. Fix `Destroy()` to handle ephemeral mode.
- **`internal/sandbox/devpod.go`**: Replace hardcoded
  `"none"` with `opts.IDE` in `Create()` and `Start()`.
  Add `waitForHealth()` call in `Start()`.
- **`internal/sandbox/config.go`**: Add `IDE` to
  `DefaultConfig()` resolution (flag > env > config > "none").
- **`internal/sandbox/workspace.go`**: Add `ide` field to
  `SandboxConfig` YAML struct.
- **`internal/scaffold/assets/devcontainer/devcontainer.json`**:
  Add `postStartCommand` to auto-start OpenCode server.
- **`internal/sandbox/sandbox_test.go`**: Tests for IDE
  passthrough, Attach persistent detection, Destroy
  ephemeral handling.
- **CLI wiring**: Add `--ide` flag to `create` and `start`
  cobra commands.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

This change adds a CLI flag passthrough. It does not
affect inter-hero artifact communication.

### II. Composability First

**Assessment**: PASS

The IDE flag is optional with a backward-compatible
default (`none`). VS Code and the OpenCode TUI work
independently and simultaneously — neither requires
the other. Users who don't want IDE integration are
unaffected.

### III. Observable Quality

**Assessment**: N/A

No new output formats or quality claims introduced.

### IV. Testability

**Assessment**: PASS

The `IDE` field is added to the existing `Options`
struct and injected via the standard DI pattern. Tests
verify the flag is passed through to `devpod up`
arguments via the existing `ExecCmd` injection.
