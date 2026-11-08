---
tag: seam-verification
author: jay-flowers
category: gotcha
created_at: 2026-11-06T21:12:43Z
identity: seam-verification-20261106T211243-jay-flowers
tier: draft
---

Seam-evidence tools that detect test coverage by filename convention silently report existing coverage as missing when the providing suite is renamed or reshaped and nothing pins the coupling. Verified instance (unbound-force PR #959, corrected during review): tool-inventory.ts seamReport() credits the DE-007 registration layer only to `.opencode/test/registration-surface.test.ts` and extracts plugin ids from `{ id: "..." }` object entries. At head the extraction succeeds — the real registration-surface.test.ts registers uf-workflow, uf-finale, review-dispatch, and invoke-agent as object entries that match — but the seam report's own test pinned the extraction only through self-authored scratch fixtures in the same parsed shape, so the coupling to the real suite was unverified: a rename of the suite or a reshape of its plugin entries would silently flip credit to `missing: ["registration", ...]` and trigger a false CRITICAL DE-007 finding from the determinism persona. (An earlier draft of this learning claimed the extraction resolved to an empty set at head; that conflated the tuple-shaped PLUGINS list in tool-inventory.test.ts's runtime-registry section with the object-shaped list in registration-surface.test.ts.) Fix pattern: pin the convention with a repo-level discoverability test — run the real evidence tool's extraction against the real repository's suite and assert the coverage it should credit is credited — so the convention between test layout and evidence tool cannot drift silently; when adding convention-keyed detection, ship the convention-named artifact (or derive the convention from what exists) in the same increment.
