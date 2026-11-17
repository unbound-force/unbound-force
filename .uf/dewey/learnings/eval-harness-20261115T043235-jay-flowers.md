---
tag: eval-harness
author: jay-flowers
category: gotcha
created_at: 2026-11-15T04:32:35Z
identity: eval-harness-20261115T043235-jay-flowers
tier: draft
---

Eval fixture setup.sh scripts that patch repo config in the eval workdir (opencode.json providers, .uf/review-matrix.yaml model pins) must commit those patches onto the mini-repo's base branch and rebase the feature branch onto it, because capture_review_diff in base mode diffs the working tree against the base ref — uncommitted config patches leak into the reviewed diff and silently change what the council reviews (discovered in the review-council-empty-reviewer fixture: a 19-line .uf/review-matrix.yaml patch appeared alongside the 9 intended feature files). Note that opencode.json at the repo root is gitignored, so provider patches stay untracked workdir files that no diff can see — only tracked config like the review matrix needs the commit-and-rebase treatment. A dry-run of `git diff --numstat` against the base inside the setup script catches this class of contamination before it ships.
