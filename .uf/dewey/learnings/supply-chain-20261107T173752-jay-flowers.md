---
tag: supply-chain
author: jay-flowers
category: gotcha
created_at: 2026-11-07T17:37:52Z
identity: supply-chain-20261107T173752-jay-flowers
tier: draft
---

Before adding an executable anchor to a hand-maintained hash record, establish which state the recorded hash refers to — current-state gate and rollback-target anchor are different contracts. In unbound-force, plugin-package-rollback.json's baseline_lock_sha256 is documented (SC-FR-003; review-council-multi-model-fanout design D13; fix-triage-issue-dispatch-docs "anchors verified rollback to 1.4.10") as the SHA-256 of the ROLLBACK-TARGET lock — the superseded pin state that rollback restores — not the current lock. A test that hashes the current lock and compares it to the baseline either fails (correctly recorded baseline) and tempts re-pointing the record at the current hash — silently destroying the rollback information, since superseded values are not retained elsewhere in the tree — or passes against a record already computed from the wrong artifact, locking in vacuous semantics. Safe implementations: (a) track the rollback-target lock as a testdata artifact and hash that, making the spec's "tracked baseline lock" literal, or (b) add a separate, clearly-named current-state whole-file lock-hash const as a new pin-enforcement record, leaving the rollback record's meaning intact. Symptom that the record needs this analysis at all: when a hash is only compared const-to-fixture with both values hand-updated in one commit, the tree cannot demonstrate which artifact the hash refers to. Also enumerate every carrier before moving a pin — the baseline hash lives in the test const, the testdata JSON, and the unarchived spec MUST text; moving only the first two leaves an active spec contradicted by the shipped tree (the clarification-marker-superset delta still pins Zod 4.1.8 against the shipped 4.6.6 after the PR-948 bump).
