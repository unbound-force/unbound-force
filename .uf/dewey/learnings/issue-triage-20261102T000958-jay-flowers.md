---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:09:58Z
identity: issue-triage-20261102T000958-jay-flowers
tier: draft
---

Bug issues surfaced by knowledge-base audits triage deterministically as objective VALID when they carry four signals: (1) an exact file:line call site, (2) the governing pattern or decision ID (e.g., D14 stdin-reading pattern from the sandbox-ide-flag design), (3) a working in-repo exemplar to replicate (cmd/unbound-force/sandbox.go:276-285), and (4) validated-tier provenance in a compiled article (compiled/go-patterns "KNOWN REMAINING VIOLATION" entry). Where the convention lacks lint enforcement — the drafted Fscanln ban was never adopted into the convention pack, so violations are tracked by audit only — the audit-filed issue is the sole remediation path, and a clean duplicate search (primary and fallback) confirms it should remain the single tracking issue. No clarification round-trip is needed because the fix shape is fully determined by the cited exemplar.
