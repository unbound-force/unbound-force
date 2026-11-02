---
tag: input-validation
author: jay-flowers
category: gotcha
created_at: 2026-11-02T05:25:31Z
identity: input-validation-20261102T052531-jay-flowers
tier: draft
---

A validator that exists and passes its own tests provides no runtime protection until it is wired into the consumption path — green unit tests on the validator do not exercise the read path. In unbound-force (issue #561), internal/schemas/validate.go (ValidateArtifact/ValidateBytes via santhosh-tekuri/jsonschema/v6) is CI-exercised, yet ReadEnvelope (internal/artifacts/artifacts.go) unmarshals artifact envelopes without calling it, so FindArtifacts/FindArtifactsByHero/FindArtifactsSince consume unvalidated cross-agent file contents — and enforcement is further gated by a writer/schema mismatch (the writer emits the schema-required `context` field as omitempty, issue #557), so a validator's enforcement readiness also depends on writer-schema reconciliation. When auditing input validation, trace the call chain from each external-input boundary to the validator; "the validator exists and is tested" is an integration claim to verify, not a fact.
