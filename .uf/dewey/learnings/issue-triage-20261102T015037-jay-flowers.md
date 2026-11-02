---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:50:37Z
identity: issue-triage-20261102T015037-jay-flowers
tier: draft
---

Refactor-issue triage in the unbound-force meta repo validates cleanly when three signals align: (1) provenance — a review-council run ID and dispatch correlation ID recording where the deferred finding originated; (2) verifiable evidence — target file, line count, and named inline functions checkable via grep/read, verifying by function name rather than line anchor since line numbers drift as the file grows; (3) convention-aware remediation — extraction into .opencode/lib/ following the branch-classification.ts / review-dispatch-artifact.ts precedent, which mandatorily carries a byte-identical scaffold mirror under internal/scaffold/assets/opencode/ for every extracted module (all 12 existing lib modules are mirrored, plus the plugin entry point). A proposal citing the extraction pattern but omitting the mirror obligation is structurally incomplete. Verified on issue #741 (uf-workflow index.ts decomposition): the file had grown from ~3,700 lines at filing to 5,402 lines and 27 tool registration sites by triage time — post-filing growth strengthens rather than weakens a decomposition issue.
