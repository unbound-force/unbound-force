---
tag: review-fix-sweep
author: jay-flowers
category: pattern
created_at: 2026-11-02T23:07:47Z
identity: review-fix-sweep-20261102T230747-jay-flowers
tier: draft
---

When a review-fix pass tightens or changes specified behavior, sweep every artifact of the change, not just the code: implementation, spec deltas, regression tests, and CHANGELOG — AND the prose artifacts (design.md decisions, proposal.md "What Changes" items). On branch opsx/fix-uf-finale-ci-poll-gh-schema (2026-11-02), the code-review round-1 fix tightened the PR-not-found stderr pattern in the code, the FCP-FR-002 spec delta, the regression test, and the CHANGELOG, but left design.md D5 and proposal.md item 5 describing the superseded loose four-phrase pattern — and all four round-2 reviewers (adversary, architect, curator, guard) independently flagged the intra-change contradiction: the design record was documenting exactly the defect class the fix had removed, so a future maintainer implementing from D5 would reintroduce it. The verdict was still APPROVE WITH ADVISORIES because the added regression test bounded the practical risk, but the entire advisory cluster was preventable with a fix-pass checklist step that greps the change directory for text describing the superseded behavior.
