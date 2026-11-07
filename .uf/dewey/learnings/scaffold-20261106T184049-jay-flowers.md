---
tag: scaffold
author: jay-flowers
category: decision
created_at: 2026-11-06T18:40:49Z
identity: scaffold-20261106T184049-jay-flowers
tier: draft
---

internal/scaffold/scaffold.go (PR 948, head 5663552f) carries exactly two legacy-directory migration functions sharing a structural skeleton: migrateCommandDir (scaffold.go:1978-2084; flat os.ReadDir merge; two-branch policy — file in both dirs: byte-compare, warn on diff, keep the commands/ copy and remove the old; old-only: move) and migrateSkillDir (scaffold.go:2100-2251; recursive filepath.WalkDir merge; four-branch policy — new-only: MkdirAll+move; byte-identical: remove old as duplicate; tool-owned scaffold skill via isScaffoldSkillFile: remove old as stale; other differing: keep BOTH copies with conflict warning; plus deepest-first empty-dir cleanup and idempotent-rerun "skipped" action). The shared skeleton — DivisorOnly guard, Lstat probe with silent no-op, symlink skip, atomic rename when only the old dir exists, subToolResult failure/migrated reporting — spans roughly 35-45 lines. DEFERRAL: extraction was deferred at two occurrences because default.md CS-004 mandates extraction only for logic appearing in MORE THAN two locations (exactly two is below the written threshold), following the repo's established two-occurrence posture (scaffold-20260816T184148: "if a third consumer appears, that triggers extraction"; tolerant-gh-runner-20261102T230744: twin-pointer with third-consumer extraction trigger). TRIGGER — recorded to fulfill the PR 948 review-thread commitment (comment 4484508376) so it survives context loss: when a THIRD directory migration is added to internal/scaffold, extract the shared probe/rename/report skeleton FIRST, with the per-directory merge policy as the varying part, BEFORE implementing the third migration's policy; span the extraction across all copy sites in that change (review-dispatch-20261019T212331), not just the new one.
