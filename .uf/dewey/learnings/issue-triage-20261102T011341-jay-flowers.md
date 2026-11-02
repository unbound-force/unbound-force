---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:13:41Z
identity: issue-triage-20261102T011341-jay-flowers
tier: draft
---

Documentation issues for merged spec-driven changes can be triaged objectively via checklist-to-FR traceability: verify that each "what needs documenting" checklist item maps to a requirement or scenario in the referenced spec delta. Issue unbound-force/unbound-force#778 (docs for the fix-dispatch-verdict-derivation change) exemplifies the pattern — all six checklist items map onto FR-009 (vote-derived workflow verdict: feedback split-panel ACCEPT, no-success pass-through), FR-010 (triage majority), FR-011 (council/test-review non-regression), and FR-012 (severity-derived verdict_reason/advisories) in openspec/changes/fix-dispatch-verdict-derivation/specs/dispatch-finalization/spec.md. A checklist item with no FR anchor signals either out-of-scope content or a missing spec requirement — both worth surfacing at triage time. Related structural check: cross-repo docs counterparts (e.g. unbound-force/website#316 citing #778 as upstream) are complementary trackers, not duplicates.
