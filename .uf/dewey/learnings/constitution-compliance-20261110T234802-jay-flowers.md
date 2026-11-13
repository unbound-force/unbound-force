---
tag: constitution-compliance
author: jay-flowers
category: pattern
created_at: 2026-11-10T23:48:02Z
identity: constitution-compliance-20261110T234802-jay-flowers
tier: draft
---

Quick-tier security fixes (quick/<name> branches, no spec artifacts) that ship interim prose choreography instead of Principle VI tool extraction have exactly one compliant landing spot for the constitution's Conflict Resolution justification: the CHANGELOG entry or PR body. The rule's named locations ('the relevant spec or plan') do not exist on a quick branch, so an entry that documents only the mechanical change — as PR #1024's triage-issue-heredoc entry did (+24/-14 instruction delta per DE-004, but no trade-off note) — leaves the written justification absent from the repo entirely. The complete deferral record needs three parts: (1) the explicit Principle V/VI trade-off statement (injection closed now via interim choreography because the linked issue prescribed that shape), (2) a tracked tool-extraction follow-up issue reference, and (3) the follow-up's DE-007 seam layers named (registration surface, documented output shape, round-trips), matching the deferral path the determinism-domain dispatch requires. Without the record, interim mitigations become the de-facto mechanism of record and remediated command families diverge in shape for one vulnerability class — cf. #1007 (tool-routed uf.address-feedback.md) vs #1011 (prose-choreographed uf.triage-issue.md). The record is necessary but not sufficient: per learning/determinism-20261108T171654, the determinism pack has no deferral clause for new/modified instruction, so the record pairs with — never replaces — the extraction itself.
