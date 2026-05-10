---
tag: sandbox
author: jay-flowers
category: pattern
created_at: 2026-05-09T22:09:54Z
identity: sandbox-20260509T220954-jay-flowers
tier: draft
---

The DevPod --ide flag controls what IDE DevPod launches after workspace provisioning. It does not affect the OpenCode TUI server which runs independently on port 4096. Supported values are: none, vscode, openvscode, fleet, jupyternotebook, cursor. The IDE flag only applies to the DevPod backend — the ephemeral Podman path silently ignores it. The resolution chain follows the same pattern as other sandbox options: CLI flag > UF_SANDBOX_IDE env var > .uf/config.yaml sandbox.ide field > default "none".
