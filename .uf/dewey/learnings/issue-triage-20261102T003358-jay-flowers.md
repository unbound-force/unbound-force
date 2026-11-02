---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:33:58Z
identity: issue-triage-20261102T003358-jay-flowers
tier: draft
---

Migration phase-child issues triage cleanly as non-duplicates when the relationship edges are declared in both directions: the child's body opens with "Umbrella: #NNN" (and optionally "Incorporates #NNN workstream N"), and the parent umbrella's phase-tracking checklist reciprocally links the child by number. Keyword duplicate searches surface the umbrella as the top candidate; the two declared edges disambiguate child-vs-duplicate without issue-graph lookups. Verified triaging unbound-force #801 against candidate #797: #801's body opens with "Umbrella: #797 / Phase: 2 — Scaffold & installer", and #797's "Phase 2 — Scaffold & installer" section lists "#801 — Scaffold version-gated dual-emission + v2 installer paths". When the child-side header is missing, the triager should check the candidate parent's body for a reciprocal phase reference before classifying the issue as a duplicate.
