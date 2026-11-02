---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:07:25Z
identity: issue-triage-20261102T040725-jay-flowers
tier: draft
---

Triage feature requests against the repository's current state before judging actionability — the requested capability may have landed while the issue sat open. Example: unbound-force/unbound-force#636 (filed 2026-09-30) requested a declarative sibling-repo contract (clone URL, role, fetch policy, default branch, read-scope globs) plus a downstream-impact review persona. The repo now ships nearly that exact field set — .uf/sibling-repos.yaml v1 with roles downstream-consumer/upstream-source/sibling, fetch modes clone-on-review/contract-only, default branch, and bounded contract globs, consumed by the acquire_sibling_evidence tool behind untrusted-evidence delimiters (schemas/sibling-repos/README; review-council-multi-model-fanout design D8) — but delivers it as evidence embedded into all council reviewer prompts rather than a dedicated persona. An implemented request is still VALID (well-formed, real gap when filed); the triage action becomes scope reconciliation — diff the requested surface against what landed, then close or narrow — not fresh planning.
