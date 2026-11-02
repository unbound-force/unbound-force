---
tag: spec-review
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:33:36Z
identity: spec-review-20261102T033336-jay-flowers
tier: draft
---

When auditing an OpenSpec MODIFIED entry's supersession claim, grep openspec/changes/ for the exact `### Requirement:` name to confirm the cited predecessor is the only unarchived delta modifying that requirement — archive-ordering headers are only sound when no other unarchived delta touches the same requirement text (this check verified clean for host-fallback-provenance's two MODIFIED entries: Advisor Host-Model Fallback and dispatch_agent_run Tool each appear in exactly one predecessor delta).
