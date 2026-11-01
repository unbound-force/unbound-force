---
tag: spec-traceability
author: jay-flowers
category: gotcha
created_at: 2026-10-31T23:10:13Z
identity: spec-traceability-20261031T231013-jay-flowers
tier: draft
---

A MUST-level fail-closed security control in a spec requirement is not traceable until a task whose file scope covers the implementing code owns it. A test-scenario row inside a test-reconciliation task ("a tracked file with populated classifier.llm values fails load naming the override file") makes the gap visible once the test is written, but assigns no one to write the check — and adjacent implementation tasks already marked [x] can claim the requirement is done while its most security-critical clause is missing. When reviewing tasks.md against a spec, verify each MUST/fail-closed security clause maps to an implementation task naming the file that must change, not merely a test that must fail. Seen in model-matrix-delegation iteration 3 (2026-10-31): MM-FR-001's tracked-role credential-population rule existed only as task 2.4's test row while loadModelMatrix (task 2.1, complete) silently accepted populated classifier.llm credentials in the tracked .uf/model-matrix.yaml — the closed schema deliberately accepts populated values for merged views, so no other layer could catch a committed credential.
