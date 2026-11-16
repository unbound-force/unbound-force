---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-15T16:30:39Z
identity: determinism-20261115T163039-jay-flowers
tier: draft
---

Intent-level gate-extraction pattern, shipped as assess_prior_review_staleness (issue #1055, PR #1077): when a command gate needs mechanical classification, move selection (latest standing review by submitted_at), retrieval (paginated gh reads), mapping (verdict→review-state via a constant table), and comparison (commit_id vs the FROZEN head_sha from the caller's context, never the live PR) into a registered plugin tool with a zod-strict input schema; keep only the materiality judgment and the human gate in prose, and state the DE-006 boundary explicitly in both the tool description and the consuming prose ("the tool deliberately does not decide materiality"). Judgment products — the verdict this run will post — arrive as validated parameters; prose must not restate the tool's routing ("do not restate, recompute, or override it"). Seam verification for the gh-fetch tool family lives in per-tool test files that execute the registered tool against mocked gh and assert the documented envelope, NOT in golden-output-shape.test.ts — whose fixed site registry has never covered resolve_review_mode, fetch_pr_checks, or fetch_pr_review_context — so tool-inventory's filename-based classifier reporting missing ["output-shape"] for these tools is a taxonomy artifact, not a coverage gap; round-trip is N/A for read-only tools. Pin the migrated gate structure (option-array literals, tool-ownership directive, untrusted-input marker) with span-scoped pins in scaffold_test.go so the prose cannot drift.
