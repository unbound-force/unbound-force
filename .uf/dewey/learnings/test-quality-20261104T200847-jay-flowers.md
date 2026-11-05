---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-04T20:08:47Z
identity: test-quality-20261104T200847-jay-flowers
tier: draft
---

Exact-command string fakes pin external CLI contracts in unit tests without running the real tool: injected ExecCmdInDir/ExecCmd fakes switch on the full command string (e.g. the complete "npm install --ignore-scripts --omit=dev --no-package-lock --no-audit --no-fund") and the test asserts the recorded event sequence. Used in internal/scaffold/plugin_activation_test.go, this makes flag-level requirements (like --no-package-lock, which the spec mandates) regression-protected — removing or altering any flag f
