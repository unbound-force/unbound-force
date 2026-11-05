---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-04T22:58:35Z
identity: review-dispatch-20261104T225835-jay-flowers
tier: draft
---

When a change introduces a requirement whose scenarios govern command prose (consumer contracts), enumerate every Scenario block in the spec delta and map each to a command-contract-test clause before considering coverage done — scenarios without a pin are the gap, exactly as error-code enumeration finds uncovered fail-closed codes. Applied to PR #949 (RD-FR-056, artifact-sourced consumer contract): four of five scenarios had pins in the new TestCommandContracts_ArtifactSourcedLessons, but the failed-finali
