---
tag: seam-verification
author: jay-flowers
category: gotcha
created_at: 2026-11-08T19:24:49Z
identity: seam-verification-20261108T192449-jay-flowers
tier: draft
---

Convention-keyed extraction is pinned only by a real-repo discoverability test, and claimed failure states must be verified against the real repository before being recorded as learnings. Verified instance (PR 959 feedback validation): tool-inventory.ts seamReport() credits the DE-007 registration layer by regex-extracting `{ id: "..." }` entries from the hardcoded path .opencode/test/registration-surface.test.ts; its tests used only self-authored scratch fixtures, so nothing pinned the real file. The committed retrospective learning stamped 'Verified instance' on the claim that genericPlugins 'resolves to an empty set for the very repository that ships it' — false at head: the real file's PLUGINS array matches the pattern and all four plugins are credited. The author conflated the tuple-shaped plugin list in tool-inventory.test.ts with the object-shaped list in registration-surface.test.ts; that same two-shapes-one-concept divergence is the actual drift vector, since harmonizing either file would silently empty the credit set and turn future plugin-tool changes into false CRITICAL DE-007 findings. Fix pattern: in the same increment, run the production extraction against the actual repository tree — export the extraction rather than re-implementing the regex in the test (a re-implemented regex pins nothing), and note a plain seamReport run may report zero seams at head, so exercise the extraction directly or use a historical base ref — asserting the known entries are credited as a stable subset rather than an exact list; never label a learning 'verified instance' without reproducible evidence.
