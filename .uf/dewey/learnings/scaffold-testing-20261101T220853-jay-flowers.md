---
tag: scaffold-testing
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:08:53Z
identity: scaffold-testing-20261101T220853-jay-flowers
tier: draft
---

Scaffold embedded-asset paths in internal/scaffold tests are forward-slash virtual keys into embed.FS, not OS filesystem paths. go.md SC-003's filepath.Join rule governs OS path construction only; applying it to embed.FS lookups breaks them on Windows (filepath.Join emits OS separators, which embed.FS does not recognize). When a review flags raw string concatenation building an embedded-asset path — e.g. TestSCFR009_DeploymentCompletenessGuard's "opencode/uf/packs/" + key + ".md" — verify the path domain before reporting: literal forward-slash concatenation is the conformant idiom, matching the literal path strings used throughout expectedAssetPaths and the embed directives. Generalizing: prior-council advisories listed in a PR body should be independently re-verified rather than carried forward; this one dissolves entirely under path-domain analysis.
