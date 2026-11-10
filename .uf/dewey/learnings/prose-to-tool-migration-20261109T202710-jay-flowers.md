---
tag: prose-to-tool-migration
author: jay-flowers
category: pattern
created_at: 2026-11-09T20:27:10Z
identity: prose-to-tool-migration-20261109T202710-jay-flowers
tier: draft
---

When a prose→tool extraction lands, audit prior learnings for failure modes that named the extraction target as their fix vehicle — the learning's fix shape often extends beyond the issue's acceptance criteria. In PR #1002 (checkout-pr-head), learning gh-pr-checkout-20261105T225116-jay-flowers documented a wrong-repository push failure mode (a local branch named exactly head_ref whose remote points at a different repository carrying the same commits passes every name-based gate) and named the #965 checkout_pr_head tool as the place to verify head_remote against the PR's head repository; the implemented tool ports all six prose gates faithfully but not the repository-identity check, and the design doc's Risks section does not acknowledge the carried-forward gap. Extraction reviews should diff the new tool's gate set against documented failure modes, not just against the prose it replaces.
