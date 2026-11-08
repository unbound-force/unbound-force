---
tag: cli-tooling
author: jay-flowers
category: gotcha
created_at: 2026-11-08T19:10:35Z
identity: cli-tooling-20261108T191035-jay-flowers
tier: draft
---

`import.meta.main` is a Deno API and is undefined in Node before 24.2. In a repo whose plugin-suite engine floor is ^22.22.2 with CI on node 22, an `if (import.meta.main) runMain(main)` entry guard is dead code: the module loads, the CLI exits 0 with empty output, and every documented consumer silently no-ops (unbound-force PR 959, .opencode/lib/tool-inventory.ts:602 — the divisor-determinism persona's permission-allowlisted evidence commands included). The Node-compatible technique compares realpath(process.argv[1]) against fileURLToPath(import.meta.url) (instruction-lint.ts precedent); extract it as a shared `isMainModule()` helper beside `runMain` in the lib CLI shared module so entry guards cannot drift between CLIs. Vitest never executes entry guards — it imports the module — so the seam stays dead unless a subprocess smoke test spawns `node <cli> --help` and asserts non-empty stdout and the expected exit code on the CI node version. When the guard idiom changes, migrate any source-text detection keyed to it (tool-inventory's collectClis regex matched `import.meta.main` literals, including matches inside comments) in the same increment.
