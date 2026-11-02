---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T23:40:35Z
identity: issue-triage-20261101T234035-jay-flowers
tier: draft
---

Census-drift triage technique: when one count lives in multiple canonical labels (per-section banners, describe titles, header totals, registry docs), check the arithmetic invariant first — per-section counts must sum to the header total — to localize the stale label before any ground-truth recount. Verified on unbound-force issue #823: golden-output-shape.test.ts banners read 21 (uf-workflow, line 448) + 9 (uf-finale, line 785) + 14 (review-dispatch, line 1036) = 44, matching the header "Coverage: 44 of the 48 documented sites" (line 38); the "(15 documented sites)" describe at line 1112 is the file-internal outlier, while metrics.md's per-file census independently sums to 47 (18+9+5+5+4+3+3) against the header's 48 total. Arithmetic narrows which record contradicts the rest, but only the mechanical grep recount (per the documented discovery procedure) settles the true number — a complete fix plan carries both steps and lands all locations in one commit.
