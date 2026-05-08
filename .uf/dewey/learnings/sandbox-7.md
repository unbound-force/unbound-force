---
tag: sandbox
category: pattern
created_at: 2026-05-03T19:30:12Z
identity: sandbox-7
tier: draft
---

The spec review council identified that autoDetectBackend() silently preferring DevPod when devpod is in PATH would be a behavioral break for existing users. All 5 reviewers independently flagged this. The resolution was to require BOTH devpod in PATH AND .devcontainer/devcontainer.json existence before auto-selecting DevPod. This means users must explicitly opt in via uf sandbox init before DevPod becomes the default. This pattern of requiring both binary availability AND project-level config for auto-detection should be applied to any future backend additions to prevent silent behavioral changes.
