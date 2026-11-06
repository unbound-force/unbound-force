---
tag: spec-conventions
author: jay-flowers
category: gotcha
created_at: 2026-11-06T14:48:03Z
identity: spec-conventions-20261106T144803-jay-flowers
tier: draft
---

An OpenSpec change that self-reports delta-format deviations in its PR "Known Issues" section does not discharge the MP-013/MP-014 obligation: the capability header comment, requirement-numbering allocation note, and requirement ID lines (or the documented unnumbered-predecessor exception) must land in the delta before merge. openspec validate passes without them, and MP-002 forbids renumbering/retrofit after archive, so review time is the last fix window — treat the self-report as confirmation of the deviation, not resolution. Observed on fix-pinact-gaze-pin (PR #972): its ci delta omitted all four overlay elements while 23 sibling deltas carried capability headers and 100+ ID lines existed across active changes.
