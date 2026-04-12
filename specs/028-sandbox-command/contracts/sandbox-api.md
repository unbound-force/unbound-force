# Contract: Sandbox API

**Package**: `internal/sandbox`
**Date**: 2026-04-12

## Public Functions

### Start

```go
// Start launches a sandbox container with the project directory
// mounted. Checks prerequisites (Podman, OpenCode), detects
// platform, pulls the image if needed, starts the container,
// waits for the health check, and attaches the TUI (unless
// Detach is true).
//
// Returns an error if:
// - Podman is not in PATH (FR-001)
// - A sandbox is already running (FR-016)
// - Container start fails
// - Health check times out after 60 seconds (FR-005)
// - opencode attach fails (when Detach is false)
func Start(opts Options) error
```

**Preconditions**:
- `podman` in PATH
- `opencode` in PATH (unless Detach is true)
- No existing `uf-sandbox` container running

**Postconditions**:
- Container `uf-sandbox` is running
- OpenCode server is healthy on port 4096
- TUI is attached (unless Detach is true)

**Error behaviors**:
| Condition | Error message | FR |
|-----------|--------------|-----|
| Podman missing | "podman not found. Install: brew install podman or https://podman.io" | FR-001 |
| Already running | "sandbox already running. Use `uf sandbox attach` or `uf sandbox stop` first." | FR-016 |
| Image pull fails | "failed to pull image <image>: <error>" | FR-003 |
| Health timeout | "health check timed out after 60s. Check container logs: podman logs uf-sandbox" | FR-005 |
| Attach fails | "failed to attach: <error>. Connect manually: opencode attach http://localhost:4096" | FR-006 |

---

### Stop

```go
// Stop stops and removes the sandbox container.
// Returns nil if no container is running (idempotent).
func Stop(opts Options) error
```

**Preconditions**: None (idempotent).

**Postconditions**:
- No container named `uf-sandbox` exists.

**Error behaviors**:
| Condition | Behavior | FR |
|-----------|----------|-----|
| No container | Print "no sandbox to stop." Return nil. | FR-014 |
| Stop fails | Return wrapped error | FR-014 |

---

### Attach

```go
// Attach connects the TUI to the running sandbox's OpenCode
// server via `opencode attach`.
func Attach(opts Options) error
```

**Preconditions**:
- `opencode` in PATH
- Container `uf-sandbox` is running

**Postconditions**:
- TUI is connected to the container's OpenCode server.

**Error behaviors**:
| Condition | Error message | FR |
|-----------|--------------|-----|
| No container | "no sandbox running. Run `uf sandbox start`." | FR-013 |
| OpenCode missing | "opencode not found. Install: brew install anomalyco/tap/opencode" | FR-013 |

---

### Extract

```go
// Extract generates a patch from the container's git history,
// presents it for review, and applies it to the host repo on
// confirmation.
func Extract(opts Options) error
```

**Preconditions**:
- Container `uf-sandbox` is running
- Container has commits beyond the mount point

**Postconditions** (on confirmation):
- Patch applied to host repo via `git am`

**Error behaviors**:
| Condition | Error message | FR |
|-----------|--------------|-----|
| No container | "no sandbox running." | FR-010 |
| No changes | "no changes to extract." | FR-010 |
| User declines | Exit cleanly, no patch applied. | FR-012 |
| git am conflict | "patch conflict: <details>. Run `git am --abort` to undo." | FR-012 |
| Direct mode | "sandbox is in direct mode — changes are already on the host filesystem." | — |

---

### Status

```go
// Status returns the current state of the sandbox container.
// Returns a zero-value ContainerStatus with Running=false if
// no container exists.
func Status(opts Options) (ContainerStatus, error)
```

**Preconditions**: None.

**Postconditions**: None (read-only).

**Output format** (when running):
```
Sandbox Status
  Container:  uf-sandbox (abc123def)
  Image:      quay.io/unbound-force/opencode-dev:latest
  Mode:       isolated
  Project:    /Users/dev/myproject
  Server:     http://localhost:4096
  Uptime:     2h 15m
```

**Output format** (when stopped):
```
No sandbox running.
```

---

## Internal Functions

### DetectPlatform

```go
// DetectPlatform detects the host platform configuration
// (architecture and SELinux status) for container flag
// selection.
func DetectPlatform(opts Options) PlatformConfig
```

**Behavior matrix**:
| Host | Arch | SELinux | Volume suffix |
|------|------|---------|--------------|
| macOS arm64 | arm64 | false | (none) |
| macOS amd64 | amd64 | false | (none) |
| Fedora amd64 | amd64 | true | `:Z` |
| Fedora amd64 | amd64 | false (disabled) | (none) |
| Ubuntu amd64 | amd64 | false | (none) |

### buildRunArgs

```go
// buildRunArgs constructs the podman run argument list from
// Options and PlatformConfig.
func buildRunArgs(opts Options, platform PlatformConfig) []string
```

**Returns**: Complete argument list for `podman run`,
including `-d`, `--name`, `-p`, `-v`, `-e`, `--memory`,
`--cpus`, and the image name. All values (including
user-provided `--image` and resource limits) are passed
as discrete `exec.Command` arguments — never shell-
interpolated — preventing command injection.

### waitForHealth

```go
// waitForHealth polls the OpenCode server health endpoint
// with exponential backoff until it responds or the timeout
// expires.
func waitForHealth(opts Options, timeout time.Duration) error
```

**Behavior**: Polls every 500ms initially, doubling up to
5s intervals. Total timeout per FR-005: 60 seconds.

### isContainerRunning

```go
// isContainerRunning checks if a container with the given
// name exists and is in the running state.
func isContainerRunning(opts Options) (bool, error)
```

**Implementation**: `podman inspect --format '{{.State.Running}}' uf-sandbox`
