---
tag: openspec-integration
author: jay-flowers
category: gotcha
created_at: 2026-11-02T02:43:31Z
identity: openspec-integration-20261102T024331-jay-flowers
tier: draft
---

Hardcoded enumerable counts cascade into stale governance artifacts across generator eras: unbound-force's openspec schema.yaml instructs assessment against "all three org constitution principles" (schema.yaml:15-20), openspec/config.yaml's context block says "Four core principles" (listing I-IV only), while the live org constitution v1.4.0 has six principles — and issue #713 (filed 2026-10-13, one day before the 2026-10-14 Principle VI amendment; the constitution already had five principles at filing) then propagated the config.yaml count ("the constitution has FOUR principles") into its agreed decisions. Verified via grep of .specify/memory/constitution.md principle headings (lines 58/85/113/141/162/199) against schema.yaml and config.yaml. Pattern: generated instructions should not hardcode counts of enumerable sets (principles, workflows, files) when the source of truth is a readable file — instruct consumers to read the source, as schema.yaml already does ("Read the constitution from .specify/memory/constitution.md"), and re-verify counts at every regeneration or generator upgrade.
