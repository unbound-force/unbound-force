---
tag: persona-design
author: jay-flowers
category: pattern
created_at: 2026-11-08T19:49:49Z
identity: persona-design-20261108T194949-jay-flowers
tier: draft
---

A wrapper pinned as a repository's mechanism of record for a tool invocation (Makefile routing + guard test + preflight declaration) is a repo-wide doctrine, but a guard test that reads one exemplar file protects only that exemplar. PR 959 shipped node-direct instruction-lint invocations on three surfaces the wrapper-routing guard test cannot see — the new divisor-determinism persona's Step 1 command and frontmatter bash allow, and AGENTS.md's "# or: node .opencode/lib/instruction-lint.ts" build-command alternative — in the same PR that hardened divisor-linguist to wrapper-only invocation. The wrapper exists precisely because node-direct load failures exit 1 and masquerade as lint findings; the wrapper converts them to a clean exit 2 with the reason on stderr. Authoring rule: before writing any command that invokes a wrapped tool, grep the invocation doctrine (Makefile target, guard tests, sibling personas) and mirror it in BOTH the prose command and the frontmatter bash allow — the two form one contract. Where no wrapper exists for a tool (tool-inventory.ts has none), node-direct is the conforming choice; the doctrine is "wrapper where a wrapper exists", and the asymmetry should be explicit in the frontmatter comments.
