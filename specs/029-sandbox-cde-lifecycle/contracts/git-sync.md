# Contract: Bidirectional Git Sync

**Package**: `internal/sandbox`
**Date**: 2026-04-13

## Overview

Git sync is the coordination mechanism between the
engineer's host and the CDE workspace. It replaces the
`uf sandbox extract` workflow for persistent workspaces.

## Workspace Git Setup (during Create)

When `Create()` provisions a workspace, the git
configuration is set up for bidirectional sync:

```go
// setupGitSync configures the workspace's git remote
// and branch for bidirectional sync.
//
// Steps:
// 1. Detect the current branch on the host
// 2. Configure the same remote origin in the workspace
// 3. Create a workspace branch (optional, for isolation)
// 4. Set up credential forwarding (Podman: follows
//    Spec 028 env var pattern; CDE: K8s secrets per
//    FR-014)
func setupGitSync(opts Options) error
```

### Podman Backend

For Podman persistent workspaces, git sync works via
the shared named volume:

1. **Initial setup**: `podman cp` copies the project
   source (including `.git/`) into the named volume
2. **Agent pushes**: Agent commits and pushes from
   inside the container via `podman exec git push`
3. **Engineer pulls**: Engineer runs `git pull` on the
   host to get agent changes
4. **Engineer pushes**: Engineer pushes from host;
   workspace pulls on next `/unleash` run

### CDE Backend

For CDE workspaces, git sync works via the Che
workspace's built-in git support:

1. **Initial setup**: Che clones the repo from the
   devfile's `projects` section
2. **Agent pushes**: Agent commits and pushes using
   Che's credential management
3. **Engineer pulls**: Engineer runs `git pull` on the
   host
4. **Engineer pushes**: Engineer pushes from host or
   Che IDE; workspace pulls automatically or on next
   `/unleash` run

## Conflict Detection

```go
// checkGitSync verifies the workspace's git state is
// clean and up-to-date with the remote. Called at the
// start of each `/unleash` run.
//
// Returns:
// - nil if workspace is clean and up-to-date
// - error with conflict details if merge is needed
func checkGitSync(opts Options) error
```

**Conflict handling**:

| Scenario | Behavior |
|----------|----------|
| Workspace clean, remote unchanged | Continue |
| Workspace clean, remote has new commits | Auto-pull (fast-forward) |
| Workspace has uncommitted changes | Warn: "uncommitted changes in workspace" |
| Workspace and remote diverged | Error: "merge conflict — resolve before continuing" |

## Extract Compatibility

The existing `uf sandbox extract` command continues to
work for both ephemeral and persistent workspaces:

- **Ephemeral mode**: Uses `git format-patch` / `git am`
  (Spec 028 behavior, unchanged)
- **Persistent mode (Podman)**: Uses `git format-patch` /
  `git am` (same mechanism, different container name)
- **Persistent mode (CDE)**: Suggests using `git pull`
  instead of extract, since the workspace has push access

```go
// Extract behavior for persistent workspaces:
//
// If the workspace has push access (CDE or configured
// Podman), suggest git pull instead:
//   "This workspace has git push access. Use
//    `git pull` on the host instead of extract."
//
// If the workspace does not have push access, fall back
// to the Spec 028 format-patch/am workflow.
```
