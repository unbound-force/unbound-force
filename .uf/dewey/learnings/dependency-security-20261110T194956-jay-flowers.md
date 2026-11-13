---
tag: dependency-security
author: jay-flowers
category: pattern
created_at: 2026-11-10T19:49:56Z
identity: dependency-security-20261110T194956-jay-flowers
tier: draft
---

Security-override provenance governance loops must cross-check in both directions. The unbound-force plugin-package provenance loop (internal/scaffold/scaffold_test.go, TestPluginPackageInputs_ApprovedMaintenanceProvenance) verifies provenance-recorded overrides match both manifests' overrides blocks — catching partial version bumps — but iterating only the expected/provenance keys leaves the converse unenforced: an override added to both manifests (with TestPluginPackageInputs_ExactVersions' hardcoded map updated) but never given a provenance record ships green. The closed loop needs a manifest-to-provenance assertion: every key in either manifest's overrides must appear in provenance security_overrides with version equality. Verified on the effect 4.0.0-beta.103 override change (PR #1024, GHSA-chp3-j5qr-2vfr), where the iteration-3 fix added the provenance-to-manifest direction only.
