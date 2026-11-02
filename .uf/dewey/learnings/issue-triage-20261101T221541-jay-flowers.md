---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:15:41Z
identity: issue-triage-20261101T221541-jay-flowers
tier: draft
---

Duplicate detection in unbound-force repositories must distinguish tutorial issues (title prefix "tutorial:", label "tutorial") from documentation-gate issues (title prefix "docs:") even when both cover the same command. A "tutorial: X" issue is not a duplicate of "docs: X new command" (the docs-gate tracker whose reference-doc updates — AGENTS.md table, docs/usage.md section, CHANGELOG entry — completed with the implementing change) or "docs: X re-route" (docs for internal re-plumbing): the tutorial delivers an audience-facing command-workflow walkthrough judged against a significance threshold (new command, phase count, human-decision gates, diagnosis taxonomy, bounded-iteration semantics) with the /uf.rebase-pr tutorial #855 as precedent. Verified on /uf.fix-ci: tutorial issue #870 legitimately coexists with docs issues #862 and #873 — different deliverables, different audiences, no scope overlap. The structural test for tutorial-vs-docs duplication: does the candidate deliver the same artifact to the same audience, or merely reference the same subject?
