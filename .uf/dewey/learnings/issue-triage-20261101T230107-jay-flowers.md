---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T23:01:07Z
identity: issue-triage-20261101T230107-jay-flowers
tier: draft
---

Documentation-gate supersession should be bidirectional. When a new change's docs-gate issue supersedes an older issue's behavioral assessment for the same command file — as unbound-force/unbound-force#838 (update-run-pipeline-plan) does for #747's "three-subagent pipeline, gate relay" assessment of /uf.run-pipeline-plan — the superseding issue documents the relationship on its own body, but the superseded issue (which stays open for a distinct follow-up) still presents the stale assessment with no forward pointer, and the originating change's spec delta (REMOVED Requirements: Gate Relay Protocol) then contradicts what a reader of the older issue sees. When triaging or filing a superseding docs-gate issue, add a cross-reference comment on the superseded issue linking to the superseder, so neither record misleads in isolation.
