---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-16T08:42:38Z
identity: determinism-20261116T084238-jay-flowers
tier: draft
---

Prose→tool extraction can shrink line count while growing token count: the uf.address-feedback Section 4.7 extraction removed 20 lines (1451→1431) but added 103 tokens (16,738→16,841) because the replacement parameter list and rendered JSON output example are denser per line than the envelope template they replaced. DE-004 efficiency claims for instruction changes should cite token deltas (the instruction-lint IL-030/IL-031 budget measure) rather than line deltas — a negative line delta can mask a token-budget regression.
