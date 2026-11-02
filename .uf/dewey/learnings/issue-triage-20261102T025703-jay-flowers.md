---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:57:03Z
identity: issue-triage-20261102T025703-jay-flowers
tier: draft
---

Documentation-gate issues that track a shipped tool change can be triaged against three objective evidence surfaces before any clarification is requested: (1) the tool's registered description in the runtime environment, (2) the originating change's spec artifacts, and (3) the implementing PR. Triage of unbound-force issue #690 (document resolve_base_ref) was confirmed VALID without author interaction because all three surfaces matched the issue's claims verbatim — the registered resolve_base_ref tool carries the exact three-level fallback chain (upstream/main → origin/main → main, no network access), the openspec/changes/fix-stale-base-ref/ requirements specify the same ordering and failure semantics, and PR #691 (Closes #687) shipped all three changes the issue enumerates. Additionally, AGENTS.md's documentation gate mandates filing exactly this kind of issue against the current repo for user-facing changes, so such issues are process artifacts to validate against their originating change — not proposals whose necessity needs re-litigating.
