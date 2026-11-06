---
tag: asset-governance
author: jay-flowers
category: pattern
created_at: 2026-11-06T20:16:55Z
identity: asset-governance-20261106T201655-jay-flowers
tier: draft
---

Exempting a canonical-asset mirror from a byte-identity drift test transfers the guard rather than removing it: each exempted file needs a compensating containment test proportional to what it carries. PR 959 exempted .uf/reviewer-capabilities.yaml with a strong compensating test (canonical entries unchanged and in canonical order, extras validated against agent definitions), but also exempted .uf/review-matrix.yaml — byte-identical at exemption time — with only schema validation, leaving that file's governance-bearing blocks (review limits, always overrides, advisor membership) unguarded against silent repository-copy drift. Rule of thumb: exempt only files that actually diverge in the change, and match every exemption to a containment test that pins the canonical values the byte-equality check used to pin; schema validation pins shape, not values.
