---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:16:38Z
identity: issue-triage-20261102T031638-jay-flowers
tier: draft
---

Triage of prose-to-tool extraction issues (Constitution Principle VI / determinism pack): verify the issue's line and size claims against the CURRENT file before ruling. Issue #668 cited uf.finale.md lines 625-953, but earlier increments (finale-token-reduction) had already extracted the git operations to tools, leaving a 399-line file whose 7b Conflict Recovery section is ~52 lines of orchestration prose. Stale estimates do not invalidate the issue: the DE-005 violation — judgment-free tool-chaining prose that MUST collapse into a single intent-level tool — is size-independent. Security-domain validity checklist for destructive git-path extraction issues: acceptance criteria must name the preserved guards (force-with-lease, never bare --force; merge/rebase abort rollback paths), tests must isolate git execution through the injectable GitExec seam in uf-workflow-types.ts (repo convention is Bun.spawn argument arrays, not BunShell template literals — review-dispatch/index.ts documents this to prevent injection), and free-string parameters reaching subprocess execution (branch, base_branch, remote) need Constitution Principle V validation requirements carried into the spec phase.
