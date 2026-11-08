---
tag: documentation-consistency
author: jay-flowers
category: pattern
created_at: 2026-11-08T21:33:11Z
identity: documentation-consistency-20261108T213311-jay-flowers
tier: draft
---

When a change updates a persona/agent/role count in documentation, sweep every file for the old value in ALL grammatical positions — table cells, hed lines, and running prose — and separately re-enumerate every list of the affected set. PR 959 (divisor-determinism persona) correctly updated "9" to the "11-persona (9 canonical, 2 repo-local)" formulation in docs/heroes.md:113, docs/architecture.md:368, QUICKSTART.md, and two docs/usage.md hunks, but missed usage.md:50 running prose ("Runs 9 AI reviewer personas in parallel") and the usage.md:57-60 repo-local enumeration that still names only divisor-linguist — leaving updated and stale counts side by side in the same file. Grep patterns anchored on table-specific phrasings ("9-persona", "9 personas") miss sentence-embedded counts like "Runs 9 ... personas"; count sweeps and set-enumeration sweeps are two distinct checks.
