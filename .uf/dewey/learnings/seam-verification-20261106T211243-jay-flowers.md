---
tag: seam-verification
author: jay-flowers
category: gotcha
created_at: 2026-11-06T21:12:43Z
identity: seam-verification-20261106T211243-jay-flowers
tier: draft
---

Seam-evidence tools that detect test coverage by filename convention silently report existing coverage as missing when the providing suite lives at a different filename or entry shape. Verified instance (unbound-force PR #959): tool-inventory.ts seamReport() credits the DE-007 registration layer only to `.opencode/test/registration-surface.test.ts` and extracts plugin ids from `{ id: "..." }` object entries; the change that shipped the tool placed its registration-surface pin inside tool-inventory.test.ts with tuple-shaped plugin entries, so genericPlugins resolves to an empty set for the very repository that ships it — every future plugin-tool change would report `missing: ["registration", ...]` and trigger a false CRITICAL DE-007 finding from the determinism persona. Fix pattern: pin the convention with a repo-level discoverability test — run the real evidence tool against the real repository and assert the coverage it should credit is credited — so the convention between test layout and evidence tool cannot drift silently; when adding convention-keyed detection, ship the convention-named artifact (or derive the convention from what exists) in the same increment.
