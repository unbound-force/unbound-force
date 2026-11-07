---
tag: content-pipeline
author: jay-flowers
category: pattern
created_at: 2026-11-07T19:14:36Z
identity: content-pipeline-20261107T191436-jay-flowers
tier: draft
---

CHANGELOG entries and OpenSpec proposal Impact sections that name their governance-issue references inline make documentation gates verifiable from the review diff alone. In unbound-force PR 948 (quick/uf-workflow-tool-fixes), each user-facing CHANGELOG entry carried (Fixes: #N) plus (Docs: unbound-force#963, unbound-force/website#379), and the retroactive proposal's Impact section stated "Docs issues filed" with both issue numbers — so a reviewer running without GitHub access (dispatch-confined, gh prohibited, or offline) could confirm both documentation-issue gates (the repo AGENTS.md documentation gate requiring an issue against the current repo, and the org constitution's Cross-Repo Documentation rule requiring an issue in unbound-force/website) were discharged without live gh issue list searches. Pattern: when a change satisfies a tracked-issue gate, record the issue reference in a diff-visible artifact (CHANGELOG entry, proposal Impact section) rather than only in the PR body or commit message — the reviewer's copy of the diff then carries the gate's evidence, and reviewers confined from GitHub can verify compliance instead of degrading to an unverified-gate finding.
