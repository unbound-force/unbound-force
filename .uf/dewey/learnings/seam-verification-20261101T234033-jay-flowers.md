---
tag: seam-verification
author: jay-flowers
category: gotcha
created_at: 2026-11-01T23:40:33Z
identity: seam-verification-20261101T234033-jay-flowers
tier: draft
---

Census-drift triage (unbound-force issue #823): when the three documented-site census records disagree — golden census comment 14 vs describe label 15 vs metrics.md registry 47 vs suite header 48 — arbitrate from the command files, never from any of the disagreeing records. The describe label matched reality: the golden review-dispatch section tests 15 sites because capture_review_diff [uf.unleash.md:620] has a golden test (golden-output-shape.test.ts line 1386) that neither the census comment's 14-site enumeration nor the metrics.md registry (47 total; uf.unleash.md counted at 18, excluding capture_review_diff) records — coverage ran ahead of the census of record. The drift also exceeded the issue's two cited numbers: metrics.md's registered-tool attribution says uf-finale 8 sites vs the suite's 9-site uf-finale section, and its drift baseline says the suite "covers the other 43" vs the header's "44 of the 48". A census recount must therefore sweep every count in all three locations (per-plugin attribution, per-command-file totals, discovered/pinned/uncovered line, suite header, census comments, describe labels), not just the pair a report cites, and land them in one commit per the census discipline.
