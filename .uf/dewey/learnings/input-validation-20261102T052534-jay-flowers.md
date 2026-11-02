---
tag: input-validation
author: jay-flowers
category: gotcha
created_at: 2026-11-02T05:25:34Z
identity: input-validation-20261102T052534-jay-flowers
tier: draft
---

Validation infrastructure that exists and passes CI does not imply enforcement at the consumption boundary. In unbound-force (issue #561), internal/schemas ValidateBytes/ValidateArtifact have extensive _test.go coverage across eight test files, yet a repo-wide grep shows zero production call sites — ReadEnvelope (internal/artifacts/artifacts.go) unmarshals artifact envelopes with no schema check. CI tests that exercise a validator prove the validator works, not that consumers call it. When triaging or reviewing validation-gap reports, verify the call chain from the input boundary to the validator — grep for production call sites — rather than accepting the validator's existence as evidence of enforcement.
