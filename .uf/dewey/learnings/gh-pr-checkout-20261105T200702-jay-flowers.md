---
tag: gh-pr-checkout
author: jay-flowers
category: gotcha
created_at: 2026-11-05T20:07:02Z
identity: gh-pr-checkout-20261105T200702-jay-flowers
tier: draft
---

gh pr checkout does not guarantee a local branch named after a PR's head_ref: cross-repo PRs whose head is the fork's default branch check out as <fork-owner>/<branch> (e.g. contrib/main for head_ref main), and gh writes the branch's tracking config under that actual local name (branch.contrib/main.remote/.merge). Command procedures that verify PR-head presence must key identity on git rev-parse HEAD against the frozen head_sha, and push-target lookups must read branch.<name>.remote/.merge using the actually checked-out branch name. Keying either on head_ref false-STOPs for fork default-branch PRs and can resolve an unrelated local branch's push config (e.g. the operator's own main), sending git push to the wrong remote — so both corrections must ship in the same change. The port-uf-rebase-pr design (decision D3) records the same non-guarantee about gh pr checkout tracking config.
