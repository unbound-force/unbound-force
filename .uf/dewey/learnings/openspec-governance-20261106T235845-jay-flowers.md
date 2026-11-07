---
tag: openspec-governance
author: jay-flowers
category: gotcha
created_at: 2026-11-06T23:58:45Z
identity: openspec-governance-20261106T235845-jay-flowers
tier: draft
---

Retroactive spec records created late in a council fix loop close the spec-first gap for already-shipped behavior, but CHANGELOG entries amended in earlier rounds of the same loop do not automatically gain their (Spec:) pointers — traceability then ends at the Fixes/Docs issue references instead of the canonical spec. When a retroactive OpenSpec change is the accepted remedy for a tier-routing finding, its CHANGELOG task should include back-filling `(Spec: openspec/changes/<name>/)` onto the entries it canonizes. Observed on unbound-force PR 948: retroactive-push-skill-contracts restates the skill-dir-migration/skill-root-file/push-tool-target contracts, but those CHANGELOG entries (written before the record existed) point only at issues #933/#934/#947 and the docs issues.
