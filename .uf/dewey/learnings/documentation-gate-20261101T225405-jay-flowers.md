---
tag: documentation-gate
author: jay-flowers
category: gotcha
created_at: 2026-11-01T22:54:05Z
identity: documentation-gate-20261101T225405-jay-flowers
tier: draft
---

When triaging documentation-impact tracking issues (the AGENTS.md documentation-gate artifact), cross-check the originating change's tasks.md for recorded issue numbers before treating a "no website tracking issue known" disposition as current. Documentation-gate issues are filed early (before PR merge) while the sibling unbound-force/website issue may be filed later and recorded only in the change's tasks.md, leaving the issue body's disposition section as the stale side of the pair. Verified during triage of unbound-force/unbound-force#844 (replicator-version-floor docs impact, filed 2026-10-27): the issue body says no website tracking issue is known, but the change's tasks.md task 3.2 — marked complete by 2026-10-28 — records unbound-force/website#337 as the filed website counterpart. The #739 per-surface disposition pattern anticipates this via its "confirmed or updated in the implementing PR" acceptance criterion, so the fix is linking the recorded number in the implementing PR rather than re-filing.
