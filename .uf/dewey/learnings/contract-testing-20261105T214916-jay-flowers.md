---
tag: contract-testing
author: jay-flowers
category: pattern
created_at: 2026-11-05T21:49:16Z
identity: contract-testing-20261105T214916-jay-flowers
tier: draft
---

When a command-file fix establishes a tool-call invariant across multiple instruction sites — the main-flow section, the session-resume guard, and companion tracking/TodoWrite sections — pin each corrected site's conditional wording with its own phrase-presence contract clause, in addition to the invariant statement and any ordering pin. The ordering pin's anchor must skip legitimately-early mentions (a resume guard references the governed tool before the gated section while ordering the actual call after i
