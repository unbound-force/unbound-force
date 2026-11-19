---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-16T07:12:22Z
identity: determinism-20261116T071222-jay-flowers
tier: draft
---

Shipped DE-split precedent for human-authorization gates (SYNTH-001..006, synthetic-compaction-continue-prompt-rec): the write tool machine-stamps identity (UUID v4 decision_id) and recorded_at through a strict zod contract that rejects caller-supplied stamps; the enforcement point (checkpoint) resolves the reference against the persisted append-only record fail-closed, checks shape before any record read, and overwrites any caller-supplied same-name data key with the machinery stamp (anti-spoofing provenance); verdict-to-hold mapping (APPROVE WITH ADVISORIES → escalated relay hold) lives in a pure function (isEscalatedAdvisorySpecVerdict), not agent judgment; and instruction prose keeps only the judgment half — the stop-and-obtain-a-real-human-decision recovery path that names the tool's error codes (missing_human_decision, undeclared_step, escalated_advisory_unrelayed) for output recognition — pinned by a Go contract test on both canonical and scaffold-mirror copies. Two write surfaces (in-session record_human_decision, orchestrator-relay envelope persisted at delegation spawn time) serialize through one shared append store, so authorization cannot drift between plugins.
