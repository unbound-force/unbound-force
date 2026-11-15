---
tag: resilience
author: jay-flowers
category: gotcha
created_at: 2026-11-15T01:15:54Z
identity: resilience-20261115T011554-jay-flowers
tier: draft
as_of: "unbound-force@7f44cba6be6f30f6c3cec406daab342434cc1474"
---

Node.js process.env coerces undefined to the literal string "undefined" on assignment (process.env.TMPDIR = undefined). On Linux where TMPDIR is unset by default, os.tmpdir() will then return the relative path "undefined", causing path.join(os.tmpdir(),"opencode") to write into <cwd>/.opencode/undefined/ and pollute the worktree. Fix is to delete the var (delete process.env.TMPDIR) when restoring an originally-unset value, and assert the env is back to undefined after the test.
