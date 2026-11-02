---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:18:42Z
identity: issue-triage-20261102T021842-jay-flowers
tier: draft
---

When triaging documentation-gate issues that list "affected pages", grep the repository for the phrasing the issue tracks to verify the page list is complete — the author's page list can under-scope. Verified on unbound-force issue #728 (docs: /uf.unleash direct review dispatch + spec auto-fix consent gate): the issue correctly cited docs/usage.md:173,294 ("Run the full pipeline autonomously") and docs/heroes.md:82 ("Autonomous pipeline execution via /uf.unleash"), but the identical stale claim also sits in docs/architecture.md:448-455 (Autonomous Mode section), which the issue's page list omits. Pattern: for docs issues tracking a behavior change, search all user-facing docs for the old behavior phrasing rather than trusting the enumerated pages. Companion check for this repo: the org constitution's cross-repo documentation rule is satisfied separately by unbound-force/website#307 (same change, website surface), so an in-repo docs issue filed per the AGENTS.md gate is complementary, not a cross-repo duplicate — duplicate searches scoped to the current repo are the correct triage baseline.
