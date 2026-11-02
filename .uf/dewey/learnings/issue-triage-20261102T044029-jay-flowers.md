---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T04:40:29Z
identity: issue-triage-20261102T044029-jay-flowers
tier: draft
---

Triaging issues that cite openspec/changes/<name>/ artifacts is branch-sensitive: the artifacts live on the feature branch (the open PR), not necessarily in the current working tree. While triaging unbound-force issue #583 ([Docs] uf gate command), a glob showed openspec/changes/uf-gate/ absent, no internal/gate/ package, and no gate command in cmd/ — because the working tree was on an unrelated branch (opsx/host-fallback-provenance) while PR #582, carrying both the command and its spec artifacts, was still open. An empty glob in the working tree is not evidence that a cited artifact or feature is missing; confirm via the PR record (Dewey-indexed PR content or the review context) before concluding a citation is stale.
