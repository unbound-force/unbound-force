---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:51:39Z
identity: issue-triage-20261102T035139-jay-flowers
tier: draft
---

Diff-line-based review tiering over-weights byte-copy scaffold mirrors. In the unbound-force meta repo, files under internal/scaffold/assets/ are byte-identical twins of canonical sources (enforced by drift-detection tests), and syncing them is a mandatory phase of nearly every command/plugin change — so scaffold syncs contribute large diff line counts with zero original review signal. Observed on PR #653: two asset syncs (uf.review-pr.md 315+/759-, review-matrix.yaml 3+/6-) contributed 1083 of 1337 diff lines, inflating a ~20-meaningful-line change from lightweight to heavy review tier. When triaging or sizing diffs dominated by internal/scaffold/assets/ (or any generated/mirrored path), discount those lines before judging review scope — the review signal lives in the canonical source change, which typically appears in the same diff.
