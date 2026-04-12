# Data Model: Sandbox Command

**Branch**: `028-sandbox-command` | **Date**: 2026-04-12

## Core Types

### PlatformConfig

Detected host platform properties that influence container
configuration. Produced by `DetectPlatform()` in `detect.go`.

```go
// PlatformConfig captures host platform properties that
// influence container flags. Detected once at sandbox start
// and passed to configuration builders.
type PlatformConfig struct {
    // OS is the host operating system ("darwin" or "linux").
    OS string

    // Arch is the host CPU architecture ("arm64" or "amd64").
    Arch string

    // SELinux is true when SELinux is in enforcing mode.
    // Always false on macOS. When true, volume mounts
    // require the :Z relabeling flag.
    SELinux bool
}
```

**Relationships**: Consumed by `buildRunArgs()` in
`config.go` to determine volume mount flags and
`--platform` selection.

---

### Options

Configures a sandbox operation. All external dependencies
are injected as function fields for testability per
Constitution Principle IV.

```go
// Options configures sandbox operations. All external
// dependencies are injected as function fields for
// testability per Constitution Principle IV.
type Options struct {
    // ProjectDir is the project directory to mount into
    // the container. Defaults to current working directory.
    ProjectDir string

    // Mode is the mount mode: "isolated" (read-only with
    // overlay) or "direct" (read-write). Default: "isolated".
    Mode string

    // Detach skips auto-attach after container start.
    // When true, prints the server URL and exits.
    Detach bool

    // Image is the container image to use.
    // Default: "quay.io/unbound-force/opencode-dev:latest".
    // Overridden by UF_SANDBOX_IMAGE env var or --image flag.
    Image string

    // Memory is the container memory limit (e.g., "8g").
    Memory string

    // CPUs is the container CPU limit (e.g., "4").
    CPUs string

    // Stdout is the writer for user-facing output.
    Stdout io.Writer

    // Stderr is the writer for progress/status messages.
    Stderr io.Writer

    // LookPath finds a binary in PATH.
    LookPath func(string) (string, error)

    // ExecCmd runs a command and returns combined output.
    ExecCmd func(name string, args ...string) ([]byte, error)

    // ExecInteractive runs a command with stdin/stdout/stderr
    // connected to the terminal. Used for `opencode attach`
    // which requires interactive I/O.
    ExecInteractive func(name string, args ...string) error

    // Getenv reads an environment variable.
    Getenv func(string) string

    // ReadFile reads a file's contents.
    ReadFile func(string) ([]byte, error)

    // HTTPGet performs an HTTP GET request and returns the
    // status code. Used for health check polling.
    // Default: http.Get wrapper.
    HTTPGet func(url string) (int, error)
}
```

**Relationships**: Passed to all public functions in
`sandbox.go`. The `defaults()` method fills zero-value
fields with production implementations.

---

### ContainerStatus

Represents the current state of the sandbox container.
Returned by `Status()`.

```go
// ContainerStatus represents the current state of the
// sandbox container, parsed from `podman inspect` output.
type ContainerStatus struct {
    // Running is true when the container is active.
    Running bool

    // Name is the container name (always "uf-sandbox").
    Name string

    // ID is the container ID (short form).
    ID string

    // Image is the container image used.
    Image string

    // Mode is the mount mode ("isolated" or "direct").
    // Determined by inspecting the volume mount flags.
    Mode string

    // ProjectDir is the mounted project directory.
    ProjectDir string

    // ServerURL is the OpenCode server URL.
    ServerURL string

    // StartedAt is the container start time.
    StartedAt string

    // ExitCode is set when the container has stopped.
    // -1 when the container is running.
    ExitCode int
}
```

**Relationships**: Produced by `Status()`, consumed by
`start` (to detect already-running), `attach` (to get
server URL), `extract` (to verify running state).

---

### PatchSummary

Describes changes available for extraction from the
container. Returned by the extraction preview step.

```go
// PatchSummary describes changes available for extraction
// from the sandbox container.
type PatchSummary struct {
    // CommitCount is the number of commits since the
    // mount point (origin/HEAD).
    CommitCount int

    // FilesChanged is the number of files modified.
    FilesChanged int

    // Insertions is the total lines added.
    Insertions int

    // Deletions is the total lines removed.
    Deletions int

    // Patch is the raw patch content (format-patch output).
    Patch string

    // StatOutput is the human-readable diffstat.
    StatOutput string
}
```

**Relationships**: Produced by the preview step of
`Extract()`, displayed to the user for review before
applying.

---

## Constants

```go
const (
    // ContainerName is the fixed name for the sandbox
    // container. Only one sandbox is supported at a time.
    ContainerName = "uf-sandbox"

    // DefaultImage is the default container image.
    DefaultImage = "quay.io/unbound-force/opencode-dev:latest"

    // DefaultMemory is the default memory limit.
    DefaultMemory = "8g"

    // DefaultCPUs is the default CPU limit.
    DefaultCPUs = "4"

    // DefaultServerPort is the OpenCode server port.
    DefaultServerPort = 4096

    // HealthTimeout is the maximum time to wait for the
    // OpenCode server health check (FR-005).
    HealthTimeout = 60 * time.Second

    // ModeIsolated mounts the project directory read-only.
    ModeIsolated = "isolated"

    // ModeDirect mounts the project directory read-write.
    ModeDirect = "direct"
)
```

---

## Environment Variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `UF_SANDBOX_IMAGE` | Override container image | `quay.io/unbound-force/opencode-dev:latest` |
| `ANTHROPIC_API_KEY` | Forwarded to container | (from host) |
| `OPENAI_API_KEY` | Forwarded to container | (from host) |
| `GEMINI_API_KEY` | Forwarded to container | (from host) |
| `OPENROUTER_API_KEY` | Forwarded to container | (from host) |
| `OLLAMA_HOST` | Forwarded as `host.containers.internal:11434` | (overridden) |

---

## Type Relationships

```text
Options ──────────────────────┐
  │                           │
  ├─► DetectPlatform() ──► PlatformConfig
  │                           │
  ├─► buildRunArgs() ◄────────┘
  │       │
  │       ▼
  ├─► Start() ──► ContainerStatus
  │       │
  │       ├─► waitForHealth()
  │       └─► ExecInteractive("opencode", "attach", ...)
  │
  ├─► Stop()
  │
  ├─► Attach() ──► ExecInteractive("opencode", "attach", ...)
  │
  ├─► Status() ──► ContainerStatus
  │
  └─► Extract() ──► PatchSummary ──► git am
```
