---
tag: command-file-consistency
author: jay-flowers
category: gotcha
created_at: 2026-11-05T20:23:12Z
identity: command-file-consistency-20261105T202312-jay-flowers
tier: draft
---

When a change adds a new invariant rule to a command file (for example, \"Never call `get_progress` or `checkpoint` from another branch\"), it must reconcile every pre-existing passage in that file that performs the guarded action — especially session-resume guards whose steps run before the newly fixed sequence. PR #955 added branch-keyed progress rules to /uf.address-feedback but left the SESSION-RESUME GUARD's step 2 calling `get_progress` unconditionally before Section 1.2.1's `gh pr checkout`, so a compressed-context resume on a foreign branch could read and write progress under the wrong branch key — the exact #954 failure the PR closed. Fix pattern: a resume guard's state-recovery step references the branch-establishment section by number and defers the state read until after it; the branch re-check is unconditional on resume because the checkout step has no checkpoint of its own and \"first incomplete phase\" can otherwise bypass it. Mirror the edit into the scaffold asset and lock the ordering with a command-contract test clause.
