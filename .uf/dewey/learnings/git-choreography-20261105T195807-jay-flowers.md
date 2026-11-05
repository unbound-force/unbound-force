---
tag: git-choreography
author: jay-flowers
category: gotcha
created_at: 2026-11-05T19:58:07Z
identity: git-choreography-20261105T195807-jay-flowers
tier: draft
---

A pre-push divergence gate built as `git fetch <remote> <ref>` followed by `git status` silently fails when <remote> is a URL: URL fetches write only FETCH_HEAD (no remote-tracking ref is created or updated), and `git status` ahead/behind is computed against the configured upstream tracking ref, which a URL remote cannot have. Whenever a push target is recorded from `git config branch.<name>.remote` (e.g., after `gh pr checkout` of a cross-fork PR, where the value may be the fork's URL), gate on an explicit comparison against the fetched tip instead: `git rev-list --left-right --count HEAD...FETCH_HEAD`, where a non-zero right-hand count means the remote holds commits HEAD lacks and the push would be rejected as non-fast-forward. The comparison must immediately follow the fetch because FETCH_HEAD reflects only the most recent fetch.
