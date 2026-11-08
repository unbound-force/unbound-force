---
tag: spec-enforcement
author: jay-flowers
category: pattern
created_at: 2026-11-05T14:37:19Z
identity: spec-enforcement-20261105T143719-jay-flowers
tier: draft
---

A .gitignore entry added for a file whose presence violates a spec MUST (stop-shipping-plugin-lock, PR 957: .opencode/package-lock.json, SC-FR-003 "MUST NOT track or ship") silently deletes the last passive signal of a violation — git status no longer shows the stray file. Pair every such ignore entry with an automated existence assertion in the recipe or path that could create the file (e.g., @test ! -e .opencode/package-lock.json in the make plugin-test recipe), so dropping the governing flag (--no-package-lock) fails CI loudly instead of regressing silently. A one-time manual task verification (tasks.md 4.2) does not survive flag drift; Constitution Principle IV requires quality claims to carry automated, reproducible evidence. Caveat: the assertion checks absence, not authorship — opencode is a documented second writer of the lock path, so local runs can fail on a foreign-written, spec-tolerated lock; CI fresh checkouts cannot.
