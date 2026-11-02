---
tag: opencode-v2-migration
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:34:06Z
identity: opencode-v2-migration-20261102T003406-jay-flowers
tier: draft
---

Runtime major-version cutover sequencing (installer side): when a tool's installer ships unpinned "latest" (internal/setup/setup.go installing opencode), downstream machines receive the new major immediately — that unpinned path is the forcing function that makes dual-runtime support land BEFORE any default flip. The safe ladder, modeled by unbound-force issue #801 (OpenCode v2 migration Phase 2): (1) version-gated dual emission keyed on the existing runtime version probe, (2) an explicit opt-in install path for the new major behind a flag, (3) the default flip deferred to a later, separately-tracked phase, and (4) post-install/doctor guidance reporting which major is installed so version skew is observable rather than silent. Complementary to the existing dual-emission drift-test learning (opencode-v2-migration-20261102T003016-jay-flowers), which covers the emission side; this covers the install-path and observability side. Applies to every hero repo with an install path during the org-wide v2 migration.
