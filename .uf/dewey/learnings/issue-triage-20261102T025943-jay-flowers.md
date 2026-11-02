---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:59:43Z
identity: issue-triage-20261102T025943-jay-flowers
tier: draft
---

Prose→tool extraction issues (the unbound-force #659/#716 wave) can misdescribe the behavior they extract: issue #675 says the review-pr fix branch is cut "from PR head" with `git fetch`, but the live command prose cuts it from the base branch (`git checkout <baseRefName>` → `git checkout -b fix/pr-...`, matching FR-011's pre-existing-failure semantics), and the issue's own tool sketch (`base_branch` arg defaulting to "main") contradicts its prose. When triaging extraction issues, diff the issue's behavior description against the actual command file — line citations drift as files evolve, and the parity acceptance criteria plus the live prose, not the issue's mechanical description, are the authoritative contract the extracted tool must preserve.
