# Contract: Cobra Commands

**Package**: `cmd/unbound-force`
**Date**: 2026-04-13

## New Subcommands

### uf sandbox create

```
Usage: uf sandbox create [flags]

Provision a persistent sandbox workspace for the current
project. Uses Eclipse Che/Dev Spaces when configured,
Podman with named volumes otherwise.

Flags:
  --backend string   Backend: auto, podman, or che
                     (default "auto")
  --image string     Container image (Podman only;
                     default from UF_SANDBOX_IMAGE or
                     quay.io/unbound-force/opencode-dev:latest)
  --memory string    Memory limit (default "8g")
  --cpus string      CPU limit (default "4")
  --name string      Workspace name override
                     (default "uf-sandbox-<project-name>")
  --detach           Start without attaching TUI
  --demo-ports ints  Additional ports to expose for demos
                     (comma-separated, e.g., 3000,8080)
```

**Params struct**:
```go
type sandboxCreateParams struct {
    projectDir  string
    backend     string
    image       string
    memory      string
    cpus        string
    name        string
    detach      bool
    demoPorts   []int
    stdout      io.Writer
    stderr      io.Writer
}
```

**Delegation**: `runSandboxCreate(p) → ResolveBackend()
→ backend.Create(opts)`

---

### uf sandbox destroy

```
Usage: uf sandbox destroy [flags]

Permanently delete the sandbox workspace and all
associated state (named volumes, CDE workspace).

Flags:
  --yes    Skip confirmation prompt
  --force  Force destroy even if workspace is running
```

**Params struct**:
```go
type sandboxDestroyParams struct {
    projectDir string
    yes        bool
    force      bool
    stdout     io.Writer
    stderr     io.Writer
    stdin      io.Reader
}
```

**Delegation**: `runSandboxDestroy(p) → ResolveBackend()
→ backend.Destroy(opts)`

**Confirmation prompt** (unless `--yes`):
```
Destroy sandbox "uf-sandbox-myproject"?
This will permanently delete all workspace state.
[y/N]
```

---

## Updated Subcommands

### uf sandbox start (updated)

**New behavior**: When a persistent workspace exists
(named volume or CDE workspace), `start` resumes it
instead of creating an ephemeral container. When no
persistent workspace exists, falls back to Spec 028
ephemeral behavior.

**New flags**:
```
  --backend string   Backend override (default "auto")
```

**Detection logic**:
1. Check for named volume `uf-sandbox-<project-name>`
2. Check for CDE workspace `uf-<project-name>`
3. If either exists → persistent start
4. If neither → ephemeral start (Spec 028)

---

### uf sandbox stop (updated)

**New behavior**: When a persistent workspace is running,
`stop` stops the container but preserves the named
volume. When an ephemeral container is running, `stop`
removes the container (Spec 028 behavior).

---

### uf sandbox status (updated)

**New output format** (persistent workspace):
```
Sandbox Status
  Workspace:  uf-sandbox-myproject
  Backend:    podman (persistent)
  Image:      quay.io/unbound-force/opencode-dev:latest
  State:      running
  Project:    /Users/dev/myproject
  Server:     http://localhost:4096
  Demo:       http://localhost:3000 (demo-web)
              http://localhost:8080 (demo-api)
  Started:    2026-04-13T10:00:00Z
```

**New output format** (CDE workspace):
```
Sandbox Status
  Workspace:  uf-myproject
  Backend:    che (persistent)
  Che URL:    https://che.example.com
  State:      running
  Project:    https://github.com/org/myproject
  Server:     https://uf-myproject-opencode.apps.che.example.com
  Demo:       https://uf-myproject-demo-web.apps.che.example.com
              https://uf-myproject-demo-api.apps.che.example.com
  Started:    2026-04-13T10:00:00Z
```

---

## Updated Parent Command

```
Usage: uf sandbox [command]

Manage containerized OpenCode sessions.

Subcommands:
  create   Provision a persistent sandbox workspace
  destroy  Permanently delete a sandbox workspace
  start    Launch or resume a sandbox
  stop     Stop a sandbox (preserves persistent state)
  attach   Connect to a running sandbox's TUI
  extract  Extract changes from the sandbox as git patches
  status   Show sandbox workspace status
```

---

## Command Registration

```go
func newSandboxCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "sandbox",
        Short: "Manage containerized OpenCode sessions",
        // ...
    }

    cmd.AddCommand(
        newSandboxCreateCmd(),   // NEW
        newSandboxDestroyCmd(),  // NEW
        newSandboxStartCmd(),    // UPDATED
        newSandboxStopCmd(),     // UPDATED
        newSandboxAttachCmd(),   // unchanged
        newSandboxExtractCmd(),  // unchanged
        newSandboxStatusCmd(),   // UPDATED
    )

    return cmd
}
```
