---
tag: devpod
category: gotcha
created_at: 2026-05-08T18:21:29Z
identity: devpod-1
tier: draft
---

When adding tools to `uf setup` and `uf doctor`, the DevPod ecosystem has changed significantly: the standalone `podman` provider no longer exists. Users must alias the Docker provider via `devpod provider add docker --name podman -o DOCKER_COMMAND=podman`. This command uses DOCKER_COMMAND (not DOCKER_HOST) because the socket path varies by UID and platform. The provider detection must use exact first-column name matching on `devpod provider list` output, not substring matching, to avoid false positives from providers like "podman-custom".
