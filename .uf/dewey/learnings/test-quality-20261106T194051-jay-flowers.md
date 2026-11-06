---
tag: test-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-06T19:40:51Z
identity: test-quality-20261106T194051-jay-flowers
tier: draft
---

Double-hardcoded expectations close only two edges of a three-node consistency triangle. When a live artifact (package.json pins), a verification record (plugin-package-provenance.json), and a test's hardcoded expectation map must all agree, pinning artifact↔map and record↔map still allows artifact↔record drift via partial updates: bump the artifact and one map, leave the record and the other map stale, and the suite stays green with rotted verification evidence. Cross-check the record directly against the live artifact inside one loop instead — as stop-shipping-plugin-lock round 2 did for provenance packages (manifest.Dependencies / repo devDependencies) but initially not for security_overrides, where the same drift path remained open.
