---
tag: supply-chain
author: jay-flowers
category: gotcha
created_at: 2026-11-02T02:43:23Z
identity: supply-chain-20261102T024323-jay-flowers
tier: draft
---

Unpinned tool installs undermine regeneration-based upgrade plans. When an upgrade reconciles scaffolded artifacts by regenerating them with a version-sensitive generator — `openspec update` reads `generatedBy` markers and emits version-specific output — an unpinned install of that generator (`npm install -g @fission-ai/openspec@latest`) makes the reconciliation non-reproducible: contributors on different CLI versions produce different assets from the same plan, silently reintroducing the drift the upgrade exists to fix. Constitution Principle V treats name-addressed pins as insufficient and `@latest` as fully mutable. Triage rule: when an issue's current-state section flags an unpinned install, check whether any plan phase depends on that tool's version; if so, the issue should scope in pinning or content-integrity verification, or explicitly record the exclusion as a decision — as written, the reproducibility gap is an oversight, not a choice.
