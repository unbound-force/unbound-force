---
tag: command-contract-tests
author: jay-flowers
category: pattern
created_at: 2026-11-05T16:15:29Z
identity: command-contract-tests-20261105T161529-jay-flowers
tier: draft
---

A PR carrying multiple prose→tool migrations must apply the command-contract pin convention uniformly — pinning the migration that drew round-1 feedback while shipping its sibling unpinned leaves a silent-regression window that presence-only tool-name clauses and canonical↔embedded drift tests cannot catch (drift tests verify byte-identity between copies, not that the content carries the tool contract). When adding consumer-side pins for a tool-contract migration, derive them by census, not from the feedback's named spans: enumerate every tool mention in the rewritten commands, classify call sites vs guardrails/prose (a NEVER-guardrail legitimately naming a raw `git push --force` form false-positives a naive command-wide substring ban; use exact-form or per-span/between() assertions instead), and pin each call site's parameter shape per command — parameters deliberately differ by command, so a uniform parameter pin either false-fails or under-pins.
