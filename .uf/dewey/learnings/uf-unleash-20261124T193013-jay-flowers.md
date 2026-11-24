---
tag: uf-unleash
author: jay-flowers
category: gotcha
created_at: 2026-11-24T19:30:13Z
identity: uf-unleash-20261124T193013-jay-flowers
tier: draft
as_of: "unbound-force@12bb5f8799eea38393959349f27354e1bb031a77"
evidence:
  - "session:uf.unleash"
  - ".uf/progress/opsx-ci-skip-learnings-only-prs/uf.unleash.json"
---

Phase-checkpoint inventory must be verified evidence, never reconstructed from memory. On 2026-11-24 the audit caught implement-phase-3/4 checkpoints recorded green with guessed inventory (a pre-existing errcheck entry asserted without a fresh baseline run). The correction procedure: re-run every gate command fresh, establish the pre-existing classification with a Tier 2 baseline worktree at the merge base (git worktree add <tmp> <base-sha>; run the failing check; remove the worktree), then re-record each checkpoint with the verified inventory and the required human_decision_ref citations. A guessed inventory in a checkpoint is a fabricated gate result — the most severe process violation an autonomous pipeline can commit short of touching main.
