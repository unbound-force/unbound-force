---
tag: spec-governance
author: jay-flowers
category: pattern
created_at: 2026-11-07T16:10:41Z
identity: spec-governance-20261107T161041-jay-flowers
tier: draft
---

When a PR reconciles multiple unarchived OpenSpec sibling changes, verify each sibling received the same artifact-class coverage — the detection signal is treatment inconsistency across siblings, not per-sibling completeness. Observed on unbound-force PR #957 (stop-shipping-plugin-lock): the reconciliation amended ts-plugin-coverage-gate and init-override-stubs-gitignore-sync across spec + proposal + design + tasks.md, but fix-triage-issue-dispatch-docs got only spec + proposal + design, leaving its tasks.md (six stale references: deleted lock mirrors, deleted rollback testdata, pre-rename test names, lockfile-regeneration guidance) and its design's Migration Plan ("npm ci... CI installs from the lockfile" — now operationally false) unreconciled. Review recipe: after confirming the term sweep (replaced command, artifact filename, renamed identifiers) ran, diff the AMENDED-FILE SETS across the reconciled siblings — a sibling missing an artifact class that its co-siblings received (here: tasks.md) marks the gap; then read that sibling's unamended artifact class for term hits, classifying each as forward-contract (amend) or historical record (document the classification in the reconciliation section). A repo-wide grep per term class per sibling directory is the cheap completeness check; the inconsistency between siblings is what makes the miss visible without re-reading everything.
