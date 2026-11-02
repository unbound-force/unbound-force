---
tag: documentation-pipeline
author: jay-flowers
category: pattern
created_at: 2026-11-02T17:47:55Z
identity: documentation-pipeline-20261102T174755-jay-flowers
tier: draft
---

When a change introduces a new error/status code into a plugin whose artifacts are validated against the schemas/ registry (e.g., host_model_structural joining host_model_unavailable in review-dispatch run errors), the documentation-pipeline downstream check is two-step: (1) grep schemas/ for the sibling codes and inspect the constraint SHAPE — runError.code in schemas/review-dispatch/v1.0.0.schema.json is a pattern (^[a-z][a-z0-9_]*$), not an enum, so a new lowercase snake_case code validates with no schema change, whereas an enum constraint would have required a same-PR schema update; (2) check schema samples depicting the affected scenario for staleness, but only flag a sample when the change makes it newly misleading — valid-local-review-dispatch.json depicts a host-source run carrying host_model_unavailable inside runs[], while host-resolution failures are pre-execution early exits that write no run file; that illustrative-sample tension predates the 2026-11 split change (RD-FR-037 semantics) and stays out of diff-confinement scope.
