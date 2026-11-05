---
tag: content-pipeline
author: jay-flowers
category: pattern
created_at: 2026-11-04T22:49:33Z
identity: content-pipeline-20261104T224933-jay-flowers
tier: draft
---

In the unbound-force meta repo, CHANGELOG entries for changes that satisfied the documentation gate conventionally end with cross-references to their gate issues — a \"(Docs: #N, unbound-force/website#N)\" pair alongside \"(Spec: openspec/changes/<name>/)\". A new Changed entry citing only (Spec: ...) with no (Docs: ...) pair is a cheap, local textual signal that the change's documentation-gate issues (in-repo docs-labeled + unbound-force/website) are unfiled: check the entry tail and the change's tasks.md for 
