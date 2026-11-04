---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-04T15:56:25Z
identity: review-dispatch-20261104T155625-jay-flowers
tier: draft
---

When writing the suite for a bounded section-serving reader like read_dispatch_artifact, enumerate the reader's full section domain (every enum value) and assert served entry content in-bounds, not only past-the-end: the PR #949 suite requested `advisories` never and `findings` only at offset 5 past a 1-entry total, so the selector ternary's advisories branch had no content-selecting test while the paging machinery was well covered via proposals. Two traps compound this pattern: (1) sample payloads often carry empty sections — the review-dispatch sample has advisories: 0 — so a naive "request the section" test against the unmodified sample is vacuous (it asserts another empty page); inject entries first, mirroring existing fixture helpers like proposals(count, length). (2) Assert the reader's own contract — full, untruncated entries (e.g. a consolidated finding with its run_ids) — not the compact/truncated shapes of other projections (consolidate_dispatch's summary retrieval), which describe a different consumer surface. Extends learnings review-dispatch-20261104T143125 (derive the test list from the requirement body's full condition enumeration, not the scenarios) from error branches to section-selector coverage.
