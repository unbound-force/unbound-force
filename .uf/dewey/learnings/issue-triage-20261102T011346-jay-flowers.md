---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:13:46Z
identity: issue-triage-20261102T011346-jay-flowers
tier: draft
---

Triaging a meta-repo documentation-tracking issue gains a strong corroboration signal from its cross-repo companion: Dewey's GitHub-indexed sources surfaced unbound-force/website#316 ("Docs: split-panel dispatch verdicts now finalize per documented rules"), which explicitly names unbound-force/unbound-force#778 as its upstream tracker and mirrors the same checklist (split-panel verdict semantics, zero-success pass-through, vote-derived verdict vs severity-derived verdict_reason/advisories, uf init scaffold note). A bidirectionally-linked companion in the website repo confirms the cross-repo documentation-sync gate was fulfilled, shows the tracking work is real on both documentation surfaces, and rules out both orphaning and duplication in a single check — website#316 is a companion, not a duplicate, precisely because it targets a different surface (public website pages vs in-repo command behavior contracts). Existing learnings issue-triage-20261101T221610 (per-artifact dispositions) and issue-triage-20261101T233554 (grep-verification of claimed gaps) cover the in-repo quality signals; this adds the cross-repo companion check as a third corroboration layer for docs-issue triage.
