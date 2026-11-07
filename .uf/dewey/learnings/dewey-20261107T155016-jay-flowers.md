---
tag: dewey
author: jay-flowers
category: pattern
created_at: 2026-11-07T15:50:16Z
identity: dewey-20261107T155016-jay-flowers
tier: draft
---

Standalone Dewey learning-batch commits can be mechanically audited for silent writer truncation before merge: strip the diff '+' prefix and measure each body's content length. A cluster of bodies ending mid-word at one uniform cap length (unbound-force PR #968, chore/dewey-learnings: 14 of 50 bodies at exactly 512 content characters, one at 511, against 35 complete bodies of 568-1376 chars ending in sentences) identifies a fail-open cap in the export path — and because every learning file is +10/-0 (frontmatter plus one body line), the diffstat hides the cut. Pre-merge gate: terminal-punctuation or byte-equality against the Dewey store source for every body in the batch. The presence of a gotcha learning documenting the defect does not substitute for the check — PR #968 shipped 15 new truncated instances while carrying PR #949's truncation gotcha (dewey-20261104T225833) in the same batch, itself truncated at 'changelog-accuracy-20261'.
