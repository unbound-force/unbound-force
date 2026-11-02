---
tag: documentation
author: jay-flowers
category: pattern
created_at: 2026-11-02T21:23:49Z
identity: documentation-20261102T212349-jay-flowers
tier: draft
---

Extension-point changes must ship the execution contract in the authoring surface, not only in spec/design artifacts and the tracked website issue. Triage pattern for documentation review: enumerate the new capability's spec-mandated contract elements (environment variables, cwd semantics per run mode, exit codes) and grep for each in the user-facing authoring docs — the seeded template and the configuration-reference row — because those are the only documents every extension consumer reads. Mode-dependent details are the gap magnets. Verified on preflight-project-checks (unbound-force, opsx branch, council iteration 2): PF-FR-009's UF_PREFLIGHT_PROJECT_ROOT reached the spec delta, design.md D5, a lib code comment, a unit test, and website issue #373, yet was absent from the .uf/preflight.yaml template whose comment block teaches the rest of the schema — leaving stealth-install check authors with no in-repo way to learn the only correct config-discovery mechanism (cwd is the merge-base worktree during soft-gate baseline re-runs, where git-excluded .uf/ trees do not exist). The recommended fix must edit the canonical file and its byte-identical scaffold mirror together, or the drift test fails.
