---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:59:36Z
identity: issue-triage-20261102T025936-jay-flowers
tier: draft
---

Triage verification pattern for unbound-force plugin-tool extraction issues: four checks establish VALID. (1) The cited mechanical prose exists in the named command file — grep for the choreography tokens (e.g. "checkout -b", "fix/pr-") rather than trusting line numbers; citations drift as command files evolve (issue #675 cited uf.review-pr.md lines 829-993, but the fix-branch workflow sits at ~700-828 in current HEAD). (2) Judgment boundaries stay in the command per determinism.md DE-006 — human confirmation gates and push decisions remain LLM-controlled while only mechanics move to the tool. (3) Acceptance criteria name the repo's real invariants: the uf-workflow plugin has a dual-copy mirror (.opencode/plugins/uf-workflow/index.ts canonical + internal/scaffold/assets/opencode/plugins/uf-workflow/index.ts embedded, drift-detected), git tools follow the createXxxTool(exec: GitExec, cwd) factory pattern with tests in .opencode/test/uf-workflow.test.ts, and make plugin-test with the 94% coverage floor is the verification path. (4) Sibling issues under the same parent epic (e.g. #659 Phase 2) covering different tools are scope-mates, not duplicates. Minor under-specified parameters (e.g. an optional arg with no described consumer) are plan-time details, not NEEDS-CLARIFICATION blockers, when intent and acceptance criteria are unambiguous.
