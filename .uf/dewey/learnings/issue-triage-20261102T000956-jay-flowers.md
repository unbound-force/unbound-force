---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T00:09:56Z
identity: issue-triage-20261102T000956-jay-flowers
tier: draft
---

When triaging convention-violation issues sourced from knowledge-base audits, verify "last remaining violation" claims with a repo-wide grep but distinguish live call sites from comment mentions. In unbound-force, grepping for Fscanln returns two matches: internal/sandbox/sandbox.go:793 (the executable violation) and cmd/unbound-force/sandbox.go:277 (a comment inside the fixed D14 exemplar that names fmt.Fscanln while explaining why it is banned). Counting string matches instead of call sites would misreport the violation count and could wrongly challenge the issue's "last remaining" claim. Read each grep hit's line context before confirming or rejecting such claims.
