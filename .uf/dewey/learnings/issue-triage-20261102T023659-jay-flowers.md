---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:36:59Z
identity: issue-triage-20261102T023659-jay-flowers
tier: draft
---

A documentation-tracking issue's behavioral prescriptions can be superseded by a later change, and scope currency is verifiable in three artifact layers before ruling: (1) the follow-up change's spec-delta supersession annotations — finale-token-reduction's uf_finale_stage delta records "Superseded by fix-tui-git-output-leak ... the staged file list is returned in the JSON result envelope as files: string[]"; (2) the follow-up proposal's Documentation Impact — fix-tui-git-output-leak states docs issue #710 is "referenced rather than duplicated; its scope was updated with a comment"; (3) the scope-update comment on the tracker itself — unbound-force/unbound-force#710 comment 6175394955 (envelope-only output, uf_finale_stage returning {fileCount, files}, silent uf_finale_ci_poll). Triage rule: when a docs issue prescribes how to describe tool behavior, verify scope currency through all three layers. A tracker with a recorded supersession chain is current and actionable; a tracker whose described contracts were superseded with no comment anywhere has a stale-description problem to fix in the issue body — not grounds for an invalid or duplicate verdict.
