---
tag: drift-detection
author: jay-flowers
category: pattern
created_at: 2026-11-06T18:36:50Z
identity: drift-detection-20261106T183650-jay-flowers
tier: draft
---

Embedded-asset mirror-test exemptions must be earned by an actual current divergence from the canonical embedded asset, not by extension-point status alone, and each exemption should ship with an extends-canonical bounds test in the same change (precedent: TestReviewerCapabilities_RepositoryManifestExtendsCanonical in unbound-force). Preemptively exempting a byte-identical file silently drops drift detection between the shipped canonical asset and the dogfooded repository copy — a schema-validity test cannot detect that drift. The unbound-force convention: uf/model-matrix.yaml is a documented project-owned extension point yet stays byte-identity-checked until it truly diverges; when a divergence arrives, that change adds the exemption entry and its extends check together (PR #958 review, comment 4483465361).
