---
tag: plugin-tool-extraction
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:16:12Z
identity: plugin-tool-extraction-20261102T031612-jay-flowers
tier: draft
---

When adding gh-CLI-backed tools to the uf-workflow plugin (the epic #659 Phase 2 "GitHub API Tools" pattern), the shared gh execution core — runGhCommand/ghFailureError with timeout/rate-limit/api_error classification — lives inside review-dispatch/index.ts, and the shared Bun.spawn test harness lives at .opencode/test/helpers/gh-spawn-mock.ts. The shipped precedent (PR #745, uf.unleash migration) establishes the correct pattern: extract shared execution logic to .opencode/lib/ modules imported by both plugins, and mirror every new plugin/lib asset to internal/scaffold/assets/ with extended drift tests, rather than duplicating gh execution code per plugin. Issues proposing new gh tools to uf-workflow (e.g., #669 ensure_labels/fetch_issue_comments) should plan for that lib extraction in the spec phase to avoid a CS-004 DRY violation.
