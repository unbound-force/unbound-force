---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-10T19:50:38Z
identity: test-quality-20261110T195038-jay-flowers
tier: draft
---

When a pinned value is maintained in N synchronized places, every pairwise join needs an automated check (or all expected literals must derive from one source). Pinning each place against its own independent expected literal leaves silent-drift paths: in unbound-force's plugin-package security-override pin, tests covered canonical-manifest↔scaffold-mirror (byte compare), manifest↔test-literal, and provenance-fixture↔test-literal, but not manifest↔provenance-fixture — so a future version bump that updates two of the three literals passes green while the recorded git_head and dist integrity verify a version that is no longer pinned. Cross-check provenance records against the manifest pins they claim to verify, the way the packages loop in TestPluginPackageInputs_ApprovedMaintenanceProvenance already does.
