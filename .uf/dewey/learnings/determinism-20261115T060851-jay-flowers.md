---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-15T06:08:51Z
identity: determinism-20261115T060851-jay-flowers
tier: draft
as_of: "unbound-force@e5e077b35f0c7984517024431e696737e330b080"
---

DE-002 output-format examples must land in the consuming instruction files, not only in the change's design doc — and must be copied from the shipped renderer's golden-test assertions, not drafted at design time. Verified on harden-spec-review-gate: design.md carried a "Rendered-relay example (DE-002 — preserved near the presentation point)" that (a) never reached either consuming command (uf.run-pipeline-plan.md / uf.unleash.md describe the relay only as a field list) and (b) diverged from renderSpecReviewGateRelay's actual output ("Spec-review gate holds: iteration 1 (REQUEST CHANGES), correlation …" / "Findings from the dispatch artifact:" / "1. [HIGH] description (file:line) — mark"), which only the relay test assertions pin. Design-doc examples drift silently because nothing tests them; the golden/relay test assertions are the source of truth to quote when carrying an example into command prose.
