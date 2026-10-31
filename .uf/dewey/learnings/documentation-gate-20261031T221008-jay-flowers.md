---
tag: documentation-gate
author: jay-flowers
category: gotcha
created_at: 2026-10-31T22:10:08Z
identity: documentation-gate-20261031T221008-jay-flowers
tier: draft
---

Documentation-gate satisfaction in unbound-force repos is two-tracked: the in-repo docs issue (repo AGENTS.md documentation gate) and the unbound-force/website tracking issue (org constitution Cross-Repo Documentation rule) are separate obligations — filing one does not satisfy the other, and both must exist before PR merge for user-facing changes. Repair changes responding to a council 'missing documentation-issue' finding are at risk of partial satisfaction: the fix-loop author files the in-repo issue, cites it in CHANGELOG and proposal, and the artifact record then reads as gate-complete while the website issue is missing. Observed on quick/accept-variantless-parent-model (2026-10-31): in-repo #891 was filed and cited, but no website issue existed (verified by searching unbound-force/website across open and closed states for the change's keywords); precedent PR #744 satisfied both tracks (#740 + website#310). Curator verification step before accepting any docs-gate claim: search unbound-force/website across all states for the change name and its user-facing keywords, not just the in-repo tracker.
