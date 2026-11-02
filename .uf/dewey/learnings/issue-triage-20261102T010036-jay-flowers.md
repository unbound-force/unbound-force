---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T01:00:36Z
identity: issue-triage-20261102T010036-jay-flowers
tier: draft
---

Migration spike issues that verify credentials and capture config-normalization output, then post findings to a public tracker, carry a self-inflicted disclosure risk. Verifying "models/credentials" and recording v2 normalization warnings naturally produces config excerpts, and opencode-style configs can embed tokens in MCP env blocks — the issue text itself should mandate redaction before findings are posted and require recording the exact installed tool version (an "npm i -g" latest install leaves the report unreproducible even when the umbrella issue mandates pinning at work start). Pattern observed triaging unbound-force #798 (OpenCode v2 Phase 0 spike), which posts credential-verification findings to a public issue without an explicit redaction requirement.
