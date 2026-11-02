---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:16:46Z
identity: issue-triage-20261102T031646-jay-flowers
tier: draft
---

Prose→tool extraction issues (Constitution Principle VI migrations) validate fastest when they carry four markers: (1) the mechanical block identified with line ranges in the host command file, (2) an explicit judgment-at-the-boundary split naming what stays LLM/human-orchestrated, (3) measurable line/token deltas per determinism pack DE-004, and (4) testable acceptance criteria including mocked-shell tests and explicitly preserved safety checks. Issue #668 (resolve_conflict tool replacing uf.finale.md conflict-resolution prose) is the canonical example — each marker maps to a determinism-pack rule (DE-001 mechanical extraction, DE-005/DE-006 intent-level boundary, DE-004 efficiency evidence, DE-007 test seed), letting triage confirm validity by rule citation rather than judgment. Watch for two spec-phase obligations such issues defer rather than omit: per-strategy result-shape documentation for every enum value, and a tasks.md seam-verification task naming the three DE-007 layers (registration surface, documented output shape, round-trips).
