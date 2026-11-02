---
tag: artifact-provenance
author: jay-flowers
category: gotcha
created_at: 2026-11-02T05:21:35Z
identity: artifact-provenance-20261102T052135-jay-flowers
tier: draft
---

Value-drift failure mode of the unbound-force artifact envelope contract (issue #556; complements the #557 required-ness learnings schema-contracts-20261102T051816/17): the envelope's required `version` field — documented as "Hero version (semver)" in schemas/envelope/v1.0.0.schema.json, the envelope README, website docs, and Spec 002 — was populated from an exported package constant `const Version = "1.0.0"` in internal/artifacts/artifacts.go, so every envelope from every release reported 1.0.0 while passing all schema checks. Schema-valid is not provenance-correct: structural validation cannot catch a required field populated from a constant, so triage of metadata defects must verify the value's source, not just field presence. Wiring note for the fix: the release pipeline already injects the binary version via `-X main.version` (.goreleaser.yaml), but internal/ MUST NOT import cmd (go pack AP-007), so the convention-consistent fix is link-time `-X` injection directly into a package-level `var Version = "dev"` in internal/artifacts (const→var is read-compatible; no runtime mutation) or threading through an Options struct (go pack AP-001); a runtime SetVersion(v) mutator reintroduces the global mutable state that AGENTS.md and go pack CS-007 prohibit. Gaze demonstrates the intended pattern via gaze_version in gaze/internal/taxonomy/types.go Metadata.
