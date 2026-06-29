---
tag: sandbox
author: jay-flowers
category: gotcha
created_at: 2026-05-13T18:50:48Z
identity: sandbox-20260513T185048-jay-flowers
tier: draft
---

The persistent Podman sandbox path (buildPersistentRunArgs in podman.go) was missing --workdir and WORKSPACE env var, causing OpenCode to start in /workspace instead of /workspace/project-name. This is the same class of bug as PR #123 which fixed the ephemeral path (buildRunArgs in config.go). When podman cp copies a directory into a volume, Go's filepath.Join normalizes away trailing dots so the copy creates /workspace/project-name/ (directory itself, not contents). Both --workdir and WORKSPACE must be set to /workspace/basename so the entrypoint's cd "$WORKSPACE" lands in the correct directory. The lesson: when fixing a bug in one code path (ephemeral), always check if the same pattern exists in parallel code paths (persistent) that may have been added independently.
