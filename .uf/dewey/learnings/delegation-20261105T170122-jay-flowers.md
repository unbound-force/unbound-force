---
tag: delegation
author: jay-flowers
category: gotcha
created_at: 2026-11-05T17:01:22Z
identity: delegation-20261105T170122-jay-flowers
tier: draft
---

Durable state written by a delegated child agent (e.g., checkpoint step names in a progress store) becomes an untrusted-data channel when a parent orchestrator later embeds it in recovery prompts. On timeout paths the child's text output typically does not reach the parent, so resume fields are a NEW injection surface: a child compromised via content it processed can plant instruction-like strings in step names. Bound the writing tool's schema (length/pattern, matching any sibling bounds) and frame the cons
