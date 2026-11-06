---
tag: documentation-pipeline
author: jay-flowers
category: gotcha
created_at: 2026-11-06T19:40:43Z
identity: documentation-pipeline-20261106T194043-jay-flowers
tier: draft
---

A fix-loop commit inside an already-gated PR must extend the content-pipeline trackers, not just the code lockstep surfaces. On PR #957 round 2, the CVE-remediation commit (49cfdcbe) correctly updated all seven code/spec lockstep files (both package.json manifests, provenance testdata, scaffold_test gate expectations, AGENTS.md, the OpenSpec spec delta, CHANGELOG.md) but left two tracker surfaces frozen at the pre-fix state: the website gate issue's tracking checkboxes still enumerated only the original task-4.3 areas, and the design doc's numbered decisions ended at D5 — while the follow-up docs issue (#977) cited a "design decision D6" that was never added, and was filed under the non-canonical `documentation` label instead of `docs`. Curator verification of a fix-loop round should therefore re-check, per fix commit: (a) whether the new user-facing delta appears in every gate tracker's enumeration (in-repo docs issue, website issue tracking section), (b) whether filed cross-references resolve to artifacts that actually exist, and (c) whether newly filed gate issues carry the label the duplicate-check queries filter on.
