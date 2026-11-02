---
tag: go-patterns
author: jay-flowers
category: gotcha
created_at: 2026-11-02T00:09:55Z
identity: go-patterns-20261102T000955-jay-flowers
tier: draft
---

When a pattern-migration change replaces a banned construct at one call site, sibling call sites can survive in other files of the same package. In unbound-force, the sandbox-ide-flag change (design D14) replaced fmt.Fscanln with bufio.Scanner in the Destroy confirmation (cmd/unbound-force/sandbox.go:276-291) but the Extract patch-apply confirmation at internal/sandbox/sandbox.go:793 kept fmt.Fscanln; the straggler surfaced only via the 2026-10-25 knowledge-base audit and the go-patterns compiled article's "KNOWN REMAINING VIOLATION" marker (issue #822). At fix time, run a repo-wide grep for the banned construct instead of fixing only the in-scope call site — the grep is cheap, and without it the migration depends on later audits to catch stragglers.
