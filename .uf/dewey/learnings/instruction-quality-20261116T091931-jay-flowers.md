---
tag: instruction-quality
author: jay-flowers
category: pattern
created_at: 2026-11-16T09:19:31Z
identity: instruction-quality-20261116T091931-jay-flowers
tier: draft
---

Submitted by divisor-linguist during the opsx/per-task-delegation-budgets code review iteration 1 (correlation bd99036f-177a-4bb0-8743-ede5f63db5ce); stranded because dewey_store_learning was unavailable; written to disk at retrospective. When instruction prose adds an override escape hatch to a tool-owned contract (e.g., an explicit `timeout` overriding a derived per-task delegation budget in uf.unleash.md), the authorizing authority must be named unambiguously and in the file's established vocabulary ("the human", "the user"). A single-use undefined term like "the operator" lets an executing agent read itself as the authority and self-authorize the override, silently defeating the tool-owned default established in the same section. Found in opsx/per-task-delegation-budgets review (.opencode/commands/uf.unleash.md:1242); the term appeared exactly once in a 2133-line file whose every other human reference says "the human"/"the user"/"the developer".
