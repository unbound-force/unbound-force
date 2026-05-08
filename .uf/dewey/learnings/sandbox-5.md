---
tag: sandbox
category: pattern
created_at: 2026-04-28T19:01:55Z
identity: sandbox-5
tier: draft
---

The busybox:latest image with --entrypoint stat override is the correct approach for lightweight Podman probes that need to check file ownership or mount behavior. Using the full dev image (quay.io/unbound-force/opencode-dev:latest) for a simple stat probe is wrong because: (1) it pulls a ~1GB image when only ~5MB is needed, (2) the image's default entrypoint may execute services or bind ports, (3) the mount is :ro but the entrypoint could still read sensitive project files. The --entrypoint override prevents any image-defined entrypoint from running. This pattern was identified by the Adversary reviewer as a HIGH security finding and should be followed for any future probe containers.
