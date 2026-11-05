---
tag: untrusted-output-rendering
author: jay-flowers
category: gotcha
created_at: 2026-11-03T19:14:51Z
identity: untrusted-output-rendering-20261103T191451-jay-flowers
tier: draft
---

Fence-safe markdown rendering of untrusted tool output is only half structural. A fence of max(3, longestBacktickRun + 1) backticks with a `text` info string and a newline-terminated body provably keeps output inside its CommonMark block (a closing fence must be at least as long as the opening one, so no output line can end it), but the surrounding LABEL — e.g. `#### <name> — <status> (exit N)` — is injection-safe only because of non-local invariants. Verified in unbound-force `.opencode/lib/preflight-declared.ts` renderer: the label template interpolates only enumerated values (tool names, exit codes), never free text.
