---
tag: sandbox
category: decision
created_at: 2026-05-03T19:30:25Z
identity: sandbox-8
tier: draft
---

When integrating an external CLI tool as a Backend via subprocess (DevPod, Podman, chectl), the env var injection mechanism is a security boundary that must be specified at design time, not discovered during implementation. For DevPod, the correct flag is --workspace-env KEY=VALUE which sets env vars in the workspace container. The alternative --dotfiles-env only applies during dotfile installation and would NOT make vars available to OpenCode at runtime. The spec review caught that the design deferred this decision with "will be determined during implementation" — this was a HIGH finding because the credential injection path must be documented before code is written.
