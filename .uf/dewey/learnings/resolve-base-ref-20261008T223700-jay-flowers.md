---
tag: resolve-base-ref
author: jay-flowers
category: pattern
created_at: 2026-10-08T22:37:00Z
identity: resolve-base-ref-20261008T223700-jay-flowers
tier: draft
---

When implementing a resolve_base_ref tool for Git fork environments, a three-level fallback chain (upstream/main, origin/main, main) reliably handles the stale base ref problem. The key insight is that forks have divergent main branches — `origin/main` on a fork may be thousands of commits behind `upstream/main`, causing `git diff main...HEAD` to report phantom files that aren't part of the feature branch. Using a factory function pattern with injectable executor (GitExec type) makes the tool fully unit-testable without real git operations. The closed set of string literal candidates is also a security invariant — it prevents command injection since no user input flows into the execSync call.
