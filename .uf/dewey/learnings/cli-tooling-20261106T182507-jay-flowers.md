---
tag: cli-tooling
author: jay-flowers
category: gotcha
created_at: 2026-11-06T18:25:07Z
identity: cli-tooling-20261106T182507-jay-flowers
tier: draft
---

existsSync and statSync().isDirectory() are not interchangeable for repo-root marker detection: existsSync(join(dir, '.opencode')) is true for a regular file named .opencode, so a root resolver built on it stops at the wrong directory, while an isDirectory-based resolver keeps walking to the real root. The downstream failure is silent — listing '.opencode/plugins' returns [] and the CLI exits 0 with an empty inventory — so the divergence surfaces as contradictory evidence rather than an error. Pin the directory-only semantics with a unit test asserting that a file named .opencode does not terminate the walk.
