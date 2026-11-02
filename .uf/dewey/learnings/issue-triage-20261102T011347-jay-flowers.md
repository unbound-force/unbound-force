---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:13:47Z
identity: issue-triage-20261102T011347-jay-flowers
tier: draft
---

Documentation-tracking issues for merged behavior fixes validate cleanly in triage when they (a) enumerate each user-facing outcome change as a checkbox item traceable to the change's spec FRs, (b) include an explicit non-regression item for commands whose outputs are unchanged, and (c) delineate adjacent non-overlapping docs issues by number. unbound-force issue #778 (docs for fix-dispatch-verdict-derivation) exemplifies the pattern: six checklist items map to FR-009–FR-012 outcomes across /uf.address-feedback and /uf.triage-issue panel shapes, with issues #766 and #754 explicitly scoped out — the duplicate search then returns only the issue itself, and no split is warranted because a single change's docs obligation shares target surfaces.
