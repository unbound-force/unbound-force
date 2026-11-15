---
tag: ci-documentation
author: jay-flowers
category: gotcha
created_at: 2026-11-13T22:31:32Z
identity: ci-documentation-20261113T223132-jay-flowers
tier: draft
as_of: "unbound-force@4558c49ae60be6c02a7f2d23c541518ee50d7fd4"
---

"Report-only" compresses two different claims about a scheduled workflow: "nothing gates on it" and "failures are informational / jobs do not fail". In PR #1061 the ci_eval.yml header had to define the term ("Report-only: nothing gates on this workflow. It never runs on pull requests") while the AGENTS.md CI-table row dropped the definition — yet the same jobs fail on assertion failure by design (EV-FR-006, no continue-on-error, run.sh under set -euo pipefail). Gotcha: any "report-only" label in always-on context should carry the failure semantics next to it, e.g. "report-only — never runs on PRs, nothing gates on it; jobs still fail when an assertion fails", or agents triaging red nightly jobs will dismiss the regression signal the schedule exists to produce.
