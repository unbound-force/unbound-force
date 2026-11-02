---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:19:40Z
identity: issue-triage-20261101T221940-jay-flowers
tier: draft
---

When a change adds plugin tools through internal mechanical extraction (no new user-facing capability debuts, command surface unchanged, observable behavior preserved), the documentation gate and content-pipeline significance thresholds evaluate independently: the gate fires because new plugin tools are user-facing capability surface shipped to every uf-init consumer via the scaffold mirror, while the blog thresholds (new agent, new CLI command or subcommand, architectural migration, new hero capability) and tutorial thresholds (new multi-step slash-command workflow, new tool integration requiring setup, new workflow pattern) legitimately decline. This complements the repair-change case (learning content-pipeline-20261016T172845-jay-flowers: restored capability is not new capability) — both patterns hold that gate applicability and content significance are separate evaluations. Issue #873 (extend-fetch-pr-review-context) exemplifies the extraction case: the docs issue was filed as the gate record, blog and tutorial were both declined with threshold-by-threshold justification recorded in the issue body, and website tracking was consolidated on existing trackers (unbound-force/website#342, unbound-force/website#310) instead of filing duplicates.
