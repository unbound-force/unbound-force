---
tag: documentation-gate
author: jay-flowers
category: reference
created_at: 2026-10-16T11:50:36Z
identity: documentation-gate-20261016T115036-jay-flowers
tier: draft
---

A documentation-gate task marked complete is not verifiable unless the change record cites the filed issue number(s). In persist-unleash-summary (PR 737), task 4.2 said "assess the documentation gate... record the assessment either way" and was checked [x], but neither tasks.md, the CHANGELOG entry, nor the change artifacts named the filed website/repo documentation issues — a reviewer must perform `gh issue` archaeology to confirm the org constitution's Cross-Repo Documentation MUST was satisfied. Reliable pattern: when a documentation-gate task is marked [x], append the concrete outcome to the task text itself (e.g., "filed unbound-force/website#NNN and unbound-force/unbound-force#NNN" or "exemption claimed: <reason>"), so the gate is auditable directly from the diff. Task wording should name the target repository at authoring time and the completion should cite the artifact number at check-off time; both omissions independently make the gate unauditable.
