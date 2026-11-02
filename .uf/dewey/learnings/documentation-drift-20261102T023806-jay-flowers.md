---
tag: documentation-drift
author: jay-flowers
category: gotcha
created_at: 2026-11-02T02:38:06Z
identity: documentation-drift-20261102T023806-jay-flowers
tier: draft
---

Convention-pack listing surfaces (AGENTS.md Convention Packs list, docs/usage.md pack listing, docs/architecture.md Pack Types section) drift from the deployed pack set every time a new pack pair lands: the ci/ci-custom and python/python-custom deployments left all three surfaces stale in unbound-force/unbound-force (issue #719). Two consequences: pack-deploying changes must sweep every listing surface, not only AGENTS.md; and docs-gap issues can be silently and partially remediated by later changes while staying open — #719's AGENTS.md omission was fixed by subsequent pack deployments, and its predicted pack-directory count (15 files) was stale within weeks (18 after the determinism pack and meta-prompt pair landed). When fixing or re-triaging pack-listing gap issues, re-verify each named surface and recount from .opencode/uf/packs/ instead of trusting issue-stated counts.
