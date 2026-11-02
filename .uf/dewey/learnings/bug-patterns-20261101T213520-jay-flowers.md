---
tag: bug-patterns
author: jay-flowers
category: pattern
created_at: 2026-11-01T21:35:20Z
identity: bug-patterns-20261101T213520-jay-flowers
tier: draft
---

Triage of unbound-force/unbound-force#895 (uf_finale_ci_poll failing with `Unknown JSON field: \"conclusion\"`) surfaced an external-contract transcription failure pattern. The finale-token-reduction origin spec mandates the tool return `failedChecks: [{name, conclusion, url}]`, and the implementation transcribed that field name into its `gh pr checks --json` request — but the installed gh CLI (2.91.0) exposes `bucket`/`state`, not `conclusion`, so gh rejects the field request outright. Meanwhile the host-side fetch_pr_checks repair-lane tool parses `bucket`/`state` correctly: two consumers of the same gh contract drifted to different schema assumptions, and the spec's error taxonomy (GH_CLI_ERROR covering only gh absence/auth failure) never anticipated a valid-gh/invalid-field failure mode. Pattern for triage and review of tools that shell out to external CLIs: distinguish the tool's own output contract (which may keep normalized field names like `conclusion`) from the external CLI's input schema (which must use real field names like `state`/`bucket`); map between them explicitly; pin the external schema assumption in tests so an upstream CLI release breaks CI instead of a production workflow like /uf.finale; and when both a plugin tool and a host tool consume the same external CLI output, prefer one shared parser over two independent field lists.
