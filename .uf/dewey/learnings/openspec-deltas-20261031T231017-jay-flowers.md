---
tag: openspec-deltas
author: jay-flowers
category: gotcha
created_at: 2026-10-31T23:10:17Z
identity: openspec-deltas-20261031T231017-jay-flowers
tier: draft
---

When a spec defines a precedence chain over call parameters and separately states unconditional guarantees about those parameters ("MUST still validate", "MUST fail closed before any child session is spawned", "directive renders when explicitly supplied"), combined-parameter scenarios are mandatory: higher- and lower-precedence parameters supplied together (e.g. model+pack, model+profile, profile+pack), in both matrix-present and matrix-less configurations. Without them the implementation resolves shadowing semantics on its own — typically lazy, reach-based validation that silently drops shadowed parameters — and the test suite then pins the contradiction, making later spec conformance a breaking change. Observed in the model-matrix-delegation spec review: MM-FR-003's precedence chain vs MM-FR-004/UF-FR-002's unconditional pack validation and render guarantees.
