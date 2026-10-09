---
tag: testing-patterns
author: jay-flowers
category: pattern
created_at: 2026-10-08T22:37:13Z
identity: testing-patterns-20261008T223713-jay-flowers
tier: draft
---

The uf-workflow plugin test file (.opencode/test/uf-workflow.test.ts) uses a mock injection pattern for testing git subprocess calls. Tests create a mock executor function that matches the GitExec type signature (ref: string, cwd?: string) => string, and the factory function createResolveBaseRefTool(exec, cwd) accepts this mock. Key test scenarios: upstream resolves (first candidate succeeds), origin fallback (upstream throws, origin succeeds), main fallback (both upstream and origin throw), all-fail error path, empty string skip (candidate returns empty string, treated as failure and tries next), and cwd parameter passing verification. All 6 tests passed with 100% branch coverage on the resolve_base_ref tool.
