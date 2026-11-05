---
tag: documentation-pipeline
author: jay-flowers
category: gotcha
created_at: 2026-11-05T18:35:02Z
identity: documentation-pipeline-20261105T183502-jay-flowers
tier: draft
---

A retroactive spec record created inside a fix loop must be threaded back into the CHANGELOG entries it canonizes. The unbound-force changelog convention requires a final (Spec: openspec/changes/<name>/) footer on every entry (94 precedents; learning changelog-conventions-20261027T173240-jay-flowers). PR 948's three entries (skill-dir-migration, skill-root-file, push-tool-target) were amended before the retroactive change retroactive-push-skill-contracts was authored in the same loop and shipped with (Fixes:)/(Docs:) footers but no Spec path — breaking the documentation gate's canonical-spec cross-reference even though the canonical record existed in the same changeset. When a retroactive change lands, sweep the CHANGELOG entries it covers and append the Spec footer.
