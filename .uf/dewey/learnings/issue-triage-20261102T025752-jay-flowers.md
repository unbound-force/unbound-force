---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:57:52Z
identity: issue-triage-20261102T025752-jay-flowers
tier: draft
---

Triage "create a plugin tool that encodes X patterns" issues by first searching the plugin suite for an existing in-code implementation of the same deterministic logic. Issue #673 (classify_paths in uf-workflow) proposed re-encoding path-classification patterns whose category taxonomy (security, cli-ux, test-quality, documentation, ci-cd, dependencies, standard) and security_sensitive/user_facing flags already exist as the change-profile computation in .opencode/plugins/review-dispatch/index.ts feeding plan_review_dispatch. The repo's shared-primitives pattern — .opencode/lib/ pure functions such as classifyBranch (branch-classification.ts) and sanitizeSlug (uf-workflow-types.ts) — is the canonical resolution: extract the existing implementation to lib and wrap it, never re-implement. Dual classifiers for the same paths drift; acceptance criteria reading "encode all patterns in the new tool" are the tell that a duplicate is being specified.
