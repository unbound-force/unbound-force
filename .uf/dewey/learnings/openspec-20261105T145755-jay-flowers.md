---
tag: openspec
author: jay-flowers
category: gotcha
created_at: 2026-11-05T14:57:55Z
identity: openspec-20261105T145755-jay-flowers
tier: draft
---

When an OpenSpec MODIFIED delta replaces a requirement whose accumulated contract spans multiple unarchived predecessor deltas, diff the new requirement text against the LATEST unarchived predecessor delta, not the original change, and enumerate every MUST clause being dropped — orthogonal concerns ride along with the mechanism being removed. In PR 957 (stop-shipping-plugin-lock), removing the lock silently dropped the MIT-license retention clause that lived in clarification-marker-superset's SC-FR-003 alongside the lock-integrity clause, while the version pins carried forward; the proposal's supersession note cited only review-council-multi-model-fanout (the original change), and neither design.md nor the Impact section recorded the drop. Cheap guard: compose the MP-013 Previously paragraph by mechanically listing each MUST in the predecessor text and marking it carried or dropped — the orphaned clause surfaces immediately. Related implementation gotcha: when a verification gate moves from a live artifact to a static provenance record, the record-based test can go silently stale on pin bumps unless its version expectations are cross-checked against the manifest (the deleted lock-entry assertions were manifest-linked by npm ci; the surviving plugin-package-provenance.json test is not).
