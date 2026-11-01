---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-01T21:01:22Z
identity: review-dispatch-20261101T210122-jay-flowers
tier: draft
---

Manifest-delivered diff reviews (RD-FR-032) can verify canonical-to-embedded byte-identity cryptographically from the diff itself: compare a modified canonical file's post-image blob hash in its `index <old>..<new>` line against a new-file mirror's `index 0000000..<new>` line — equal hashes prove the pair byte-identical without reading either file. In the model-matrix-scaffold-deployment review (issue #889) both deployed pairs verified this way (meta-prompt.md -> 44803a20, meta-prompt-custom.md -> 8788256f), with the repo's drift tests as primary enforcement and the hash match as independent confirmation. Caveat: hash equality proves identity only for the diff's snapshot; it does not replace drift tests for detecting future divergence. (Lesson proposal from divisor-guard, #889 review-council dispatch.)
