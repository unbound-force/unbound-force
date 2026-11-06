---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-05T16:29:38Z
identity: test-quality-20261105T162938-jay-flowers
tier: draft
---

When a multi-branch diagnostic (doctor check) is rewritten with an outcome test matrix, enumerate the covering tests from the implementation's full branch set — every distinct Pass/Warn/Fail return path, including the json.Unmarshal failure path for each parsed file — not from the spec scenarios alone. The routinely missed case is the present-but-malformed input branch: specs describe valid, absent, and mismatched states but rarely corrupt ones (unbound-force PR 957 checkManifestInstallConsistency shipped six outcome tests yet missed the malformed-manifest Warn that design D4 explicitly documents; contrast scaffold-dcp-config FR-007, where the malformed-config Warn branch received a full GIVEN/WHEN/THEN scenario). One-test-per-branch enumeration makes silent outcome changes — a Warn reworded to a Fail, a dropped remediation hint — fail the suite immediately.
