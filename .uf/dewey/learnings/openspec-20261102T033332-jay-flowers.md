---
tag: openspec
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:33:32Z
identity: openspec-20261102T033332-jay-flowers
tier: draft
---

OpenSpec capability deltas for the same capability can live in both layouts across changes: flat `openspec/changes/<change>/specs/<capability>.md` (e.g., shared-agent-executor/specs/dispatch-agent-run.md, derive-per-run-timeout/specs/review-matrix.md) and directory form `specs/<capability>/spec.md` (e.g., host-fallback-provenance/specs/dispatch-agent-run/spec.md, derive-per-run-timeout/specs/review-dispatch/spec.md — the same change can mix both). When verifying cross-delta supersession chains or planning archive-time materialization, search both layouts; archive-ordering header comments inside each delta declare the intended sequence, and main specs under openspec/specs/ cover only already-archived capabilities.
