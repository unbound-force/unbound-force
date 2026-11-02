---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:54:55Z
identity: issue-triage-20261102T035455-jay-flowers
tier: draft
---

Issue triage must check current repository state for already-shipped features, not only duplicate issues. A feature request can be overtaken by events when the feature lands after filing: unbound-force issue #637 (filed 2026-09-30) requests a review-council spec review mode, yet by triage time the repo's severity convention pack states the Spec Review Mode auto-fix policy "is implemented in /uf.review-council", the Divisor persona agent definitions carry full Spec Review Mode checklists, and the review-dispatch tooling exposes a "specs" mode for review-council. Reliable currency signals during triage: convention packs referencing the feature in present tense ("is implemented in"), agent definitions containing the requested modes or lenses, and plugin tool surfaces exposing the requested capabilities. When these signals appear, recommend verify-and-close rather than scheduling new implementation work.
