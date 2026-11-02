---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:57:53Z
identity: determinism-20261102T025753-jay-flowers
tier: draft
---

In unbound-force/unbound-force, deterministic file-path classification (path regexes → categories including security/cli-ux/documentation/ci-cd/dependencies, plus security_sensitive, user_facing, and tier derivation) already exists as the change-profile computation inside .opencode/plugins/review-dispatch/index.ts. New classification consumers (e.g., issue #673's proposed classify_paths tool) MUST NOT re-implement the patterns — two classifier implementations drift. The correct shape: extract the existing classifier to .opencode/lib/ as a shared pure function (precedent: classifyBranch, matchTaskLine, sanitizeSlug), then expose a thin wrapper over it or fold it into whichever tool needs it.
