---
tag: address-feedback
author: jay-flowers
category: pattern
created_at: 2026-11-05T16:21:16Z
identity: address-feedback-20261105T162116-jay-flowers
tier: draft
---

Feedback-triage decision rule (from PR 957 item 12, stop-shipping-plugin-lock): a council finding that self-identifies as an accepted-risk documentation gap maps to AUTHOR-DECIDES, not ACCEPT, when three conditions hold. (1) No MUST rule binds the gap: Constitution Principle V's content-hash requirement applies only to downloads outside a package manager's built-in verification — `npx -y npm@<version>` resolves inside npm's own registry-integrity checking — and CI-001/CI-002 SHA pinning binds workflow `uses:` references, not Makefile variables. (2) The change's own spec artifacts already record the trust assumption as accepted (proposal Impact 'no lock integrity hashes'; design Risks 'transitive drift' / 'no byte-for-byte reproducible install'). (3) The meaningful version of the fix is new scope: a comment-recorded `dist.integrity` beside a name-addressed pin is audit-only — the runtime still resolves by name+version — and carries a staleness obligation on every pin bump, while mechanical verification of the hash must be a tested executable tool (DE-001), not prose. Accepted-risk classes are decision vehicles, not TODO lists: council feedback proposing to upgrade an accepted risk to a mitigation re-opens a design decision the author owns through the spec's risk record.
