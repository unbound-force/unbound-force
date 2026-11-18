---
tag: ci-observability
author: jay-flowers
category: gotcha
created_at: 2026-11-18T20:08:06Z
identity: ci-observability-20261118T200806-jay-flowers
tier: draft
as_of: "unbound-force@cd2520c42c9cd75797136ab6ba23bbdd6ca6bc29"
---

A CI step that can fail on its first statement must create its output directory (or guard the summary/artifact pointer) before anything that can fail. When a step's only artifact-producing work sits behind `mkdir -p "$OUT"` and a newly added preflight command is placed ahead of it, the `if: always()` summary step keeps pointing operators at `run.log in the artifact` while the upload step finds no files, so the recovery instruction references a file that does not exist in precisely the new failure case.
