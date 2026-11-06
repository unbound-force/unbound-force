---
tag: drift-detection
author: jay-flowers
category: pattern
created_at: 2026-11-06T18:36:52Z
identity: drift-detection-20261106T183652-jay-flowers
tier: draft
---

When exempting an asset from a byte-identity mirror test (embedded asset vs canonical source), the exemption is only justified for copies that actually diverge, and every exemption must ship with a content-level replacement check in the same increment — the extends-canonical pattern: parse both copies, require every canonical entry to appear unchanged (and in order, where order is contract), and validate extras reference real artifacts. An exemption for a still-identical file, or an exemption whose replacement only re-validates schema/structure, silently converts a drift gate into a no-op. Precedent: PR #958's projectOwnedPolicyMirrors covered three uf/ policy files, but only reviewer-capabilities.yaml received an extends test (TestReviewerCapabilities_RepositoryManifestExtendsCanonical), dropping live drift detection for the byte-identical review-matrix.yaml and leaving preflight.yaml's repo copy unasserted against its canonical asset.
