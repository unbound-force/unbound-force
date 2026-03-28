## 1. Update Setup Code

- [x] 1.1 Update `installOllama()` in `internal/setup/setup.go`: change `"brew", "install", "ollama"` to `"brew", "install", "--cask", "ollama-app"` and update dry-run detail string
- [x] 1.2 Update `installOllama()` skip detail: change download link message to reference `ollama-app`

## 2. Update Doctor Hints

- [x] 2.1 Update `ollamaInstallHint()` in `internal/doctor/environ.go`: change `"brew install ollama"` to `"brew install --cask ollama-app"` in all hint strings (2 locations)

## 3. Update Tests

- [x] 3.1 Update `TestSetupRun_AllMissing` in `internal/setup/setup_test.go`: change expected command from `"brew install ollama"` to `"brew install --cask ollama-app"`
- [x] 3.2 Update `TestSetupRun_OllamaInstall` in `internal/setup/setup_test.go`: change assertion from `"brew install ollama"` to `"brew install --cask ollama-app"`
- [x] 3.3 Update `TestSetupRun_OllamaNoHomebrew` in `internal/setup/setup_test.go`: update assertion string
- [x] 3.4 Update `TestSetupRun_OllamaBrewFails` in `internal/setup/setup_test.go`: change error map key from `"brew install ollama"` to `"brew install --cask ollama-app"`

## 4. Verify

- [ ] 4.1 Run `go build ./...`
- [ ] 4.2 Run `go test -race -count=1 ./internal/setup/... ./internal/doctor/...`
- [ ] 4.3 Run `go test -race -count=1 ./...`
