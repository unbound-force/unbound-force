---
tag: gaze
author: jay-flowers
category: gotcha
created_at: 2026-11-16T18:21:30Z
identity: gaze-20261116T182130-jay-flowers
tier: draft
as_of: "unbound-force@9ab5fa33ff9e448615d0ec8298ee2149284f8f73"
---

On branch quick/fix-triage-issue-step-4-4-recording-inst (2026-11-16), running `gaze crap --format=text --coverprofile=coverage.out ./...` in this checkout fails its baseline comparison with "15 regressions, 1 new-function violations" because the tracked .gaze/baseline.json is a stale foreign baseline generated under /home/maburgha/GIT/ProdSec/uf-unbound-force: its 523 entries use absolute paths and reference packages absent from this repo (internal/sprint, internal/sync, internal/orchestration, internal/metrics, internal/impediment). The read-only reconciliation is `--baseline-root /home/maburgha/GIT/ProdSec/uf-unbound-force`, which normalizes the absolute paths to module-relative; regenerating the baseline instead would rewrite .gaze/baseline.json and dirty the working tree mid-review. Even after reconciliation the comparison still reports main-vs-baseline drift, so for a branch-causality check on a markdown-only diff, remember Go CRAP scores cannot change when the branch diff contains zero .go files — classify such regressions as pre-existing baseline drift, not branch-caused.
