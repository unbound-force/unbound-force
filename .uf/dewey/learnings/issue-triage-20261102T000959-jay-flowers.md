---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T00:09:59Z
identity: issue-triage-20261102T000959-jay-flowers
tier: draft
---

Pattern-migration design decisions scoped to a single call site silently leave sibling call sites un-fixed in other layers. In unbound-force, design decision D14 (sandbox-ide-flag change) replaced fmt.Fscanln with bufio.Scanner for the destroy confirmation in cmd/unbound-force/sandbox.go only; the Extract confirmation at internal/sandbox/sandbox.go:793 survived the fix because task 6.4 named exactly one function. The violation persisted for months and was caught only by a knowledge-base audit (compiled go-patterns, 2026-10-25), since no lint rule enforces the Fscanln ban. Cheap mitigation when landing any "replace X with Y" decision: grep the banned symbol repo-wide in the same change and either fix or explicitly ticket every remaining call site.
