---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:35:35Z
identity: issue-triage-20261102T023535-jay-flowers
tier: draft
---

In spec-governed repositories, triage a migration-tracking issue's workflow-tier routing before assessing validity: scope that is cross-repo or exceeds three user stories must be routed through a strategic spec workflow (Speckit tier per AGENTS.md), with the issue retained only as an execution tracker under an umbrella issue. Where an issue body's figures go stale (e.g., dependency pin versions), a superseding comment is the correction pattern — treat the comment thread as authoritative and verify pin claims against the actual package.json before flagging inconsistency.
