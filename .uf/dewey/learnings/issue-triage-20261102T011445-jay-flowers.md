---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:14:45Z
identity: issue-triage-20261102T011445-jay-flowers
tier: draft
---

Triage verification of documentation-gate records must match each claim to the right evidence epoch: point-in-time dispositions (e.g., "determinism pack bumped 1.1.0 → 1.2.0") verify against durable historical records — the CHANGELOG entry recording the bump — not against current repository state, where the artifact has since evolved (the determinism pack now sits at 1.4.0 after later DE-008 additions). Checking a filing-time claim against current state produces false INVALID signals. The strongest verification loop is three-way agreement: the issue's dispositions ↔ the parent change's completed Documentation-phase tasks (tasks.md 4.3/4.4 checked) ↔ the change's evidence artifact recording the filed issue number (metrics.md "Documentation issue filed: #781"). Observed triaging unbound-force/unbound-force#781 (seam-verification docs gate): all three agreed, and the CHANGELOG entry carried "Refs: #765" tying it to the parent change.
