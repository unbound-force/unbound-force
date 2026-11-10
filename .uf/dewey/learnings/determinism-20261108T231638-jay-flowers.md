---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-08T23:16:38Z
identity: determinism-20261108T231638-jay-flowers
tier: draft
---

Widening a prose verdict-mapping table with a severity-keyed condition (e.g. "| APPROVE WITH ADVISORIES, any CRITICAL or HIGH finding | COMMENT |") imports a whole fixed procedure into instruction text: enumerate every finding, recognize severity strings, apply the fail-closed unknown-severity default, then branch the event, the gate option set, and a fixed body note on the result. When the condition's inputs are already structured tool output (consolidation findings carry severity fields) and the plugin suite already ships pure verdict mapping (canonicalDecision in review-dispatch), the new condition is the extraction trigger, not one more table row: key the mapping inside an intent-level tool, keep the severity set and fail-closed default as tool constants (DE-006), and turn the spec's Given/When/Then scenarios into the tool's unit tests. A contract test that string-matches the command's prose clauses is the tell — behavior that needs clause-pinning wants a tool home, not pinned prose. Observed on PR #990 (uf.review-pr.md Step 8c, advisory-approval-event); PR #993 learnings on the same file already fix the adjudication pattern.
