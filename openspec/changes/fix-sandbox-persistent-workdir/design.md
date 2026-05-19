## Context

`PodmanBackend.Create()` copies the project into a named
volume via `podman cp <project-dir> container:/workspace/`.
Go's `filepath.Join(opts.ProjectDir, ".")` normalizes
away the trailing `.`, so `podman cp` copies the directory
itself (not its contents), placing the project at
`/workspace/<project-name>/`.

The ephemeral path (`buildRunArgs()` in config.go) sets
`--workdir /workspace/<basename>` and
`WORKSPACE=/workspace/<basename>` when parent mount is
active (fix from PR #123). The persistent path
(`buildPersistentRunArgs()` in podman.go) was never
updated with the same fix.

The container image's entrypoint reads
`WORKSPACE="${WORKSPACE:-/workspace}"` and runs
`cd "$WORKSPACE"` before starting `opencode serve`.
Without the env var, it defaults to `/workspace` — one
level above the project.

## Goals / Non-Goals

### Goals

- Set `--workdir` and `WORKSPACE` env var in
  `buildPersistentRunArgs()` to
  `/workspace/<project-basename>`
- OpenCode starts in the correct project directory
  after `uf sandbox create --backend podman`
- Add test coverage for the new args

### Non-Goals

- Fixing `filepath.Join` normalization of the trailing
  `.` in the `podman cp` source path — the copy
  behavior (directory into `/workspace/`) is acceptable
  as long as workdir is set correctly
- Adding `--mode` flag to `uf sandbox create` — the
  persistent path always uses a read-write named volume
- Changing DevPod backend behavior — DevPod manages its
  own working directory via devcontainer.json

## Decisions

### D1: Mirror the ephemeral fix pattern

Add the same `--workdir` and `WORKSPACE` pattern from
`buildRunArgs()` (config.go:244-250) to
`buildPersistentRunArgs()`. The persistent path always
copies the project directory (not contents) into
`/workspace/`, so the workdir is always
`/workspace/<basename>` — no `useParentMount()` guard
needed.

```go
projectSubdir := fmt.Sprintf("/workspace/%s",
    filepath.Base(opts.ProjectDir))
args = append(args, "--workdir", projectSubdir)
args = append(args, "-e",
    fmt.Sprintf("WORKSPACE=%s", projectSubdir))
```

### D2: No useParentMount guard needed

Unlike the ephemeral path which conditionally mounts
the parent directory, the persistent path always does
`podman cp <project-dir> container:/workspace/` which
creates `/workspace/<basename>/`. The workdir is
unconditionally `/workspace/<basename>`.

### D3: Existing containers unaffected

The `--workdir` and `WORKSPACE` env var are set at
container creation time (`podman run`). Existing
containers created before this fix retain their
original configuration. Users must `uf sandbox destroy`
and `uf sandbox create` to get the fix. This is
acceptable since the alternative (modifying running
containers) would be fragile.

### D4: PodmanBackend.Start() needs no changes

When a persistent workspace is resumed via
`podman start <container>`, the container retains its
original `--workdir` and env vars from creation. The
fix in `buildPersistentRunArgs()` (used only at create
time) is sufficient.

## Risks / Trade-offs

- **Risk**: Users with existing persistent workspaces
  will continue to see the wrong working directory
  until they destroy and recreate.
  **Mitigation**: This is a known limitation of
  container-level configuration. The fix applies to all
  newly created workspaces.

- **Trade-off**: The `podman cp` behavior (copying the
  directory itself vs contents) is a consequence of
  `filepath.Join` normalizing away the trailing `.`.
  We accept this behavior and set the workdir
  accordingly rather than changing the copy semantics,
  avoiding a behavioral change for existing users who
  may have scripts that depend on the current layout.
