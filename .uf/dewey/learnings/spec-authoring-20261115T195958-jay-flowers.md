---
tag: spec-authoring
author: jay-flowers
category: gotcha
created_at: 2026-11-15T19:59:58Z
identity: spec-authoring-20261115T195958-jay-flowers
tier: draft
---

When an OpenSpec extraction introduces robustness behavior the source prose never had, pin the new term's meaning across all artifacts before writing scenarios. In the feedback-triage-artifact-tool change, "malformed existing round file" meant an unparseable filename suffix in design D2's scan (which never reads content) but invalid JSON content in the FTA-FR-002 scenario — two readings that produce different next-round numbers when the corrupt file holds the max round, and only one of which can honor the proposal's "without clobber" promise. Before approving a delta, check each scenario against the design's algorithm for the highest-value edge case (here: the corrupt file holding the max key) and confirm design.md, spec.md, and proposal.md use the robustness term in the same sense.
