---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:17:00Z
identity: issue-triage-20261102T021700-jay-flowers
tier: draft
---

Triage pattern for deferred-experimental tracking issues with actionable prerequisites, verified on unbound-force issue #730 (inference gateway): the issue is parked ("Experimental... Priority may change — tracking for when ready") yet its first step — upgrading four divergent Fullsend pins (v0.47.1 / v0.43.0 / v0.38.0 / v0.38.0, gateway route needs v0.47.0+) — is independently valuable, crosses a breaking release (v0.46.0 removed per-org installs and requires explicit inference.auth, with vertex-wif preserving current behavior), and is actionable now. Recommend splitting such prerequisites into their own issue so they proceed immediately instead of inheriting the parent's deferral; verify the prerequisite has standalone value first. Companion signal: when a member comment supersedes body steps — especially destructive ones like secret deletion that would break Vertex agent runs — flag the stale body for editing, because an implementer acting on the body alone can execute superseded steps.
