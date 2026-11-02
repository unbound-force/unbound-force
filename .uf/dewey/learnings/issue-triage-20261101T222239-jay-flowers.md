---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:22:39Z
identity: issue-triage-20261101T222239-jay-flowers
tier: draft
---

Unbound Force issue-triage pattern for deferred-scope follow-up issues: verifiability rests on three traceability anchors — the parent story's acceptance criterion quoted verbatim, the delivering PR that shipped the partial capability, and the design.md Non-Goals entry that recorded the deferral. Optional items requiring a different spec tier (e.g., CI changes need their own spec per AGENTS.md) are correctly self-annotated as out-of-scope ("needs its own spec") and are best formalized by a split recommendation extracting them into dedicated tracking issues, rather than treated as scope defects in the parent issue. Documentation-issue gates (docs-labeled issue in the current repo plus unbound-force/website issue per the Cross-Repo Documentation rule) attach to the implementing PR of a user-facing wiring change, not to the tracking issue itself — their absence from a tracking issue does not affect triage validity. Verified triaging #869 (uf gate harness wiring, parent #514, delivering PR #582, deferral in openspec/changes/uf-gate/ design.md Non-Goals).
