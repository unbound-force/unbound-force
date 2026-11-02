---
tag: constitution-alignment
author: jay-flowers
category: gotcha
created_at: 2026-11-02T02:43:29Z
identity: constitution-alignment-20261102T024329-jay-flowers
tier: draft
---

Issue plans can carry stale constitution-principle counts into deliverables. Issue #713 (openspec 1.14.1 upgrade, filed 2026-10-13) agrees in Decision 2 to fix the unbound-force schema.yaml proposal instruction "all three org constitution principles" by writing FOUR — but the live org constitution (v1.4.0, amended 2026-10-14) defines SIX principles (I-VI, including V Security by Default and VI Deterministic Execution), and "four" matches only the hero constitutions (dewey/gaze/replicator each define four per sibling evidence). Implementing Decision 2 as written would replace one stale count with the next and recreate the defect in the forked schema. When triaging or implementing count-bearing fixes sourced from issues authored before or during a constitution amendment, re-verify the count against .specify/memory/constitution.md at implementation time and prefer count-free phrasing ("all org constitution principles") or runtime derivation from the constitution's Core Principles section. New carrier surface: an issue's "Agreed decisions" section feeding an openspec schema.yaml proposal instruction. Extends constitution-20261023T171911-jay-flowers and governance-drift-20261102T000211-jay-flowers (same pattern in Constitution Check sections and the constitution-check agent).
