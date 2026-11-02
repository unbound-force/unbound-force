---
tag: artifact-provenance
author: jay-flowers
category: gotcha
created_at: 2026-11-02T05:21:37Z
identity: artifact-provenance-20261102T052137-jay-flowers
tier: draft
---

Provenance metadata fields can pass schema validation while carrying meaningless constant data. In unbound-force issue #556, the artifact envelope's `version` field was stamped from a hardcoded `const Version = "1.0.0"` in internal/artifacts/artifacts.go even though the envelope schema defines the field as "Hero version (semver)", the public website docs promise "Version string of the producing hero", and .goreleaser.yaml already injects the real release version via `-X main.version={{.Version}}`. When auditing provenance metadata (producer identity, version, timestamp, artifact type), verify each field is derived from live build/runtime context rather than a constant — presence plus schema conformance satisfies the letter of a metadata requirement, not its intent. Format-version constants belong in a dedicated `schema_version` field; the envelope schema's separation of `version` (hero) from `schema_version` (format) is the pattern to copy.
