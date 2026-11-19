---
tag: scaffold
author: jay-flowers
category: gotcha
created_at: 2026-11-16T12:18:26Z
identity: scaffold-20261116T121826-jay-flowers
tier: draft
as_of: "unbound-force@e451ac2821684173ef2593af3f6fe5b3b141fe43"
evidence:
  - "dewey:cmd/unbound-force/main.go"
  - "session:ufinit-cwd-incident"
---

`uf init` accepts no positional target directory — it always scaffolds into the current working directory (cmd/unbound-force/main.go resolves targetDir from cwd; the only target controls are flags like --force and --divisor). Running `unbound-force init /some/tmp/dir` silently initializes the repo you are standing in: it adds `<!-- scaffolded by uf vdev -->` version markers and whitespace shuffles to tool-owned .opencode/ files, which then trips the scaffold mirror parity tests (TestEmbeddedAssets_MatchSource, TestCommandContracts_CanonicalScaffoldParity, TestSCFR001) until the churned files are reverted. To scaffold a scratch target, cd into a temp dir first (or call scaffold.Run with Options.TargetDir directly, as the tests do). Also note TestRunInit leaks one artifact into the test binary's CWD (cmd/unbound-force/.github/workflows/dewey-distill.yml.example) — pre-existing test hygiene issue, delete it if it appears.
