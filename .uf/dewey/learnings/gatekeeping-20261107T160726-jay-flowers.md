---
tag: gatekeeping
author: jay-flowers
category: pattern
created_at: 2026-11-07T16:07:26Z
identity: gatekeeping-20261107T160726-jay-flowers
tier: draft
---

When a change newly tracks a previously gitignored lockfile (e.g. .opencode/package-lock.json for an exact-pinned devDependency), the lockfile content becomes visible to OSV-Scanner for the first time and previously invisible vulnerable transitive pins fail CI. The compliant remediation pattern from PR #958: bump only the vulnerable pins (zod 4.1.8→4.6.6 via override past @opencode-ai/plugin's exact pin, toml override 4.2.0→5.0.1, transitive vite 8.3.2→8.3.6), cite every GHSA id in a distinct CHANGELOG "Fixed" entry that names the causal link to the newly tracked lockfile, record upstream provenance (git heads + sha512 integrities, verification date) in the provenance testdata, update the exact-version test maps and scaffold mirrors in the same change, and keep pins as exact versions (never mutable tags). The AGENTS.md Gatekeeping rule's conditions for pin updates then audit as: documented security fix (GHSA ids), verified upstream immutable state (provenance file), gates still enforced (OSV-Scanner now passes, npm audit 0), rollback documented (previous versions named in the entry, revertible commit), reproducible validation (pin tests + plugin suite), and normal review/CI approval. Rollback-path documentation is implicit (previous versions named + git revert) rather than an explicit procedure — acceptable in substance, worth stating explicitly if challenged.
