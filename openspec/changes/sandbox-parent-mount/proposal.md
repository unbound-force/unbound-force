## Why

When the sandbox container mounts only the project
directory at `/workspace`, tools like Dewey that use
relative paths (`../dewey`, `../gaze`) to index sibling
repositories cannot access them. Dewey's `sources.yaml`
defines disk sources as relative paths from the project
root (e.g., `path: "../dewey"`), but inside the
container these resolve to `/dewey` which does not exist.

Mounting the project's parent directory instead gives
the container access to the full workspace, allowing
Dewey and any other tool to traverse sibling directories
naturally.

## What Changes

### Volume mount target

Currently: `ProjectDir` → `/workspace`
```
-v /Users/j/Projects/unbound-force/unbound-force:/workspace
```

After: `filepath.Dir(ProjectDir)` → `/workspace`
with `--workdir /workspace/<basename>`
```
-v /Users/j/Projects/unbound-force:/workspace
--workdir /workspace/unbound-force
```

### New CLI flag

Add `--no-parent` flag to `uf sandbox start` that
disables parent mounting and uses the current behavior
(project-only mount). Default behavior is parent mount.

### Edge case handling

If `filepath.Dir(ProjectDir)` returns `/` (project is
at filesystem root), fall back to project-only mount
automatically and log a warning.

## Capabilities

### New Capabilities
- `Parent directory mount`: The sandbox mounts the
  project's parent directory at `/workspace` by
  default, giving container tools access to sibling
  directories via relative paths.
- `--no-parent flag`: Opt-out flag on `uf sandbox start`
  to disable parent mounting and use project-only mount
  (current behavior).

### Modified Capabilities
- `buildVolumeMounts()`: Mounts parent directory
  instead of project directory when `--no-parent` is
  not set.
- `buildRunArgs()`: Adds `--workdir` to set the
  container's working directory to the project
  subdirectory within the parent mount.

### Removed Capabilities
- None.

## Impact

- `internal/sandbox/config.go`: Modified
  `buildVolumeMounts()` and `buildRunArgs()` to support
  parent mount with workdir.
- `internal/sandbox/sandbox.go`: Add `NoParent bool`
  field to `Options` struct.
- `cmd/unbound-force/sandbox.go`: Register `--no-parent`
  flag on the `start` subcommand.
- `internal/sandbox/sandbox_test.go`: Update mount
  assertions, add `--no-parent` tests, add edge case
  tests.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

No inter-hero communication is affected. The mount
change is purely a container configuration concern —
no artifacts or protocols change.

### II. Composability First

**Assessment**: PASS

The `--no-parent` flag preserves backward compatibility.
Users who don't want parent mounting can opt out. The
sandbox remains independently usable with either mount
mode.

### III. Observable Quality

**Assessment**: PASS

No output formats change. The sandbox status output
continues to show the project directory. The mount
target is an internal implementation detail.

### IV. Testability

**Assessment**: PASS

`buildVolumeMounts()` and `buildRunArgs()` remain
testable in isolation via injected `Options` — no
external services needed. The `NoParent` field is a
simple boolean that changes the mount path
calculation. All edge cases (root directory, SELinux,
isolated vs direct mode) are testable via unit tests.
