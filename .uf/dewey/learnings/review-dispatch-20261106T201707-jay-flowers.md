---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-06T20:17:07Z
identity: review-dispatch-20261106T201707-jay-flowers
tier: draft
---

A diff captured against a stale upstream/main on a branch stacked over an unmerged sibling PR includes the sibling's files as new: PR #959's 31-file diff contained divisor-linguist.md (+203) that belongs to stacked PR #958, and the CHANGELOG/spec attribution for those files belongs to the sibling, not the head PR. Before attributing a missing spec, missing changelog entry, or documentation gap to the head change, verify attribution against the working tree (content-grep openspec/ and specs/ for the artifact names) and look for the stacked-PR relationship in the change's learnings or PR description. Two verification traps: file-level globs whose pattern matches a directory name return nothing (search with file-level patterns like */*.md or content grep), and review-context discovery that matches specs to branch names will report "no spec artifact found" even when a previously-merged change partially covers the domain — read the nearest prior change's proposal and tasks to scope what it did and did not cover.
