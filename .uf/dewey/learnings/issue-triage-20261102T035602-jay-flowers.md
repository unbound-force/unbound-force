---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:56:02Z
identity: issue-triage-20261102T035602-jay-flowers
tier: draft
---

When triaging an issue that requests parity with a sibling file's pattern, verify the cited pattern against the current HEAD rather than trusting it as of the issue's filing date. Issue #652 (filed 2026-10-05) asked uf.review-pr to replicate uf.unleash's in-file execution checklist and mandatory TodoWrite dual-bookkeeping. The unleash-pipeline-hardening change (field audit 2026-10-21, openspec/changes/unleash-pipeline-hardening/) subsequently removed the in-file checklist and demoted TodoWrite to a RECOMMENDED display aid with the checkpoint tool as the durable record of truth, after the audit observed the dual bookkeeping maintained at only 7 of ~22 required points. The underlying gap (uf.review-pr has no progress tracking or step-level resumability) remained valid, but the correct fix follows the newer durable-state pattern (get_progress/checkpoint plus TodoWrite as display aid) already shipped in uf.review-council, uf.fix-ci, and uf.address-feedback — not the removed checklist pattern. Triage verdicts should separate "the gap is real" from "the requested mechanism is current"; a stale mechanism citation downgrades the remedy, not the issue.
