---
tag: openspec-governance
author: jay-flowers
category: gotcha
created_at: 2026-11-07T19:14:42Z
identity: openspec-governance-20261107T191442-jay-flowers
tier: draft
---

A pinned-dependency security bump under the documented gatekeeping exception still requires OpenSpec delta reconciliation when an unarchived delta pins the old version. In unbound-force, SC-FR-003 (clarification-marker-superset, unarchived) carries MUST text pinning "Zod 4.1.8" and a rollback baseline lock SHA; the OSV-remediation bump to 4.6.6 updated CHANGELOG, provenance testdata, and the lockstep Go pin tests, but left the delta contradicting shipped behavior — the same class the retroactive-push-skill-contracts change remediates for five behavior deltas. The repo's own precedent (fix-triage-issue-dispatch-docs task 5.5 superseded-by annotation; clarification-marker-superset design D7 record move) shows the remedy: a pin bump touches every unarchived delta that names the pin, and the annotation or amending delta ships in the same change. A pin is a contract like any behavior MUST; the security exception authorizes the update, not the spec-corpus contradiction.
