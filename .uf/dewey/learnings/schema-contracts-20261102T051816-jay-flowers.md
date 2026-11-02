---
tag: schema-contracts
author: jay-flowers
category: gotcha
created_at: 2026-11-02T05:18:16Z
identity: schema-contracts-20261102T051816-jay-flowers
tier: draft
---

Contract-drift pattern in unbound-force: the artifact envelope contract exists in four places — specs/009-shared-data-model (User Story 1, Scenario 4 required-fields list), schemas/envelope/README.md (Required Fields table), schemas/envelope/v1.0.0.schema.json (required array), and the Go writer internal/artifacts/artifacts.go. When a schema-required field carries `json:"...,omitempty"` and the convenience write path passes nil (WriteArtifact → WriteArtifactWithContext(..., nil)), every envelope from that path omits the key and silently fails validation against the project's own schema (issue #557). Fast triage verification: read the schema's required array, the struct tag, and the convenience wrapper's nil-default in one pass. Fix-direction guard: prefer realigning the writer (always emit the field) over weakening the schema's required list — the required list is spec-mandated (Spec 009) and `context` carries the Constitution Principle III "against what input" provenance (branch, commit, backlog_item_id).
