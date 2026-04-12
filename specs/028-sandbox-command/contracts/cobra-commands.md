# Contract: Cobra Command Registration

**File**: `cmd/unbound-force/sandbox.go`
**Date**: 2026-04-12

## Command Tree

```text
uf sandbox
├── start    [--mode isolated|direct] [--detach] [--image <img>]
│            [--memory <mem>] [--cpus <n>]
├── stop
├── attach
├── extract  [--yes]
└── status
```

## Registration

```go
// newSandboxCmd returns the `uf sandbox` parent command with
// all subcommands registered.
func newSandboxCmd() *cobra.Command
```

Called from `main()` via `root.AddCommand(newSandboxCmd())`.

---

## Subcommand Specifications

### uf sandbox start

```
Usage: uf sandbox start [flags]

Flags:
  --mode string     Mount mode: isolated (read-only) or direct (read-write)
                    (default "isolated")
  --detach          Start container without attaching TUI
  --image string    Container image (default from UF_SANDBOX_IMAGE or
                    quay.io/unbound-force/opencode-dev:latest)
  --memory string   Container memory limit (default "8g")
  --cpus string     Container CPU limit (default "4")
```

**Delegates to**: `runSandboxStart(sandboxStartParams)`

```go
type sandboxStartParams struct {
    projectDir string
    mode       string
    detach     bool
    image      string
    memory     string
    cpus       string
    stdout     io.Writer
    stderr     io.Writer
}
```

### uf sandbox stop

```
Usage: uf sandbox stop
```

**Delegates to**: `runSandboxStop(sandboxStopParams)`

```go
type sandboxStopParams struct {
    stdout io.Writer
}
```

### uf sandbox attach

```
Usage: uf sandbox attach
```

**Delegates to**: `runSandboxAttach(sandboxAttachParams)`

```go
type sandboxAttachParams struct {
    stdout io.Writer
}
```

### uf sandbox extract

```
Usage: uf sandbox extract [flags]

Flags:
  --yes    Skip confirmation prompt
```

**Delegates to**: `runSandboxExtract(sandboxExtractParams)`

```go
type sandboxExtractParams struct {
    yes    bool
    stdout io.Writer
    stderr io.Writer
    stdin  io.Reader
}
```

### uf sandbox status

```
Usage: uf sandbox status
```

**Delegates to**: `runSandboxStatus(sandboxStatusParams)`

```go
type sandboxStatusParams struct {
    stdout io.Writer
}
```

---

## Flag Precedence

For `--image`:
1. `--image` flag (highest)
2. `UF_SANDBOX_IMAGE` environment variable
3. `DefaultImage` constant (lowest)

For `--memory` and `--cpus`:
1. Flag value (highest)
2. Default constant (lowest)

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error (Podman missing, container failure, etc.) |
