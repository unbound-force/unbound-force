---
tag: opencode-v2-migration
author: jay-flowers
category: gotcha
created_at: 2026-11-02T00:34:08Z
identity: opencode-v2-migration-20261102T003408-jay-flowers
tier: draft
---

Testability constraint for dual-emission scaffolds under the OpenCode v2 migration (discovered triaging unbound-force#801, Phase 2): v1 and v2 cannot be installed side by side — both claim the `opencode` command (umbrella #797 verified fact) — so drift tests asserting BOTH emission shapes cannot rely on real runtimes being present in CI. The runtimeVersion probe (unbound-force: internal/scaffold/plugin_activation.go) must be an injectable/stubbed seam so each shape is asserted deterministically under a simulated runtime version; without that seam, the done-when criterion "drift tests enforce both shapes" is satisfiable only via environment-dependent e2e on machines running one major or the other, violating the repo testability principle (tests must not depend on external environment state). Triage check for dual-emission migration issues: does the issue or its spec name the probe-injection seam? If not, flag it as a spec-phase requirement — the gap does not block issue validity, but the spec that implements it must define the seam for the dual-shape assertions to be CI-enforceable.
