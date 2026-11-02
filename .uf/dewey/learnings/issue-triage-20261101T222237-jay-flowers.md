---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:22:37Z
identity: issue-triage-20261101T222237-jay-flowers
tier: draft
---

Triage verification for follow-up issues should distinguish citation validity from artifact presence in the current checkout. Issue #869 (wire uf gate as FullSend validation_loop.script) cites parent #514's acceptance criterion, delivering PR #582, and openspec/changes/uf-gate/ design.md Non-Goals — full traceability — yet none of the cited artifacts (uf gate command, internal/gate package, the uf-gate change directory) exist in the reviewing checkout because the delivering PR is unmerged there. Such issues remain VALID sequenced work; the correct triage output is a dependency note (gate implementation on the delivering PR merging), not a verdict downgrade. Complementary pattern: optional scope items that require a different spec tier than the parent change (e.g., CI coverage ratchets, which are governance-gated under the Gatekeeping rule) should be split into their own issues rather than carried as optional scope inside the parent issue.
