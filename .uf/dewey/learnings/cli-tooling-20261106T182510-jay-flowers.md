---
tag: cli-tooling
author: jay-flowers
category: gotcha
created_at: 2026-11-06T18:25:10Z
identity: cli-tooling-20261106T182510-jay-flowers
tier: draft
---

CLI plumbing copy-adapted between two tools born in the same PR drifts within the authoring window itself: in PR 959, tool-inventory.ts's findRoot used existsSync (accepting a file named .opencode as a root marker) while instruction-lint.ts's used a directory check, so the sibling CLIs could resolve different roots. Consolidate shared CLI constructs (root discovery, IO interfaces, git wrappers, error-detail formatting) into a shared .opencode/lib/ module in the same PR that introduces the second copy — that is the cheapest extraction point, because both files and their tests are already open in the diff and no out-of-diff consumers exist; a deferred follow-up merely re-opens just-added files. Scope the extraction to the constructs actually shared (verify each claimed duplicate — one item in the reviewer's list, listDir, existed in only one file) and land it with a test pinning the unified semantics, since the drifted edge case (.opencode as a file) was untested in both suites.
