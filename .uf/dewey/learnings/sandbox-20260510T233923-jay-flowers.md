---
tag: sandbox
author: jay-flowers
category: pattern
created_at: 2026-05-10T23:39:23Z
identity: sandbox-20260510T233923-jay-flowers
tier: draft
---

The devcontainer.json spec's runArgs field passes extra arguments directly to the underlying docker run or podman run command. For DevPod workspaces using the Docker provider aliased to Podman (via DOCKER_COMMAND=podman), adding "runArgs": ["--userns=keep-id:uid=1000,gid=1000"] to devcontainer.json maps the host user to UID 1000 (dev) inside the container. This eliminates the root:nobody file ownership problem without needing chown, sudo, custom providers, or containerUser overrides. The runArgs approach reuses the exact same flag proven in the Podman backend's uidMappingArgs() function. Caveat: on macOS, the Podman machine's virtiofs must support --userns=keep-id, which most modern configurations do. If it doesn't, the fallback is to remove runArgs and add a postStartCommand with sudo chown.
