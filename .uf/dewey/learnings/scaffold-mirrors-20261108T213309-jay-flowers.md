---
tag: scaffold-mirrors
author: jay-flowers
category: context
created_at: 2026-11-08T21:33:09Z
identity: scaffold-mirrors-20261108T213309-jay-flowers
tier: draft
---

In unbound-force, TestCanonicalSources_AreEmbedded walks only four canonical directories — .opencode/commands, .opencode/agents, .opencode/skills, .opencode/uf/packs — so a new repo-local persona under .opencode/agents/ requires a knownNonEmbeddedFiles entry in internal/scaffold/scaffold_test.go (the divisor-linguist and divisor-determinism precedent), while .opencode/lib/ is not walked at all: repo-local lib CLIs (instruction-lint.ts, tool-inventory.ts, cli-common.ts) are exempt from mirror sync, and internal/scaffold/assets/opencode/lib/ mirrors only the 14 plugin-support modules. When reviewing a change that adds agent files, check the knownNonEmbeddedFiles map for the entry; when it adds lib files, no scaffold mirror is expected.
