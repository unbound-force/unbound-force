---
tag: review-council-multi-model-fanout
author: jay-flowers
category: pattern
created_at: 2026-10-02T16:48:50Z
identity: review-council-multi-model-fanout-20261002T164850-jay-flowers
tier: draft
---

Branch opsx/review-council-multi-model-fanout (2026-10-02): A scope-based Go coverage gate (coverage-gate.json + internal/coveragegate) must classify packages by their actual purity. internal/artifacts and internal/schemas perform filesystem I/O (scanning, reading, and writing artifact and schema files), not pure serialization, so they belong in the 80% helper scope, not the 90% pure artifact/schema scopes. Misclassifying them as pure caused a deterministic CI failure that was masked during local work because the gate derives changed packages from git merge-base (an empty diff when HEAD equals main and everything is uncommitted), so `make coverage-gate` passed vacuously. Always verify each declared package actually meets its threshold, and beware vacuous gate passes on uncommitted work.
