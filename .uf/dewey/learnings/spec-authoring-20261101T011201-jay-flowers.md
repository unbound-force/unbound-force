---
tag: spec-authoring
author: jay-flowers
category: gotcha
created_at: 2026-11-01T01:12:01Z
identity: spec-authoring-20261101T011201-jay-flowers
tier: draft
---

A blanket resilience rule in a spec ("X MUST NEVER block Y" / "no error is surfaced to the caller") silently swallows adjacent fail-closed requirements unless every fail-closed scenario names its failure surface. Pattern seen in model-matrix-delegation (unbound-force/unbound-force): MM-FR-005's classifier-never-blocks enumeration (timed-out, errored, malformed, below-threshold, matrix-invalid → static fallback) collides with MM-FR-006's literal-apiKey and non-loopback-endpoint scenarios ("fails closed with a non-retryable error") — both a delegation-blocking error and a silent engine-skip satisfy the text, and the design's structural/runtime failure dichotomy (D4) omitted the third class, engine-config validation. When authoring a fail-closed requirement inside a never-block flow: state in the scenario THEN-clause whether the failure blocks the calling operation or only skips the component (the spec's own unset-env-var scenario — "the engine is skipped, the cascade falls back statically" — is the model), and extend the design's failure taxonomy to classify the new failure. Symptom: a fail-closed scenario whose "it" (the engine vs. the delegation call) is the only difference between loud and silent failure on a security misconfiguration.
