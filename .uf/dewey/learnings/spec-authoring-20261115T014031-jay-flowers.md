---
tag: spec-authoring
author: jay-flowers
category: pattern
created_at: 2026-11-15T01:40:31Z
identity: spec-authoring-20261115T014031-jay-flowers
tier: draft
as_of: "unbound-force@7f44cba6be6f30f6c3cec406daab342434cc1474"
---

A spec that mandates deterministic no-LLM regression assertions on command behavior must name the deterministic component that produces the observable output. When a requirement's addressee is a model-executed command (relay, presentation, comparison prose) but the regression fixtures are mock-store tests, the assertion and the requirement do not match: model-composed output is unobservable in the fixture. Either name the tool that renders the output, or move the assertion to a command eval. Seen in harden-spec-review-gate: HARDEN-004 assigns the relay MUST to /uf.run-pipeline-plan while HARDEN-005(d) and design Decision 5 require a no-LLM fixture asserting "orchestrator relays 8 findings from the artifact".
