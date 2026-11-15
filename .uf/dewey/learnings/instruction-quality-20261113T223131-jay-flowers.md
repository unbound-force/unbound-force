---
tag: instruction-quality
author: jay-flowers
category: pattern
created_at: 2026-11-13T22:31:31Z
identity: instruction-quality-20261113T223131-jay-flowers
tier: draft
as_of: "unbound-force@4558c49ae60be6c02a7f2d23c541518ee50d7fd4"
---

When AGENTS.md documents paired make targets where one runs a subset, the summary must state the subset boundary explicitly. In PR #1061, the Makefile defines `make eval-fast` as `.eval/run.sh --only '^review-council-reject-'` (3 of 5 fixtures) while `make eval` runs every fixture, but the AGENTS.md bullet said "`make eval-fast` (a few minutes) or `make eval` (council fixtures up to 90 minutes each) runs them" — duration parentheticals do not carry scope, the referent "them" reads as the full fixture set, and an agent verifying with the fast target reports a false full-suite green. Pattern: whenever one target filters (--only, a pattern, a file list), name what it runs and what it skips in the always-on summary, not just how long it takes.
