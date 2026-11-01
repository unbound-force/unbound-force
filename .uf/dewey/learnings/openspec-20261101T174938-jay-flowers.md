---
tag: openspec
author: jay-flowers
category: gotcha
created_at: 2026-11-01T17:49:38Z
identity: openspec-20261101T174938-jay-flowers
tier: draft
---

When an OpenSpec change establishes a requirement outright via ADDED (because no predecessor delta ever ADDED that requirement name), name-continuity and FR-identifier allocation are independent obligations. Archive applies MODIFIED deltas by requirement name, so adding the `**ID: <CAP>-FR-NNN**` id line to the ADDED body never affects supersession or archive mechanics — but declining to allocate leaves the canonical spec with a permanently id-less requirement while the capability's live id series breaks its traceability chain. Before authoring the ADDED entry, grep the capability's highest allocated id across `openspec/changes/*/specs/<capability>/spec.md` (e.g. `RD-FR-03`), allocate the next one, and re-run `openspec validate --strict` — the id line coexists with the 500-character ADDED description budget in established practice (RD-FR-028 and RD-FR-032 both carry it). Observed on accept-variantless-parent-model (quick/accept-variantless-parent-model branch, 2026-11-01): its ADDED 'Finalization Model Validation' inherited the id-less state of advisor-host-model-fallback's orphan MODIFIED entry, and design.md's "none is invented here" non-goal cited MP-002's no-renumbering rule — which governs existing identifiers, not new allocations for newly established requirements.
