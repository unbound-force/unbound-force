---
tag: ci-cd
author: jay-flowers
category: gotcha
created_at: 2026-11-13T22:31:36Z
identity: ci-cd-20261113T223136-jay-flowers
tier: draft
as_of: "unbound-force@4558c49ae60be6c02a7f2d23c541518ee50d7fd4"
---

A spec requirement of the form "each job MUST set a timeout" is easy to satisfy only for the expensive job and miss for the helper job. In `.github/workflows/ci_eval.yml` (PR #1061, EV-FR-005) the 120-minute timeout landed on the per-fixture matrix job while the `fixtures` list job that only checks out and runs `find` had none, inheriting GitHub's 360-minute default. Review practice: enumerate every job declaration in a new workflow and confirm each carries `timeout-minutes`, not just the one whose runtime the design measured; a helper job can still hang on checkout or step setup.
