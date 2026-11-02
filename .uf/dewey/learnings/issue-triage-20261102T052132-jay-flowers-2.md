---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T05:21:32Z
identity: issue-triage-20261102T052132-jay-flowers-2
tier: draft
---

When triaging metadata-field bugs in artifact contracts, resolve field semantics against the registered schema and spec contract before classifying. Issue #556 (unbound-force/unbound-force) is the precedent: the artifact envelope carries both `version` ("Hero version (semver)" — producer provenance, must be dynamic) and `schema_version` ("Schema version (semver)" — format version, legitimately constant). A hardcoded "1.0.0" in `version` looks defensible until the schema description and Spec 002's Envelope Fields table ("Version of the producing hero") prove the field is producer provenance. The disambiguation check: does the contract document a separate field for format versioning? If yes, the remaining version-like field is provenance and constant values are false provenance data (a traceability/audit defect, typically MEDIUM — misleading metadata, no direct exploit path). Verification sources that settle such questions, in order: registered JSON Schema field descriptions, the owning spec's field table, then the writer implementation.
