## 1. Stray File Cleanup

- [x] 1.1 Delete `internal/scaffold/.opencode/` directory
  (31 stray files from tools running in wrong directory)
- [x] 1.2 Delete `internal/scaffold/.uf/` directory
  (7 stray files)
- [x] 1.3 Delete `cmd/unbound-force/.opencode/` directory
  (31 stray files)
- [x] 1.4 Delete `cmd/unbound-force/.uf/` directory
  (5 stray files)

## 2. Stray Prevention

- [x] 2.1 Add `.gitignore` patterns to prevent stray
  `.opencode/` and `.uf/` directories inside
  subdirectories: `cmd/**/.opencode/`, `cmd/**/.uf/`,
  `internal/**/.opencode/`, `internal/**/.uf/`

## 3. Externalize Speckit Commands

- [x] 3.1 Remove 9 `speckit.*.md` files from
  `internal/scaffold/assets/opencode/command/`:
  speckit.specify.md, speckit.clarify.md,
  speckit.plan.md, speckit.tasks.md,
  speckit.analyze.md, speckit.checklist.md,
  speckit.implement.md, speckit.constitution.md,
  speckit.taskstoissues.md
- [x] 3.2 Add 9 entries to `knownNonEmbeddedFiles` in
  `internal/scaffold/scaffold_test.go`
- [x] 3.3 Update `expectedAssetPaths` in
  `internal/scaffold/scaffold_test.go` — remove the
  9 speckit command entries, update count comment
- [x] 3.4 Update file count assertion in
  `cmd/unbound-force/main_test.go`
- [x] 3.5 Run `go build ./...` and
  `go test -race -count=1 ./internal/scaffold/...
  ./cmd/unbound-force/...` to verify

## 4. /uf-init Customizations

- [x] 4.1 Add "Speckit Custom Commands" section to
  `.opencode/command/uf-init.md` — instructions for
  the AI agent to create the 4 UF-custom speckit
  commands (speckit.analyze.md, speckit.checklist.md,
  speckit.clarify.md, speckit.taskstoissues.md) if
  they do not exist. Include the full content for
  each command inline.
- [x] 4.2 Add "Speckit Command Guardrails" section to
  `.opencode/command/uf-init.md` — instructions for
  the AI agent to inject a `## Guardrails` section
  into each of the 9 `speckit.*.md` files if not
  already present. Idempotent (check for existing
  section before appending).
- [x] 4.3 Add "Speckit UF Customizations" section to
  `.opencode/command/uf-init.md` — instructions for
  the AI agent to inject UF-specific content (Dewey
  integration, constitution references, review
  council integration) into the upstream speckit
  commands.
- [x] 4.4 Copy updated `.opencode/command/uf-init.md`
  to `internal/scaffold/assets/opencode/command/`

## 5. Code Review Marker (/unleash)

- [x] 5.1 Add `<!-- code-review: passed -->` marker
  write to Step 6 (code review) in
  `.opencode/command/unleash.md` — write after review
  council approves, append at end of tasks.md
- [x] 5.2 Update Step 2 (resumability detection) in
  `.opencode/command/unleash.md` — add item 6 checking
  for `<!-- code-review: passed -->` marker, remove
  the implicit CI-based code review detection
- [x] 5.3 Copy updated `.opencode/command/unleash.md`
  to `internal/scaffold/assets/opencode/command/`

## 6. Constitution + AGENTS.md

- [x] 6.1 Add "Phase Discipline" bullet to Development
  Workflow section in `.specify/memory/constitution.md`:
  each pipeline phase MUST produce only its designated
  artifacts, implementation code MUST NOT be written
  during specification phases
- [x] 6.2 Add "Workflow Phase Boundaries" subsection to
  Behavioral Constraints in `AGENTS.md` mapping phases
  to allowed outputs: specify/clarify/plan/tasks/
  analyze/checklist → spec artifacts only; implement →
  source code; review → findings and minor fixes

## 7. Documentation + Verification

- [x] 7.1 Add Recent Changes entry to `AGENTS.md`
- [x] 7.2 Run `go build ./...` to verify clean
  compilation
- [x] 7.3 Run `go test -race -count=1
  -run "TestAssetPaths|TestRunInit_FreshDir|
  TestEmbeddedAssets_MatchSource|
  TestScaffoldOutput_NoOldPathReferences"
  ./internal/scaffold/... ./cmd/unbound-force/...`
  to verify scaffold tests pass
- [x] 7.4 Verify no stray files remain:
  `find internal/ cmd/ -name ".opencode" -type d` and
  `find internal/ cmd/ -name ".uf" -type d` return
  empty
