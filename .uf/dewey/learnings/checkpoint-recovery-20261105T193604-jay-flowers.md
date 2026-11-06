---
tag: checkpoint-recovery
author: jay-flowers
category: gotcha
created_at: 2026-11-05T19:36:04Z
identity: checkpoint-recovery-20261105T193604-jay-flowers
tier: draft
---

A fail-closed re-entry gate that requires HEAD to exactly equal a ref frozen from GitHub blocks the recovery state a command's own push-failure path documents. Pattern (PR #955, /uf.address-feedback): Section 4.4 preserves local commits on push failure and says "Retry", leaving the head branch strictly ahead of the frozen head_sha; on re-run, Section 1.2 re-freezes the unmoved remote ref and the Section 1.2.1 re-entry gate STOPs on HEAD != head_sha before get_progress, so the checkpointed commits/push steps can never resume through the command. When a command both preserves unpushed local commits and keys resumable progress to the checked-out branch, its re-entry check must distinguish "frozen SHA is an ancestor of HEAD" (the run's own preserved commits — continue) from "stale or diverged" (stop), via `git merge-base --is-ancestor <frozen_sha> HEAD` on the head branch. Frozen-context consumers that resolve provenance from the frozen refs rather than the checkout are unaffected by accepting an ahead HEAD.
