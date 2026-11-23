---
tag: instruction-quality
author: jay-flowers
category: pattern
created_at: 2026-11-16T01:48:53Z
identity: instruction-quality-20261116T014853-jay-flowers
tier: draft
---

Tool-extraction migrations orphan prose written for the hand-composed flow. When a command step moves payload composition into a tool (fixed fields, enum-constrained values, additionalProperties: false), surviving instructions like "label it explicitly" lose their object — the payload the agent no longer writes. Audit the migrated section for verbs whose object was the hand-composed artifact; relocate those instructions to the session-report layer or delete them.
