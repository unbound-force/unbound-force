---
tag: instruction-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-08T19:49:52Z
identity: instruction-quality-20261108T194952-jay-flowers
tier: draft
---

Replacing a command inside a fenced bash block while leaving the block's original closing fence in place produces an orphaned ``` that opens a spurious code fence and silently absorbs all prose up to the next bare fence. In PR 959's divisor-linguist.md, the orphan at line 110 swallowed lines 112-129 — the wrapper-only mandate, the exit-code contract, and the JSON example — so rendered views restructured the file's most load-bearing section while raw-text readers and the instruction linter saw no error (no fence-parity rule exists). After editing inside fenced blocks in instruction files, verify fence parity: count opening and closing fences in the edited region and confirm no stray bare fence remains.
