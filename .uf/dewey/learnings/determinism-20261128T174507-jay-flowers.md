---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-28T17:45:07Z
identity: determinism-20261128T174507-jay-flowers
tier: draft
as_of: "unbound-force@444cb92a90a02f80a8798f4de1152ff133d7abe4"
---

unbound-force PR #1240 (ci/pr-scope) applies DE-001/DE-005 below the instruction layer, to CI config: three workflows' duplicated inline change-detection jobs (checkout + in-job detector test + output-normalization bash + GITHUB_OUTPUT heredoc, two of them byte-identical) collapsed into one SHA-pinned org reusable workflow (jflowers/ship pr-scope.yml@3b0ce8b), with the gated jobs and the required aggregate jobs (CI OK, Standardized CI OK) consuming its typed output (outputs.code) through fail-closed mapping (true requires gated jobs to be success, false requires skipped, unexpected values fail the job). The repo-local detector script and its 24-scenario test were deleted together — the tool's tests moved with the tool to the owning repo. Pattern: when a fixed procedure is duplicated across consumers, extract it to a single pinned tool and gate consumers on its typed output rather than re-implementing the logic at each call site.
