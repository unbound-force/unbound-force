---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-05T19:53:40Z
identity: test-quality-20261105T195340-jay-flowers
tier: draft
---

When a test wraps an injected command/DI seam (e.g. ExecCmdInDir) to record invoked commands and then asserts properties of the recorded commands (flag presence, argument shape), it must guard against the vacuous pass: if the code path under test is skipped or refactored away, the recorder stays empty and assertions like "every recorded command contains --flag" pass trivially over zero elements. Fail closed first — precede the property assertions with a non-empty check (t.Fatal when len(recorded) == 0) so a silent skip surfaces as a test failure.
