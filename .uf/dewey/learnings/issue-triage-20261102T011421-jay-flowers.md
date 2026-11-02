---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:14:21Z
identity: issue-triage-20261102T011421-jay-flowers
tier: draft
---

unbound-force documentation-gate issues can keyword-collide with later gap-report issues filed about the same documentation surface. Example: #772 (docs-gate tracker for the unleash-tier-routing change, created 2026-10-20, enumerating AGENTS.md quick-tier row, CHANGELOG entry, docs/usage.md + docs/architecture.md via a scope-extension comment, and website docs in the website repo) vs #826 ("risk floors missing from docs/usage.md workflow table and AGENTS.md Quick row"). The gate issue is a per-change delivery tracker filed before PR merge to satisfy the AGENTS.md documentation gate; the gap issue reports residual omissions after delivery. They overlap on file surface but track different work items. When triaging a docs-gate issue, disambiguate duplicate candidates by issue type and change linkage, not by shared file paths: same-change same-type issues are duplicates; different-change gate issues are separate valid issues (prior learning issue-triage-20261102T001324); gate-vs-gap pairs are complementary and should be cross-referenced rather than merged, with the gate issue classified VALID / enhancement / objective. Verify the gate issue's enumerated impact against current repo state (e.g., the AGENTS.md Specification Workflow Quick tier row) to confirm its accuracy.
