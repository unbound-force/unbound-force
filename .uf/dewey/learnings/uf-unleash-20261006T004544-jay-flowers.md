---
tag: uf-unleash
author: jay-flowers
category: pattern
created_at: 2026-10-06T00:45:44Z
identity: uf-unleash-20261006T004544-jay-flowers
tier: draft
---

The deterministic dispatch approach using plan_review_dispatch and invoke_agent tools proved far more effective than LLM-orchestrated review council delegation for the /uf.unleash pipeline. The LLM-orchestrated approach (delegating to cobalt-crush-dev agent with review-council instructions) consumed excessive context through the Task tool boundary and required complex prompt engineering that often failed at the dispatch protocol steps. The deterministic approach works reliably: (1) run a shell script to discover agents and compute git context, (2) call plan_review_dispatch with the discovered agents and changed files, (3) invoke each included agent directly via invoke_agent in parallel, (4) consolidate verdicts. This pattern should be used for all future review council invocations within /uf.unleash rather than delegating to an intermediary agent.
