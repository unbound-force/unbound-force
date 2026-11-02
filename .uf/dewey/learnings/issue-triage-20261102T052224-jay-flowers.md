---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T05:22:24Z
identity: issue-triage-20261102T052224-jay-flowers
tier: draft
---

Triage of supply-chain pinning complaints in the Unbound Force org must distinguish two pinning tiers: Constitution Principle V and CI pack CI-001 make commit-SHA pinning of GitHub Actions `uses:` references a MUST (violation = bug; the severity pack classifies "unpinned CI action on mutable tag" as Adversary HIGH), while CI-001 explicitly excludes Docker/container image references — digest-pinning those is only a SHOULD ("not subject to this rule"). Issues claiming container `FROM` tags violate the constitution overstate that half; validate the action-pinning half as the hard violation and route container digest pinning as a standards-aligned enhancement. Also verify quoted constitution text, not line citations: amendments (e.g., v1.3.1→v1.4.0 on 2026-10-14) shift lines, so an issue filed 2026-08-31 citing lines 109-112 now points at Principle II/III while the quoted rule lives at ~173-174.
