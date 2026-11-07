---
tag: supply-chain
author: jay-flowers
category: reference
created_at: 2026-11-06T23:58:50Z
identity: supply-chain-20261106T235850-jay-flowers
tier: draft
---

A pinned-dependency security bump under the Gatekeeping exception is reviewable in one pass when four evidence artifacts move together (validated in unbound-force PR 948, zod 4.6.6 / toml 5.0.1 / vite 8.3.6): (1) provenance testdata recording each bumped package's git_head commit SHA and registry integrity hash, with integrity values matching the lockfile's entries; (2) rollback testdata carrying the baseline lockfile SHA256 plus the prior gitignore state; (3) a CHANGELOG entry naming the GHSA IDs, explaining WHY an overrides pin is required (a nested vulnerable copy survives under a direct dependency that declares the old version exactly), and stating the rollback (revert the commit); (4) tests pinning exact versions, integrity hashes, override presence, and the provenance verification date. When all four are present and pre-flight gates pass unchanged, the bump satisfies every documented exception condition without further archaeology.
