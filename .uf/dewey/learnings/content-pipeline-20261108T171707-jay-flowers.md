---
tag: content-pipeline
author: jay-flowers
category: gotcha
created_at: 2026-11-08T17:17:07Z
identity: content-pipeline-20261108T171707-jay-flowers
tier: draft
---

A change that reverses a documented policy should sweep the issue tracker for open requests that asked for the old behavior before merge. Verified on feedback-loop-scope (PR #993): open issue unbound-force/unbound-force#484 ("feat: /uf.finale should ensure Dewey learning files are included in the PR") requests the exact behavior the PR removes — uf_finale_stage now excludes .uf/dewey/learnings/ and the AGENTS.md Commit scope rule forbids committing them on PR branches — leaving the issue contradicting shipped policy. The change's proposal says it reverses the unarchived dewey-learning-pr-policy and finale-dewey-commit changes but names no issue reconciliation. Triage heuristic: when a proposal says it reverses prior rules, keyword-search open issues for the old rule's ask (e.g. "learnings included in the PR") and recommend closing or annotating the stranded requests in the same change.
