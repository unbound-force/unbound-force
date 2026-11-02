---
tag: review-tooling
author: jay-flowers
category: gotcha
created_at: 2026-11-02T21:14:11Z
identity: review-tooling-20261102T211411-jay-flowers
tier: draft
---

The glob tool silently excludes dot-directories and their contents from both exact relative patterns (`.opencode/test/helpers/x.ts`) and wildcard patterns (`**/x.ts`) — fast-glob's default non-dot behavior. During the fix-uf-finale-ci-poll-gh-schema spec review (2026-11-02), both pattern forms reported "No files found" for `.opencode/test/helpers/gh-spawn-mock.ts` while the file demonstrably existed (285 lines, confirmed by direct read) and the entire test suite imports and compiles against it. A reviewer trusting the glob result would have filed a false HIGH finding ("test helper missing; suite cannot compile; tasks 5.1/5.2 completion claims false"). Discipline: never report a file as missing based on a glob result alone when the path traverses a dot-directory (`.opencode/`, `.uf/`, `.github/`, `.specify/`) — confirm existence with a direct read first. Direct reads are the existence oracle; globs over hidden trees are not.
