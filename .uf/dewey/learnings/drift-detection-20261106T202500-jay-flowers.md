---
tag: drift-detection
author: jay-flowers
category: pattern
created_at: 2026-11-06T20:25:02Z
identity: drift-detection-20261106T202500-jay-flowers
tier: draft
---

Exempting an asset from a byte-identity mirror test (embedded canonical asset vs the dogfooded repository copy) is gate narrowing: every exemption must be earned by an actual current divergence, not by extension-point status, and must ship with a content-level replacement check in the same change — the extends-canonical pattern: parse both copies, require every canonical entry to appear unchanged (and in canonical order where order is contract), and validate that additions reference real artifacts (see `TestReviewerCapabilities_RepositoryManifestExtendsCanonical` in unbound-force). Schema validation is not drift detection: a schema-valid repo-local policy file can still silently diverge from the embedded canonical shipped to generated projects, and an exemption whose replacement only re-validates schema converts the drift gate into a no-op. Keep currently byte-identical files enrolled even when divergence is anticipated, and add the exemption plus its extends check just-in-time when real divergence arrives, so the silent-drift window stays zero (unbound-force convention: `.uf/model-matrix.yaml` is a documented extension point yet stays byte-identity-checked until it truly diverges). Failure-mode precedent: unbound-force issue #941, where a canonical edit that skipped its `.uf/` mirror went unnoticed until a mirror test existed (documented at internal/scaffold/scaffold_test.go:2237-2245); and PR #958's projectOwnedPolicyMirrors in an intermediate revision, where only reviewer-capabilities.yaml received an extends test — corrected before merge, so both exempted files (uf/preflight.yaml, uf/reviewer-capabilities.yaml) carry extends-canonical tests and uf/review-matrix.yaml remains under byte-identity enforcement.
