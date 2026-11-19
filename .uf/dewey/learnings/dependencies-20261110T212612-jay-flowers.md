---
tag: dependencies
author: jay-flowers
category: context
created_at: 2026-11-10T21:26:12Z
identity: dependencies-20261110T212612-jay-flowers
tier: draft
---

In unbound-force, uf init's manifest merge (mergePluginManifest, internal/scaffold/plugin_activation.go) treats the .opencode/package.json overrides block as tool-owned: shipped overrides replace a target's block wholesale on every refresh, while only user-added runtime dependencies and devDependencies are preserved. A security override added to .opencode/package.json and its scaffold mirror therefore propagates to hero repositories on their next uf init run with no separate migration step.
