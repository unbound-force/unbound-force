---
tag: gatekeeping
author: jay-flowers
category: gotcha
created_at: 2026-11-07T20:23:20Z
identity: gatekeeping-20261107T202320-jay-flowers
tier: draft
---

An advisory pre-flight check needs its advisory-to-required escalation tracked with an owner, not only documented. PR #958's instruction-lint check records the tightening criterion ("advisory until the IL-010 error and warning backlog are cleared") in design.md and the .uf/preflight.yaml comment, but the trigger depends on cleaning a pre-existing baseline error in a file the change never touches, and no issue schedules that cleanup or the tightening review — unlike the change's other accepted residuals, which are tracked in an issue (#971). A documented-but-unowned deferral condition drifts: pair every advisory gate declaration with a tracking issue that names the tightening criterion and the baseline cleanup it depends on, mirroring the accepted-residuals pattern.
