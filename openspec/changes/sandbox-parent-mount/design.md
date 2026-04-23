## Context

The sandbox container currently mounts only the project
directory at `/workspace`. Dewey's `sources.yaml` uses
relative paths like `../dewey` to index sibling repos,
but these paths don't resolve inside the container
because the parent directory isn't mounted. This means
Dewey runs in degraded mode inside the sandbox — only
the local project is indexed, losing cross-repo semantic
search.

## Goals / Non-Goals

### Goals
- Dewey indexes sibling repos inside the sandbox
- Generic solution — works for any project layout, not
  just unbound-force
- Backward compatible via `--no-parent` opt-out flag
- No changes to Dewey config or `opencode.json`

### Non-Goals
- Changing Dewey's indexing behavior or sources.yaml
- Supporting arbitrary multi-directory mounts
- Mounting directories that aren't the immediate parent
- Changing how `uf sandbox extract` works (patches are
  still generated from the project subdirectory)

## Decisions

**D1: Mount parent at `/workspace`, set `--workdir`**

Mount `filepath.Dir(ProjectDir)` at `/workspace` and
add `--workdir /workspace/<basename>` to the podman
args. This preserves the convention that OpenCode's
CWD is the project root, while making the parent
accessible at `/workspace`.

Rationale: This requires zero changes to Dewey,
OpenCode, or any tool config. Relative paths like
`../dewey` resolve correctly because the parent is
the mount root. The `--workdir` flag ensures OpenCode
starts in the correct project subdirectory.

**D2: Default on, `--no-parent` to opt out**

The parent mount is the default behavior. Users who
don't want sibling access (e.g., security-sensitive
environments) can pass `--no-parent` to get the
current project-only mount.

Rationale: The common case benefits from cross-repo
indexing. The flag name `--no-parent` clearly
communicates what it disables. This follows the Go
convention of `--no-*` flags for disabling defaults
(e.g., `--no-verify` in git).

**D3: Fall back to project-only for root-level projects**

If `filepath.Dir(ProjectDir)` returns `/`, fall back
to project-only mount (same as `--no-parent`). Log a
debug message explaining why.

Rationale: Mounting `/` as a container volume is
dangerous and likely unintentional. This edge case
is rare but should fail safely.

**D4: Same read/write mode for parent mount**

The parent mount uses the same mode as the project
mount — read-write in direct mode, read-only in
isolated mode. No separate mode for sibling repos.

Rationale: Simplicity. The mode flag controls the
entire workspace, not individual subdirectories.
Sibling repos being writable in direct mode is an
acceptable trade-off given the development context.

**D5: `NoParent` field on `Options` struct**

Add a `NoParent bool` field to the `Options` struct,
consistent with existing boolean fields (`Detach`,
`Yes`, `Force`). Wired from the `--no-parent` CLI flag.

Rationale: Follows the established injectable options
pattern (Constitution Principle IV — Testability).
The field is testable without CLI parsing.

## Risks / Trade-offs

- **Risk**: Sibling repos are writable in direct mode.
  A bug in a tool could modify files in a sibling repo.
  **Mitigation**: Direct mode already trusts the
  container with write access. Users who want isolation
  should use isolated mode (read-only) or `--no-parent`.

- **Risk**: Parent directory may contain many large
  repos, increasing the container's visible filesystem.
  **Mitigation**: Podman volume mounts are bind mounts —
  they don't copy data. There is no performance or
  storage impact from mounting a larger directory.

- **Trade-off**: The `/workspace` path now refers to
  the parent directory, not the project. Tools that
  assume `/workspace` is the project root will break.
  **Mitigation**: The `--workdir` flag sets the CWD
  correctly. OpenCode and all tools that use CWD (not
  hardcoded `/workspace`) work correctly. The sandbox
  `extract` command uses CWD for git operations.
