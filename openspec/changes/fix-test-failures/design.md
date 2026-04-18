## Context

Scaffold drift accumulates when canonical files under
`.opencode/` and `openspec/` are modified but their
copies under `internal/scaffold/assets/` are not synced.
The TestEmbeddedAssets_MatchSource regression test
catches this.

The Ollama test uses `opts.GOOS` to determine the
expected brew command but never sets it. The `defaults()`
function in `Run()` sets `GOOS = runtime.GOOS`, but
in the test the Options struct is constructed directly
without calling `defaults()`.

## Decisions

### D1: Sync all 17 files

Mechanical copy from canonical source to scaffold asset.

### D2: Set GOOS in test

Add `GOOS: runtime.GOOS` to the test's Options struct
so the test expectation matches the runtime behavior
on both macOS and Linux.
