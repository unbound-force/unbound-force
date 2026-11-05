---
tag: documentation-gate
author: jay-flowers
category: pattern
created_at: 2026-11-05T17:48:45Z
identity: documentation-gate-20261105T174845-jay-flowers
tier: draft
---

A checked-off documentation-gate task is auditable only when it records what was filed. In one PR, the preflight change's task 4.4 recorded "Filed unbound-force/unbound-force#951; extended unbound-force/website#373" while the sibling consumer-bounds task 5.3 was marked [x] with no issue references at all — the same PR contained both the practice and its absence. Because the AGENTS.md documentation gate is a MUST rule and reviewers may lack GitHub access, record the filed issue numbers (or a pointer to the PR-body section listing them) inside the task text, so the completion marker leaves an in-repo evidence trail instead of an unverifiable claim.
