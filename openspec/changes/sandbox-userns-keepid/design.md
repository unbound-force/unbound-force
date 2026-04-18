## Context

Podman rootless containers run in a user namespace.
Bind mounts expose host files with their original
UID/GID ownership. If the container's user has a
different UID than the host user, write operations
fail with "permission denied" — even on read-write
mounts.

Podman provides `--userns=keep-id` which maps the
host user's UID/GID into the container, so the
container process runs as the same effective user.
This is the standard solution for rootless Podman
bind mount permission issues.

## Goals / Non-Goals

### Goals

- Add `--userns=keep-id` to `podman run` arguments
  on Linux
- Fix gcloud token refresh write failures
- Fix project directory write failures in direct mode

### Non-Goals

- No changes for macOS (Podman VM handles UID mapping)
- No Docker support (Docker doesn't support
  `--userns=keep-id`)
- No changes to isolated mode (read-only mount doesn't
  need write permission)

## Decisions

### D1: Linux-only, unconditional

Add `--userns=keep-id` on all Linux platforms, not
just when specific mounts are present. The flag is
harmless when not needed and prevents subtle permission
bugs across all bind mounts (project dir, gcloud dir,
opencode auth dir).

On macOS, skip the flag — Podman's VM layer handles
UID mapping transparently.

### D2: Use PlatformConfig.OS

The `DetectPlatform()` function already populates
`PlatformConfig.OS` from `runtime.GOOS`. Use this
to gate the flag — same pattern as the SELinux `:Z`
flag detection.

### D3: Add after volume mounts, before command

The `--userns=keep-id` flag is a container-level
option, not a volume-specific option. It goes in the
`podman run` arguments alongside `--name`, `--memory`,
etc. — not attached to a specific `-v` mount.
