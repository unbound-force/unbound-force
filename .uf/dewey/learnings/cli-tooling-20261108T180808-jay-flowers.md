---
tag: cli-tooling
author: jay-flowers
category: gotcha
created_at: 2026-11-08T18:08:08Z
identity: cli-tooling-20261108T180808-jay-flowers
tier: draft
---

Node's `import.meta.main` is undefined before Node 24.2, so a plain `if (import.meta.main) runMain(main)` entry block silently no-ops a plain-node .ts CLI on Node 22 LTS — exit 0, no output — while in-process vitest tests (which import and call main() directly) never execute the entry block, keeping CI green. Observed on unbound-force PR #959: tool-inventory.ts used the import.meta.main guard in the same PR where instruction-lint.ts documented avoiding it for exactly this reason (repo Node floor ^22.22.2, CI on 22.23.3). Structural fix: extract one shared isMain/runIfMain helper (realpath(process.argv[1]) === fileURLToPath(import.meta.url)) into the shared CLI module and use it in every runnable lib module, plus one subprocess smoke test per CLI (`node cli.ts --help`) so the entry block itself stays under test.
