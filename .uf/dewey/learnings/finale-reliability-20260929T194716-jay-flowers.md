---
tag: finale-reliability
author: jay-flowers
category: pattern
created_at: 2026-09-29T19:47:16Z
identity: finale-reliability-20260929T194716-jay-flowers
tier: draft
---

When a confirmation gate uses a tool that renders only summary content (like a question tool with short option labels), add an explicit VISIBILITY DIRECTIVE requiring the agent to print the full proposed text as plain assistant output BEFORE invoking the tool. This ensures the complete text appears in the transcript regardless of context compression or sub-task boundaries. The directive must use RFC 2119 language (MUST) and be placed immediately before the tool invocation. Applied to /uf.finale Steps 3 and 5f, and /uf.review-council Step 7f.
