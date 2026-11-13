---
tag: ci-cd
author: jay-flowers
category: gotcha
created_at: 2026-11-13T22:08:13Z
identity: ci-cd-20261113T220813-jay-flowers
tier: draft
as_of: "unbound-force@4558c49ae60be6c02a7f2d23c541518ee50d7fd4"
---

In nightly eval workflows, order the timeout layers deliberately: job timeout > fixture/task timeout + cold-cache setup and artifact-upload headroom. A fixture timeout set at the observed worst-case run time (e.g. 90 min cap against an 89 min max) turns model-variance slowness into red jobs that look like regressions, defeating the regression signal; and if the fixture cap meets or exceeds the job cap, the job is killed before the summary and artifacts are written.
