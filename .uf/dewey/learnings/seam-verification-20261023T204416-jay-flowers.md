---
tag: seam-verification
author: jay-flowers
category: gotcha
created_at: 2026-10-23T20:44:16Z
identity: seam-verification-20261023T204416-jay-flowers
tier: draft
---

Hard-coded `uf.*.md:<line>` site labels live in three locations, not two: golden-output-shape.test.ts, dispatch-review-runs.test.ts, and the seam-verification site registry (openspec/changes/seam-verification/metrics.md — fixture inventory, exclusion record, per-command counts, task-2.2 attribution), which the golden suite treats as its census of record. A line-shifting edit to a command file must recompute labels in all three — grep the repo for `<command-file>.md:` before finishing; the registry needed the identical 314→339 / 338→363 / 473→541 / 733→801 mapping the test files received. Adding a new documented tool-call site is also a census event: register it in the fixture inventory, bump the per-command and discovered/pinned totals, and attribute its execution-path test, or the census-diff property ('zero qualifying examples absent from the list') silently goes false. Staleness compounds — a ~42-line drift had accumulated in both the test labels and the registry refs before the address-feedback-diff-path-producer change (2026-10) fixed only the test files.
