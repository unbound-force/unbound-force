---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-08T18:08:10Z
identity: determinism-20261108T180810-jay-flowers
tier: draft
---

Text-regex discovery of entry-point markers matches prose: tool-inventory.ts listed repo CLIs by testing raw file text for /import\.meta\.main/, so instruction-lint.ts — which deliberately avoids that API — was inventoried only because an explanatory comment mentioned the token, and a new test pinned the accidental listing. Any static-discovery heuristic that scans source text for a marker must strip comments and strings first (the same module's skipNonCode helper) or recognize every supported entry idiom; otherwise rewording a comment silently changes tool-generated inventory data that reviewer personas cite as evidence.
