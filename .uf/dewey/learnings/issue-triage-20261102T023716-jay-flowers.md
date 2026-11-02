---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:37:16Z
identity: issue-triage-20261102T023716-jay-flowers
tier: draft
---

Resolve "to be linked" cross-repo website references during docs-issue triage by searching the Dewey GitHub index for the issue's own title terms before flagging a constitutional Cross-Repo Documentation gap — website trackers for the same change are filed with near-identical titles on the same day. unbound-force/unbound-force#717's "unbound-force/website (to be linked)" placeholder resolved to website#303 ("docs: one-call consolidate_dispatch finalization and kind-native run verdicts in review commands", filed 2026-10-13, corroborated by two independent indexed sources: the fix-triage-issue-dispatch-docs spec-review learning and the documentation-gate history's "#718/#717/website#303" cluster), converting an apparent gate gap into verified dual-gate compliance (in-repo #717 + website#303, the dispatch-finalize-determinism tasks 4.4/4.6 exemplar pattern). Corroborate the issue's behavioral claims the same way: #717's claims about one-call consolidate_dispatch finalization and kind-native run verdicts matched the consolidate_dispatch tool's own contract text (provenance → internal finalize, omitted → raw payload), the workflow_verdict enums, and the indexed implementing PR #718 — an objective VALID verdict reached without reading repository code.
