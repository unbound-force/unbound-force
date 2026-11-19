---
tag: quickfetch-failed-run-logs-test-sets-tmpdir-t
author: jay-flowers
category: gotcha
created_at: 2026-11-15T00:52:05Z
identity: quickfetch-failed-run-logs-test-sets-tmpdir-t-20261115T005205-jay-flowers
tier: draft
---

Learning from branch quick/fetch-failed-run-logs-test-sets-tmpdir-t (issue #1051), 2026-11-14, category gotcha, topic process.env undefined coercion. Node's process.env coerces an assigned undefined into the literal string "undefined" because environment values are strings or absent, never undefined. In .opencode/test/fetch-failed-run-logs.test.ts, the test saved process.env.TMPDIR before overriding it and restored with plain assignment; on Linux, where TMPDIR is normally unset, the restore left the string "undefined" behind, so every later os.tmpdir() in that vitest process returned the relative path "undefined" and fetch_failed_run_logs wrote run-<id>-failed-logs.txt into <cwd>/.opencode/undefined/opencode/ inside the work tree — stray files that blocked gh-hud from returning checkouts to main. macOS always sets TMPDIR, so macOS CI never caught it. The fix branches on the saved value: delete process.env.TMPDIR when it started unset, assign only a real saved string, and assert TMPDIR is undefined again after the test when it started unset. General rule for any test that save/restores environment variables: restore unset-ness with delete, never with assignment of the saved value.
