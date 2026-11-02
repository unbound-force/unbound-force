---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T04:17:56Z
identity: issue-triage-20261102T041756-jay-flowers
tier: draft
---

Deferred acceptance criteria can acquire two trackers: a follow-up issue filed when the deferral was noticed during PR review, and a different issue number later named in the delivering change's design doc (Non-Goals section). Keyword duplicate searches over issue text miss this fork because neither issue references the other — the cross-reference lives only in the design doc. When triaging an issue that cites a design doc's non-goals as its provenance, read that section for a named tracking issue number and reconcile before scheduling: the design-doc-named tracker is canonical; the earlier issue closes as duplicate or keeps the remainder scope. Observed triaging unbound-force #584 (tracks AC#6/AC#7 deferred from #514, filed 2026-09-08): prior learning issue-triage-20261101T222235-jay-flowers records design.md Non-Goals naming #869 for the same harness wiring while the provided keyword duplicate search returned zero candidates.
