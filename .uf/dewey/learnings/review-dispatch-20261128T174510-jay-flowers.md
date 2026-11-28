---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-28T17:45:10Z
identity: review-dispatch-20261128T174510-jay-flowers
tier: draft
as_of: "unbound-force@444cb92a90a02f80a8798f4de1152ff133d7abe4"
---

divisor-determinism Step 1 evidence reflects the session working tree, which in PR-mode council dispatches can be the PR BASE checkout rather than the head. Reviewing unbound-force PR #1240 (2026-11-28): the pre-flight gates ran in a separate head worktree (c1bfbed) while the reviewer session's directory sat at base 444cb92 — instruction-lint --base <PR base> --changed returned an empty delta, and grep showed pre-change workflows and the supposedly deleted scripts still on disk. An empty linter delta in that setup means "the session tree matches base", not "the PR changes no instruction files". Determinism reviewers must reconcile the caller's diff (authoritative for the PR's changes) against the linter delta (authoritative only for the session tree), disclose the mismatch in the review summary, and review the changed instruction text by reading base + diff; the .uf/review-head worktree may already be released by review time, so reconstruction from base + diff is the fallback.
