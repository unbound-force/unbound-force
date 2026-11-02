---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:15:51Z
identity: issue-triage-20261102T001551-jay-flowers
tier: draft
---

Documentation-gate issues in unbound-force verify cleanly at triage time through three anchors. (1) Provenance: the originating change's tasks.md carries the gate task — fix-review-dispatch-correlation-release task 5.5 ("File the in-repo documentation issue ... per the documentation gate") is marked [x], and the implementing PR #809 names the filed docs issues (in-repo #808 and unbound-force/website#319). (2) Targets: glob confirms every checklist path exists (docs/usage.md, docs/architecture.md, schemas/review-dispatch/README.md with v1.0.0.schema.json and samples/). (3) Gap: grep confirms the concept is absent from the target docs — zero "correlation" matches in docs/usage.md and docs/architecture.md — so the documentation debt is real, not already written. Embedded gotcha: conditional checklist items ("update any text implying X") can resolve as no-ops — the root README.md contains no review-dispatch tool text at all (grep for [Dd]ispatch matches only schemas/ READMEs), so grep the target before assuming a docs item requires edits.
