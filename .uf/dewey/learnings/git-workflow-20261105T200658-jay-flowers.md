---
tag: git-workflow
author: jay-flowers
category: gotcha
created_at: 2026-11-05T20:06:58Z
identity: git-workflow-20261105T200658-jay-flowers
tier: draft
---

gh pr checkout does not guarantee a local branch named after the PR's head_ref: cross-repo (fork) PRs are checked out as <fork-owner>/<branch>. Command logic that verifies a gh pr checkout result, or resolves its push target through `git config branch.<name>.remote/.merge`, must key verification on commit identity (git rev-parse HEAD == the frozen head_sha) and read git config from the actual checked-out branch name (git branch --show-current) — never from head_ref. Relaxing the branch-name check without re-keying the config lookup silently resolves an unrelated local branch's tracking config (e.g. local main -> origin/refs/heads/main) and pushes commits to the wrong remote/ref. Precedent: uf.rebase-pr Phase 3 uses the actual head branch (<HEAD_BRANCH>) for its checkout fallback; uf.address-feedback Section 1.2.1 initially keyed both verification and the config lookup on head_ref (PR #955, feedback item 4).
