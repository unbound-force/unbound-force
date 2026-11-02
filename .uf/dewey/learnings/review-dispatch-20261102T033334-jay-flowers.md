---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:33:34Z
identity: review-dispatch-20261102T033334-jay-flowers
tier: draft
---

When a change conditions or scopes a tool's default behavior, audit the registered tool parameter description strings (the MCP/zod descriptions that form the tool-registry discovery surface) alongside internal code doc comments — \"input schema unchanged\" does not mean \"description text still accurate\". In the host-fallback-provenance change (RD-FR-037), scoping the fix as \"no schema changes\" left dispatch_agent_run's registered tier description (\"Defaults to \"standard\" when neither is provided\") conditionally inaccurate for source: host entries, while the spec deltas and the tasked internal doc comments carried the exception. Tool descriptions live in the plugin file that mirrors byte-identical into the scaffold asset, so the update is a one-line task extension — but only if the spec tasks it.
