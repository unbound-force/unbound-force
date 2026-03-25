## 1. Add installOllama Function

- [x] 1.1 Add `installOllama(opts *Options, env doctor.DetectedEnvironment) stepResult` function to `internal/setup/setup.go` following the exact pattern of `installGaze()`: check `LookPath("ollama")`, try `brew install ollama` if Homebrew available, skip with `https://ollama.com/download` link if not
- [x] 1.2 Insert `installOllama()` call in the `Run()` function between the Swarm steps and the Dewey step (after step 8 `.hive/`, before step 9 Dewey). Update step number comments accordingly.

## 2. Remove Ollama Tip

- [x] 2.1 Remove the "Tip: Install Ollama for enhanced semantic memory" block from `internal/setup/setup.go` (the `if _, ollamaErr := opts.LookPath("ollama"); ollamaErr != nil` block around lines 253-258) since Ollama is now installed automatically

## 3. Update Tests

- [x] 3.1 Update `TestSetupRun_AllMissing` in `internal/setup/setup_test.go`: add `"brew install ollama"` to the expected commands list, positioned after the swarm steps and before Dewey steps
- [x] 3.2 Update `TestSetupRun_AllPresent` (already has `"ollama"` in LookPath stub at line 175 -- no change needed) in `internal/setup/setup_test.go`: add `"ollama"` to the LookPath stub so it's detected as already installed

## 4. Verify

- [x] 4.1 Run `go build ./...` to verify compilation
- [x] 4.2 Run `go test -race -count=1 ./internal/setup/...` to verify setup tests pass (also fixed TestSetupRun_OllamaTip → TestSetupRun_OllamaInstall)
- [x] 4.3 Run `go test -race -count=1 ./...` to verify full test suite passes (16/16)
