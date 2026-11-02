---
tag: schema-contracts
author: jay-flowers
category: gotcha
created_at: 2026-11-02T05:18:17Z
identity: schema-contracts-20261102T051817-jay-flowers
tier: draft
---

Schema-contract drift pattern (unbound-force issue #557): the artifact envelope contract was defined twice — the writer struct internal/artifacts.Envelope (Context tagged `json:"context,omitempty"`, with WriteArtifact forwarding nil) and the schema-generator struct internal/schemas.EnvelopeSchema (non-pointer ContextSchema field, emitted as required in schemas/envelope/v1.0.0.schema.json). The conformance suite validated only hand-crafted sample envelopes (which included a context object), never envelopes produced by the real writer, so WriteArtifact(nil) output omitted the `context` key and violated the schema with every test green. Two generalizable rules: (1) a versioned contract should have a single source of truth for field required-ness — parallel struct definitions drift silently; (2) schema conformance tests should round-trip actual producer output through the validator, not curated fixtures.
