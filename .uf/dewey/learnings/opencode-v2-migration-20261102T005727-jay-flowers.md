---
tag: opencode-v2-migration
author: jay-flowers
category: gotcha
created_at: 2026-11-02T00:57:27Z
identity: opencode-v2-migration-20261102T005727-jay-flowers
tier: draft
---

OpenCode v2 flips `external_directory` from ask-by-default (v1) to allow-by-default, and unbound-force's CI review sandbox depends on the old default: council-review-action/docs/architecture documents `external_directory: deny` (set via OPENCODE_CONFIG_CONTENT) as the layer confining file reads to the project workspace while processing untrusted PR diff content. During the v2 migration (issue #797; Phase 1 plugin port #714; CI v2 leg #800), every v2-path config emission MUST carry an explicit external_directory rule (`ask` for interactive runtimes, `deny` for the CI sandbox) — otherwise the sandbox posture silently regresses to allow-by-default. General rule: when migrating agent runtime major versions, audit permission-model default changes as first-class security risks and trace each changed default to every consumer that relies on it.
