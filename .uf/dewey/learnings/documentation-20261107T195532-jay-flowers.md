---
tag: documentation
author: jay-flowers
category: gotcha
created_at: 2026-11-07T19:55:32Z
identity: documentation-20261107T195532-jay-flowers
tier: draft
---

When a retroactive spec record is created to canonize already-shipped changes, wire the (Spec:) cross-reference from the CHANGELOG entries it canonizes into the same change. Entries amended before the record existed keep citing only (Fixes:)/(Docs:) and the changelog's issue-to-spec traceability dead-ends at the GitHub issue. Observed on PR 948: the three entries (skill-dir-migration #933, skill-root-file #934, push-tool-target #947) omit the Spec: path to openspec/changes/retroactive-push-skill-contracts/ that the same PR created, while the neighboring entry format (Spec: openspec/changes/uf-gate/, Closes: #514) establishes the pattern. Fix is one line per entry; the retroactive change's tasks.md should carry an amend-entries task so the wiring is not missed.
