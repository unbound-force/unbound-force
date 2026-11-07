---
tag: regression-protection
author: jay-flowers
category: gotcha
created_at: 2026-11-07T19:57:09Z
identity: regression-protection-20261107T195709-jay-flowers
tier: draft
---

A pin-parity audit performed mid-fix-series goes stale the moment a later commit in the same series adds a new surface: re-run the enumeration grep after the last commit, and never record an "only unguarded surface" claim without a fresh grep. Precedent: PR #957 round 3 — the retrospective learning regression-protection-20261107T190359 (written 19:05) claimed the four remediation constants were "the only npm-pin surface that can drift from PLUGIN_TEST_NPM without a failing test," but the closing commit 32681e4d had added the ci_scheduled.yml osv_plugin_transitive job carrying npx -y npm@11.6.2, GOTOOLCHAIN=go1.27.0, and osv-scanner@v2.7.0 with only prose "keep in sync" comments and no test sweep (grep: no *_test.go references ci_scheduled; the twin tests read ci_security.yml only). The retrospective enumerated the pin graph as it existed when the earlier learnings were written, not as the final head commit left it, so both the coverage and the learning's exhaustive claim were wrong in the same change. Rule: when a fix series lands new copies of governed constants after its retrospective artifacts were authored, either extend the parity guards to the new surface or amend the learning before merge; treat any "only X" claim about an unguarded-surface set as requiring a same-day grep against head.
