---
tag: drift-detection
author: jay-flowers
category: pattern
created_at: 2026-11-06T18:36:48Z
identity: drift-detection-20261106T183648-jay-flowers
tier: draft
---

Exempting an asset from a byte-identity mirror test (embedded-asset drift detection) is gate narrowing: every exempted file needs replacement coverage proportional to the dropped guarantee, and exemptions belong only on files that actually diverge. Schema validation is not drift detection — a schema-valid repo-local policy file can still silently diverge from the embedded canonical asset shipped to generated projects. Working pattern: an extends-canonical test asserting canonical entries are present, unchanged, and in canonical order, with additions validated (see TestReviewerCapabilities_RepositoryManifestExtendsCanonical in unbound-force). Precedent for the failure mode: unbound-force issue #941, where a canonical edit that skipped its .uf/ mirror went unnoticed until a mirror test existed (documented at internal/scaffold/scaffold_test.go:2237-2245). When a currently byte-identical file is preemptively exempted because it may diverge later, prefer keeping it enrolled and adding the exemption just-in-time when real divergence arrives, so the silent-drift window stays zero.
