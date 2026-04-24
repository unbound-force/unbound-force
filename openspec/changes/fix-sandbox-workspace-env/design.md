## Context

The container image's `entrypoint.sh` runs
`cd "$WORKSPACE"` early in startup, where `WORKSPACE`
defaults to `/workspace`. PR #119 set `--workdir` on
the container, but the entrypoint overrides it before
OpenCode starts. The `WORKSPACE` env var is already
read by the entrypoint — we just need to pass it.

## Goals / Non-Goals

### Goals
- OpenCode CWD matches the project directory when
  parent mount is active
- No container image changes required
- Backward compatible with `--no-parent` mode

### Non-Goals
- Modifying the container image or entrypoint script
- Changing mount behavior (already correct from PR #119)
- Supporting custom WORKSPACE paths

## Decisions

**D1: Pass WORKSPACE env var, not change --workdir**

The `--workdir` flag is correctly set but the
entrypoint overrides it. Rather than fighting the
entrypoint, work with it — pass `WORKSPACE` so the
`cd "$WORKSPACE"` line goes to the right place. This
requires zero image changes.

**D2: Only set WORKSPACE when parent mount is active**

When `--no-parent` is used (project-only mount), the
entrypoint's default `WORKSPACE=/workspace` is
correct — `/workspace` IS the project directory. Only
set `WORKSPACE` explicitly when parent mount changes
the meaning of `/workspace` to be the parent.

## Risks / Trade-offs

- **Risk**: Future entrypoint changes could remove the
  `WORKSPACE` variable. **Mitigation**: The entrypoint
  is in our own container image (`containerfile` repo)
  and the variable is documented. Also, `--workdir`
  remains set as a belt-and-suspenders fallback.
