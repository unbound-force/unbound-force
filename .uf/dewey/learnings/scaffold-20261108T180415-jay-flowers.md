---
tag: scaffold
author: jay-flowers
category: pattern
created_at: 2026-11-08T18:04:15Z
identity: scaffold-20261108T180415-jay-flowers
tier: draft
---

Scaffolded command prose ships twice: every .opencode/commands/*.md edit must land byte-identical in internal/scaffold/assets/opencode/commands/, and the drift tests in internal/scaffold/scaffold_test.go lock the pair together. A determinism violation in a scaffolded command therefore doubles its surface and is frozen by the parity tests — evaluate new command instruction for DE-001 tool extraction before mirroring, not after.
