---
tag: test-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-01T13:46:39Z
identity: test-quality-20261101T134639-jay-flowers
tier: draft
---

Whitelisted override fields need an end-to-end propagation assertion, not just parse/merge coverage. In the model-matrix review (PR #894), profiles.*.variant was schema-tested, override-whitelist-tested, and merge-tested — yet no test asserted it flowing through resolveSelection into the child-session spawn: every selection fixture used variant-less profiles, and the only spawn-level variant assertion covered the explicit input variant. A regression in the propagation helper would have been silent. Rule of thumb: for every user-configurable field on a whitelisted path, pin at least one assertion at the consumption seam (the spawn/request boundary), and audit selection fixtures for fields that are configurable but never carried.
