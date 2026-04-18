## 1. Scaffold Asset Sync

- [x] 1.1 Sync 17 stale scaffold assets from canonical
  sources to internal/scaffold/assets/

## 2. Ollama Test Fix

- [x] 2.1 Add GOOS: runtime.GOOS to the Options struct
  in TestSetupRun_OllamaInstall in
  internal/setup/setup_test.go

## 3. Verification

- [x] 3.1 Run go test -race -count=1 ./... and verify
  all packages pass

<!-- spec-review: passed -->
<!-- code-review: passed -->
