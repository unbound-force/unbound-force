## 1. Speckit Workflow Swarm Skill

- [x] 1.1 Create `.opencode/skill/speckit-workflow/SKILL.md` with YAML frontmatter (`name: speckit-workflow`, `description`, `tags`). Content teaches the Swarm coordinator to: check for `tasks.md` before CASS decomposition, map phases to epics, interpret `[P]` markers as parallel-safe, interpret `[US?]` labels as cell metadata, respect phase dependency ordering, fall back to standard CASS when no `tasks.md` exists.

- [x] 1.2 Add `speckit-workflow/SKILL.md` to scaffold engine embedded assets at `internal/scaffold/assets/.opencode/skill/speckit-workflow/SKILL.md`.

- [x] 1.3 Update `isToolOwned()` in `internal/scaffold/scaffold.go` to classify the new skill file as tool-owned.

- [x] 1.4 Update `knownEmbeddedFiles` count in `internal/scaffold/scaffold_test.go` to include the new file.

- [x] 1.5 Run `go test ./internal/scaffold/...` to verify drift detection passes.

## 2. Cobalt-Crush Agent Update

- [x] 2.1 Read current `.opencode/agents/cobalt-crush-dev.md` and identify insertion point for new section.

- [x] 2.2 Add "Swarm Coordination" section to `cobalt-crush-dev.md` with: (a) file reservation protocol -- call `swarmmail_reserve()` before editing files when operating as a Swarm worker, (b) session lifecycle -- `hive_sync()` + `git push` before ending, "the plane is not landed until git push succeeds", (c) progress reporting -- `swarm_progress()` at milestones, (d) completion protocol -- `swarm_complete()` with `files_touched`.

- [x] 2.3 Verify the agent file's YAML frontmatter is still valid by running `unbound doctor` agent integrity check.

## 3. Ollama Health Check in Doctor

- [x] 3.1 Add `ollama` to the `coreToolSpecs` list in `internal/doctor/checks.go` with classification: optional, not required, not recommended. Include install hint `brew install ollama && ollama pull mxbai-embed-large` and install URL `https://ollama.com`.

- [x] 3.2 After the standard binary-in-PATH check for `ollama`, add a model check: when `ollama` is found, run `ollama list` via `opts.ExecCmd`, parse the output for a line containing `mxbai-embed-large`, and set the check result message accordingly ("mxbai-embed-large model ready" or install hint "ollama pull mxbai-embed-large").

- [x] 3.3 Write test `TestCheckOllama_InstalledWithModel` in `internal/doctor/doctor_test.go`. Inject ExecCmd returning `ollama list` output with `mxbai-embed-large` line. Verify Pass severity with model ready message.

- [x] 3.4 Write test `TestCheckOllama_InstalledWithoutModel` in `internal/doctor/doctor_test.go`. Inject ExecCmd returning `ollama list` output without the model. Verify Pass severity with install hint.

- [x] 3.5 Write test `TestCheckOllama_NotInstalled` in `internal/doctor/doctor_test.go`. Inject LookPath not finding `ollama`. Verify Pass severity (informational) with install hint.

## 4. Setup Ollama Guidance

- [x] 4.1 In `internal/setup/setup.go`, after printing the completion summary, check for `ollama` binary via `opts.LookPath`. If not found, print a "Tip" line: `Tip: Install Ollama for enhanced semantic memory:` followed by `brew install ollama && ollama pull mxbai-embed-large` and `(Without Ollama, semantic memory uses full-text search)`.

- [x] 4.2 Write test `TestSetupRun_OllamaTip` in `internal/setup/setup_test.go`. Inject LookPath not finding `ollama`. Verify output contains "Tip" and "ollama" in the completion summary.

- [x] 4.3 Write test `TestSetupRun_NoOllamaTip` in `internal/setup/setup_test.go`. Inject LookPath finding `ollama`. Verify output does NOT contain the Ollama tip.

## 5. Verification

- [x] 5.1 Run `go build ./...` to verify compilation.

- [x] 5.2 Run `go test -race -count=1 ./internal/doctor/... ./internal/setup/... ./internal/scaffold/...` to verify all tests pass.

- [x] 5.3 Run `go vet ./...` to verify no vet issues.

- [x] 5.4 Run `unbound doctor` manually to verify Ollama check appears in output and new skill passes integrity check.

- [x] 5.5 Verify constitution alignment: Autonomous Collaboration (skill communicates via tasks.md artifact), Composability First (Ollama optional, skill conditional), Observable Quality (check result in JSON output), Testability (all new checks use injected ExecCmd/LookPath).
