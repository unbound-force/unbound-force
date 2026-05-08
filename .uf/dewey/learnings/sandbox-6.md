---
tag: sandbox
category: gotcha
created_at: 2026-05-03T19:30:07Z
identity: sandbox-6
tier: draft
---

When replacing a CDE backend (Che to DevPod), the isPersistentWorkspace() function in sandbox.go is a critical integration point that must be updated. It originally only checked for Podman named volumes, which meant DevPod workspaces were invisible to the persistent workspace detection path. The fix was to extend isPersistentWorkspace() to also call devpod status guarded by LookPath("devpod"). Without this, the gateway wiring for persistent Start() — a core goal of the change — would silently fail for DevPod users because Start() would fall through to the ephemeral path. The lesson: when adding a new Backend implementation, trace ALL callers of isPersistentWorkspace() and ensure the new backend's persistence model is detectable.
