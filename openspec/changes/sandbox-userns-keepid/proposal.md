## Why

When `uf sandbox start --mode direct` runs on Linux,
bind-mounted volumes preserve the host filesystem's
ownership. If the host user (e.g., UID 1000) differs
from the container's `dev` user (e.g., UID 1001),
the container process cannot write to host-owned files
— even though the mount is configured as read-write.

This breaks two critical paths:

1. **Google Cloud auth**: The gcloud config directory
   (`~/.config/gcloud/`) is mounted read-write so the
   auth library can refresh OAuth2 tokens in
   `access_tokens.db`. If the UIDs don't match, token
   refresh fails silently and authentication stops
   working after ~1 hour.

2. **Project directory in direct mode**: The project
   directory is mounted read-write in `--mode direct`.
   UID mismatch means the agent inside the container
   cannot write files to the host project.

On macOS, Podman's VM layer handles UID mapping
transparently, so the issue only manifests on Linux
(Fedora, RHEL, Ubuntu).

## What Changes

### Modified Capabilities

- `buildRunArgs()` in `internal/sandbox/config.go`:
  Add `--userns=keep-id` flag on Linux to map the host
  user's UID/GID into the container's user namespace.

### New Capabilities

None.

### Removed Capabilities

None.

## Impact

- 1 file modified: `internal/sandbox/config.go`
  (~3 lines added to `buildRunArgs()`)
- 1 test file updated: `internal/sandbox/sandbox_test.go`
- No new packages or dependencies
- `--userns=keep-id` is Podman-native (not available
  in Docker — but Docker is not supported)

## Constitution Alignment

All N/A — this is a container runtime flag addition
that fixes a Linux-specific permission issue.
