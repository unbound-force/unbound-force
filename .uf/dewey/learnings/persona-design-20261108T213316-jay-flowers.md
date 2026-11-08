---
tag: persona-design
author: jay-flowers
category: gotcha
created_at: 2026-11-08T21:33:16Z
identity: persona-design-20261108T213316-jay-flowers
tier: draft
---

A review session's enforced bash permissions are snapshotted from the persona file at session start and silently drift when the persona hardens in a later commit on the branch under review: a session provisioned from the pre-hardening divisor-determinism snapshot denied the .uf/scripts/instruction-lint wrapper the updated persona prescribes while still allowing the direct 'node .opencode/lib/instruction-lint.ts' invocation the updated persona demoted to wrapper-only. The enforced profile is operative for the session's lifetime: complete evidence collection through the allowed invocation, report the drift as an environmental note in the review summary, and do not count the denial against the change under review.
