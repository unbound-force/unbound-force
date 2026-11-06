---
tag: cli-tooling
author: jay-flowers
category: gotcha
created_at: 2026-11-06T18:25:01Z
identity: cli-tooling-20261106T182501-jay-flowers
tier: draft
---

When one change introduces two executable CLI tools into the same library directory, extract the shared CLI plumbing (root discovery, IO injection, subprocess wrappers, error formatting) into a shared module in the same increment. PR 959 added .opencode/lib/instruction-lint.ts and .opencode/lib/tool-inventory.ts with parallel findRoot/CliIO/git-wrapper helpers, and the copies diverged before merge: tool-inventory's findRoot used existsSync (a file named .opencode counts as a root marker, producing a silent "0 tools, 0 scripts" exit 0 on a bogus root) while instruction-lint used isDir (directory required — loud failure or correct ancestor root). Neither test suite pinned the divergent predicate, so the drift was invisible to CI until a reviewer diffed the two implementations by hand. Extraction with a regression test for the file-named-.opencode case closes both the duplication and the untested seam.
