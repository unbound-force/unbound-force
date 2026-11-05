---
tag: test-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-05T20:38:44Z
identity: test-quality-20261105T203844-jay-flowers
tier: draft
---

In command-contract tests (internal/scaffold/command_contract_test.go), presence-only clauses cannot pin ordering invariants. When a command fix's essence is ordering — e.g., uf.address-feedback Section 1.2.1 must precede the first **CHECKPOINT** marker because progress state is keyed to the checked-out branch (#954) — add a bespoke strings.Index comparison on normalizeCommandText output after assertCommandCopies, iterating loadCommandCopies(..., true) so canonical and scaffold-mirror copies are both pinned. Anchor choice is load-bearing: match the bold marker form ('**CHECKPOINT**') rather than the bare word, and do not anchor on the first get_progress mention — both 'checkpoint' and 'get_progress' legitimately appear in prose (SESSION-RESUME GUARD, Progress Tracking, TodoWrite resume) before Section 1.2.1, so bare-word or get_progress anchors fail against current head.
