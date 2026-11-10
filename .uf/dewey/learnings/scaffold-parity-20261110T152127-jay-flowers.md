---
tag: scaffold-parity
author: jay-flowers
category: pattern
created_at: 2026-11-10T15:21:27Z
identity: scaffold-parity-20261110T152127-jay-flowers
tier: draft
---

Byte-level round-trip tests over a merge function double as drift detection between an embedded scaffold mirror and the canonical source. unbound-force's scaffold suite already derived the embedded .opencode/package.json mirror from the canonical manifest by stripping devDependencies (expectedEmbedded/shippedPluginManifest in internal/scaffold/scaffold_test.go); merge-plugin-runtime-deps added the inverse direction — mergePluginManifest(shipped, canonical) == canonical — so mirror and canonical are now locked from both sides (canonical→strip→embedded and embedded→merge→canonical). Any pin bump, key-order change, or formatting drift on either side fails a test. When introducing a preserve/merge transform over tool-owned assets, prefer adding the inverse round-trip alongside the existing strip-transform drift test.
