---
tag: review-dispatch
author: jay-flowers
category: reference
created_at: 2026-11-03T11:11:12Z
identity: review-dispatch-20261103T111112-jay-flowers
tier: draft
---

consolidate_dispatch output contract (PR #929, openspec/changes/bound-consolidate-dispatch-output): full retrieval orders correlation_id, proposals, warnings ahead of verdict and the large arrays; `detail: "summary"` returns a bounded observability-first summary (caps 15 findings / 15 advisories / 10 proposals / 5 warnings, per-field truncation with `truncated: true` markers, four top-level truncation flags, post-assembly drop loop to strictly under 40,960 UTF-8 bytes in the order findings, advisories, proposals, warnings); provenance-bearing finalization returns a compact result (status, operation_verdict, correlation_id, artifact_path, scalar review_verdict, run_counts, calculated_assessment.authoritative) and full findings/advisories are read from the artifact at artifact_path rather than echoed.
