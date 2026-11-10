---
tag: documentation-pipeline
author: jay-flowers
category: gotcha
created_at: 2026-11-08T23:16:31Z
identity: documentation-pipeline-20261108T231631-jay-flowers
tier: draft
---

When the Curator runs as a dispatched review child (review-council dispatch), the dispatch contract prohibits creating issues and `gh issue create` is "ask"-permissioned — a prompt a parented child session cannot answer. The correct degradation for the docs-gate checklist is a blocking finding that carries the full, ready-to-run `gh issue create` commands for both gate targets (in-repo docs issue + unbound-force/website sync issue, each with labels and a body drafting the per-surface dispositions and tutorial/blog judgments), plus the CHANGELOG/proposal.md cross-reference edits, so the PR author files them before merge. Verify gate-artifact absence per repo and per label with read-only `gh issue list` (label enumeration + label-agnostic full-text searches) before reporting. Verified on PR #990 (advisory-approval-event): no docs issue existed in either repo; the finding supplied both commands.
