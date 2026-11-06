---
tag: learning-systems
author: jay-flowers
category: gotcha
created_at: 2026-11-05T22:22:48Z
identity: learning-systems-20261105T222248-jay-flowers
tier: draft
---

Committed Dewey learning artifacts can arrive truncated mid-word: five of 36 in PR #955 (commit 64457c88) end mid-word on their single body line (e.g. '...ordering the actual call after i', '...then subm'), each with exactly 10 lines and valid frontmatter, so the truncation is invisible to file-presence checks and to git status review. The harvest/proposal write path persists partial content with no sentence-boundary handling. Verify harvested learning bodies end with a sentence terminator before committing them to a feature branch; longer-term, fail or re-truncate cleanly at the harvest boundary. Affected files: contract-testing-20261105T214916, contract-tests-20261105T214916, curator-20261105T214916, workflow-resume-gates-20261105T214909, error-handling-20261105T214911.
