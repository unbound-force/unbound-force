## Why

Two test failures block CI and prevent Gaze from
running quality analysis:

1. **TestEmbeddedAssets_MatchSource** — 17 scaffold
   asset files are stale. Source files were updated
   but never copied to `internal/scaffold/assets/`.
2. **TestSetupRun_OllamaInstall** — test doesn't set
   `opts.GOOS`, causing a mismatch between the test
   expectation (Linux default) and the actual runtime
   behavior (macOS uses cask).

## What Changes

- Sync 17 stale scaffold assets
- Fix Ollama test to set `opts.GOOS = runtime.GOOS`

## Impact

- 17 scaffold asset copies (file sync, no content change)
- 1 test fix (1 line added)
- No logic changes
