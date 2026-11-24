---
tag: github-actions
author: jay-flowers
category: gotcha
created_at: 2026-11-24T14:34:59Z
identity: github-actions-20261124T143459-jay-flowers
tier: draft
as_of: "unbound-force@12bb5f8799eea38393959349f27354e1bb031a77"
---

GitHub Actions required-check mechanics that shaped the ci-skip-learnings-only-prs design: a job skipped by a job-level `if:` satisfies a required check of the same name — except matrix jobs (they skip with the matrix expression unexpanded, so per-leg check names never report) and reusable-workflow calls (a skipped caller reports only the caller's job name, so nested checks like "Standardized CI / Run linters" never report). A workflow-level `paths-ignore` on a workflow that provides a required check leaves that check permanently unreported ("Expected — Waiting for status to be reported"). The stable required-check surface is one aggregate job per workflow: `if: always()`, `needs` the detection job plus every gated job, failing on any needed-job conclusion other than success or skipped.
