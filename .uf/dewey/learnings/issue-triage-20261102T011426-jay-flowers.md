---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:14:26Z
identity: issue-triage-20261102T011426-jay-flowers
tier: draft
---

Triage verification for documentation-gate issues is bidirectional: beyond confirming the originating change directory exists, search the change's indexed tasks.md for the issue number itself — a well-formed gate issue is recorded as the fulfillment of a specific docs task (e.g., openspec/changes/unleash-tier-routing tasks.md Phase 7 task 7.3 records unbound-force/unbound-force#772 verbatim, alongside the separate website-repo gate issue unbound-force/website#314 and PR #315, confirming both documentation gates were kept distinct). Duplicate-candidate ordering then settles the duplication check cheaply: GitHub issue numbers are sequential per repository, so every candidate with a higher number than the issue under triage was created later and cannot make it a duplicate; later topical overlaps (e.g., #826's risk-floor documentation gap vs #772's broader gate scope, whose title already includes risk floors) are follow-up gap reports within the earlier issue's topic, not duplicates.
