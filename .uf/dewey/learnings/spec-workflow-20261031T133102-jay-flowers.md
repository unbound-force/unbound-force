---
tag: spec-workflow
author: jay-flowers
category: gotcha
created_at: 2026-10-31T13:31:02Z
identity: spec-workflow-20261031T133102-jay-flowers
tier: draft
---

A mid-workflow amendment that changes a fact restated elsewhere in the change's artifacts requires sweeping the whole change directory, not just the artifact being amended. In forward-aggregate-correlation-id (unbound-force), tasks.md 5.4 added constitution principles V and VI post-council and updated proposal.md's assessment to six principles, but design.md's Context narrative still read "all four principles PASS" — contradictory counts inside one change's artifact set. Before closing an amendment task, grep the change directory for restatements of the amended fact (e.g. "principles PASS", "four principles") and update every hit.
