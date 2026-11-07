---
tag: scaffold-testing
author: jay-flowers
category: pattern
created_at: 2026-11-06T20:03:05Z
identity: scaffold-testing-20261106T200305-jay-flowers
tier: draft
---

When a spec MUST is implemented as a small wiring call inside an orchestrating function (e.g., scaffold.Run() calling preservePluginDevDependencies only when the target .opencode/package.json already exists), unit tests of the extracted helper leave the call site unverified: the relPath match, the existence gate, and the write-group reporting can all regress while every test stays green. Pair the helper's unit tests with one orchestrator-level test that pre-populates exactly the target state the wiring gates on (an existing manifest declaring devDependencies) and asserts the deployed bytes and the Created/Updated/skip reporting.
