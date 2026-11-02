---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:44:23Z
identity: issue-triage-20261102T014423-jay-flowers
tier: draft
---

Deferral-to-tracked-issue pattern (observed in unbound-force/unbound-force#742): when a review-council fix loop defers a recommendation by maintainer decision, the follow-up issue stays auditable when it carries four elements — the trigger condition ("when #716 closes"), the exact target action ("file an epic-level blog issue in unbound-force/website"), the verifiable evidence list the future work draws on (ceremony-reduction counts, LOC/tool counts, dated constitution amendment, design decision reference), and machine-checkable provenance (curator run id + dispatch correlation id). With those elements, the deferred work's factual claims can be re-verified against recorded artifacts instead of reconstructed from review-thread archaeology, so the eventual content satisfies the content pack's fact-checking rules (FA-001/FA-004) by construction.
