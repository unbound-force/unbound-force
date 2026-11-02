---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-10-31T02:51:30Z
identity: review-dispatch-20261031T025130-jay-flowers
tier: draft
---

When two code paths must agree on a threshold, hoist both the constant AND the measurement function into one shared pure module. In review-diff-manifest, defining DIFF_THRESHOLD and countDiffLines together in .opencode/lib/review-diff-manifest.ts and importing both into the plugin made the capture-path exceeds_threshold warning fire exactly when manifest delivery engages, eliminating divergence from the legacy split("\n").length count (which counted the trailing empty segment). Boundary tests at exactly 1500/1501 lines pin the shared boundary from both sides. The pattern should extend to MAX_CHUNK_BYTES: interpolate the value into rendered prose (renderDiffManifestSection, tool descriptions) instead of hardcoding "256 KiB" literals so recalibration cannot cause documentation drift.
