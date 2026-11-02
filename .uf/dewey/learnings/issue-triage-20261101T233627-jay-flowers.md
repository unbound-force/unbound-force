---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T23:36:27Z
identity: issue-triage-20261101T233627-jay-flowers
tier: draft
---

A documentation-gate issue need not be fully dispositioned to triage as VALID: when residual per-file decisions are explicitly framed ("residual decision tracked here"), each naming the target file, the concrete question, and the related tracker it is decided alongside, the issue is the decision owner and remains actionable — NEEDS-CLARIFICATION applies only when a disposition is vague or unresolvable from repo context. Two verification checks beyond the standard disposition cross-check: (1) confirm the originating change's tasks.md Documentation-phase task is complete and its recorded scope agrees with the issue's dispositions (issue bodies are written at filing time and can lag the change's recorded state); (2) for a change extending a contract already tracked by a parent change's docs-gate issue, the child issue keeps its own gate record and coordinates the shared decision surface explicitly ("decide alongside #791") rather than duplicating it — distinct change scope satisfies both the documentation-gate mandate and the zero-waste rule. Recurring wording hazard: closure conditions phrased "close when the PR merges OR the dispositions resolve" leave ambiguous whether merging with undecided dispositions satisfies closure, risking an orphaned decision — "and", or a named deferral owner, removes the ambiguity (observed as a LOW, non-blocking note on unbound-force/unbound-force#834).
