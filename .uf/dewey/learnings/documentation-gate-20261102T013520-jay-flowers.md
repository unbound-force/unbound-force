---
tag: documentation-gate
author: jay-flowers
category: gotcha
created_at: 2026-11-02T01:35:20Z
identity: documentation-gate-20261102T013520-jay-flowers
tier: draft
---

Docs-sweep gate issues are scoped to the governance artifact's version at filing and go stale as the artifact advances — execute the sweep against the artifact's current state, not the issue's stated version. Issue #762 scopes its sweep to the DE-005/DE-006 rules determinism.md gained at v1.1.0 (router-pattern-codification, PR #768), but the pack now sits at v1.4.0 with DE-007 (Seam Verification) and DE-008 (Delegation Skill Enforcement) added by successor changes. An executor enumerating only DE-005/DE-006 produces a rule inventory stale on arrival unless the docs explicitly scope to the pre-1.1.0 set — the issue's acceptance criteria permit exactly that scoping, which is also the lower-maintenance alternative, since docs that mirror governance rule inventories acquire a sync obligation on every future DE-NNN addition. Before sweeping, confirm the successor bumps carried their own gate records (#781 covers DE-007 at v1.2.0) so per-change scoping stays coherent. Extends learning issue-triage-20261102T011437-jay-flowers (triage-time pack-frontmatter verification and duplicate discrimination) with the execution-time hazard.
