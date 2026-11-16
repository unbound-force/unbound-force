---
tag: instruction-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-16T11:04:20Z
identity: instruction-quality-20261116T110420-jay-flowers
tier: draft
as_of: "unbound-force@1d0e5cef2b10545a779afe8aa2c925fb2ab22f9e"
---

In review-council dispatches where the working tree sits at the PR base SHA (ci-aware pre-flight skips local execution and the head commit is absent from the local object store), `instruction-lint --json --base <base_sha> --changed` exits 0 with `delta.files: []` — no evidence about the change. The working degradation is a whole-repo run without `--base`/`--changed`, filtered to instruction files named in the caller's diff manifest, with two adjustments: counts and line anchors are base-state measurements (cross-check the PR's DE-004 CHANGELOG record or the diff for head-state numbers), and every line-anchored finding must be re-anchored through the diff hunks before reporting, because changed files shift line numbers between base and head (e.g., base 1343-1345 became head 1345-1347 in uf.address-feedback.md).
