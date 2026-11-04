---
tag: drift-detection
author: jay-flowers
category: pattern
created_at: 2026-11-04T15:45:39Z
identity: drift-detection-20261104T154539-jay-flowers
tier: draft
---

A drift-detection guard that skips files whose mirror copy is absent is presence-conditional: it catches content drift between existing pairs but silently tolerates absence drift — a new registry file committed without its mirror passes once any pair exists. In unbound-force, TestSchemaRegistryMatchesUFCopies must skip some registry files (only 7 of ~14 schemas/ families ship to .uf/schemas/), so the skip itself is load-bearing; the absence-proof scoping is per-family: when .uf/schemas/<family>/ exists, require every schemas/<family>/*.json to have a byte-identical .uf copy. Gotcha: the glob schemas/*/*.json also matches schemas/samples/*.json, which has no .uf mirror and must stay exempt. General rule: when a mirror ships a subset of a registry, scope presence requirements to the subset's families rather than skipping on absence file-by-file.
