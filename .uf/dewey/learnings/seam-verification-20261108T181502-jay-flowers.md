---
tag: seam-verification
author: jay-flowers
category: pattern
created_at: 2026-11-08T18:15:02Z
identity: seam-verification-20261108T181502-jay-flowers
tier: draft
---

A mocked-executor test that asserts argv verifies the command sequence (an implementation detail), never the external tool's behavioral contract. A new git subcommand introduced at a tool seam requires a Layer 3 round-trip test against a scratch repository (mkdtemp + git init, real executor): assert the tool envelope succeeds and the observable side effect (git index state) matches expectations, across the index-state matrix — pathspec unmatched (directory absent / untracked-only files) and pre-staged. Shipped precedent: preflight-declared-roundtrip.test.ts (self-labeled DE-007 Layer 3, real execFileSync + mkdtemp scratch project); scratch-smoke.test.ts covers Layer 1 registration on a scratch project. golden-output-shape.test.ts mocks git even on its real scratch directories, so passing Layer 2 does not substitute for Layer 3 — demonstrated by PR 993, where `git reset -q -- <pathspec>` with no matching index entries had an unpinned exit-status contract at a seam shipping to all scaffold consumers.
