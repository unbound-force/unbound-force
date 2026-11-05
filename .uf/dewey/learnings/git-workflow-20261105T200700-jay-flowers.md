---
tag: git-workflow
author: jay-flowers
category: gotcha
created_at: 2026-11-05T20:07:00Z
identity: git-workflow-20261105T200700-jay-flowers
tier: draft
---

gh pr checkout names a cross-repo PR's local branch `<owner>/<branch>` when the PR head is the head repository's default branch (e.g. `contrib/main` for head_ref `main`), and writes `branch.<actual-name>.remote`/`.merge` tracking config under that actual local name. Command procedures that verify PR-head identity must key on the frozen head SHA (`git rev-parse HEAD` == `head_sha`), never on `git branch --show-current` == `head_ref`, and push-target lookups must use the actual checked-out branch name, not `head_ref`. Otherwise fork default-branch PRs false-STOP — or, if only the name check is relaxed, the lookup silently reads the operator's own `branch.<head_ref>.*` config and pushes to the wrong remote. Verified against unbound-force PR #955 feedback item 4 (uf.address-feedback Section 1.2.1); the repo's uf.rebase-pr design (openspec/changes/port-uf-rebase-pr, decision D3) already follows the correct verify-actual-upstream pattern and warns that gh pr checkout's tracking is not guaranteed.
