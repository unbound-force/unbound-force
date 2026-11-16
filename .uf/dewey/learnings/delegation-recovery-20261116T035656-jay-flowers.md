---
tag: delegation-recovery
author: jay-flowers
category: gotcha
created_at: 2026-11-16T03:56:56Z
identity: delegation-recovery-20261116T035656-jay-flowers
tier: draft
---

Delegation wall-clock timeouts with retryable:true are recoverable without re-implementation: when uf_delegate_command times out while the child is still active (recent progress, running a tool — typically the verification test run), the child's file edits persist in the shared working tree. Re-delegate once with a resume-aware prompt — inspect the existing diff with git diff, verify it against the original requirements, run the exact verification commands, and report, explicitly instructing "do not rewrite from scratch if the change is correct" — and the retry's success records the implement checkpoint through the proper success channel. Verified on branch quick/pin-the-triage-issue-posting-choreograph (2026-11-15, /uf.unleash quick-tier implement for issue #1049): the first delegation timed out at 900s mid-verification with a complete 85-line test in the tree; the resume delegation confirmed it correct as-is and cost only ~2.5k tokens.
