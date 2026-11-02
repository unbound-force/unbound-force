<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the same group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
  Phase checkpoints (MP-018): every phase below is terminated
  by the /uf.unleash implement-phase checkpoint — a hard-gate
  pre-flight run of the full detected CI/local suite
  (.opencode/commands/uf.unleash.md Step 5) — and the next
  phase MUST NOT begin while that checkpoint fails. Tasks
  whose text carries an explicit verification command express
  the same gate in-plan.
-->

## 1. Managed-region infrastructure (foundational)

All tasks modify `internal/scaffold/scaffold.go` or its test
files and build on each other — run sequentially, no `[P]`.

- [x] 1.1 Add `extractManagedRegion(content, marker,
  terminator string)` to `internal/scaffold/scaffold.go`
  (design D7): returns the byte range of the managed block —
  the first line equal to the marker through the first
  subsequent line equal to the terminator — as
  `(start, end int, ok bool)`, with `ok` false when either
  delimiter is absent. Add table-driven tests in
  `internal/scaffold/scaffold_test.go` covering: region
  found; marker absent; marker present with no terminator
  (malformed); content before and after the region; marker
  appearing after a terminator line (still the first marker
  wins). Verify with
  `go test -race -count=1 -run TestExtractManagedRegion
  ./internal/scaffold/`.

- [x] 1.2 Rewrite `ensureGitignore`
  (`internal/scaffold/scaffold.go:1432`) onto the helper
  (design D7, SC-FR-014): dispositions in order — no marker
  → append the block (`created`, the existing path); marker +
  region byte-identical to `gitignoreBlock` → untouched
  (`unchanged`, today's "already configured"); marker +
  region differs → rewrite the region in place via
  `Options.WriteFile` (`upgraded`), preserving content before
  and after, never duplicating the marker; marker with no
  terminator → untouched (`skipped` with detail naming the
  malformed region). Add disposition tests in
  `internal/scaffold/scaffold_test.go` (stale-block upgrade
  with trailing user content, current-block no-op, malformed
  skip, append path). Verify with
  `go test -race -count=1 -run 'TestEnsureGitignore'
  ./internal/scaffold/`.

- [x] 1.3 Add the closing sentinel to the stealth exclude
  block (`internal/scaffold/scaffold_stealth.go`, design D8):
  a `stealthExcludeEndMarker` constant; `writeGitInfoExclude`
  writes start marker + escaped canonical paths + sentinel
  and applies the same created / upgraded / unchanged /
  malformed-skip dispositions to its marker-delimited region;
  `appendNewUntrackedToExclude` keeps appending at end of
  file (after the sentinel, outside the region). Confirm by
  reading the code that `stealthExcludedPaths` still counts
  extra lines (it parses all non-comment lines) and that
  rollback (which restores pre-init exclude bytes) is
  unaffected. Add tests in
  `internal/scaffold/scaffold_stealth_test.go`: fresh write
  includes the sentinel; stale region upgrades in place with
  extra lines after the region preserved; legacy block
  without sentinel reports skipped with detail. Verify with
  `go test -race -count=1 -run 'TestWriteGitInfoExclude|
  TestStealth' ./internal/scaffold/`.

## 2. gitignoreBlock sync and drift pinning

Tasks 2.1–2.4 are ordered by dependency (the drift test needs
the new block and the aligned repo file) — run sequentially,
no `[P]`.

- [x] 2.1 Update `gitignoreBlock`
  (`internal/scaffold/scaffold.go:1399-1421`, design D6,
  SC-FR-013): add `.uf/sibling-clones/`, `.uf/sibling-cache/`,
  `.uf/feedback/`, and `.uf/progress/` as a workspace-state
  group beside the runtime-data group; move
  `.uf/config.yaml` and `.uf/dewey/config.yaml` into a
  per-developer group whose comment is pluralized to cover
  configuration alongside both override files. The first line
  (marker) and final line (`.unbound-force/` terminator) stay
  unchanged. `.gaze/` stays out.

- [x] 2.2 Sync the repo root `.gitignore`: rewrite the
  managed region (lines 34–61) to equal the new
  `gitignoreBlock` byte-for-byte — this removes the
  `.uf/dewey/config.yaml` duplicate (lines 40/47) and carries
  the pluralized comment. Leave the non-managed regions
  untouched in this task.

- [x] 2.3 Add the drift test (SC-FR-013, design D7) in
  `internal/scaffold/scaffold_test.go`: read the repo root
  `.gitignore` via `findProjectRoot(t)`, extract the managed
  region with `extractManagedRegion` (the same helper as
  task 1.1), and assert it equals `gitignoreBlock`
  byte-for-byte, failing with a diff naming `.gitignore`.
  Follow the `TestPluginPackageInputs_GitignoreTracks
  CanonicalInputs` pattern for repo-file access.

- [x] 2.4 Extend the representative-patterns test
  (`internal/scaffold/scaffold_test.go:8287`, the
  `TestEnsureGitignore_*` pattern list) with
  `.uf/config.yaml`, `.uf/dewey/config.yaml`,
  `.uf/sibling-clones/`, `.uf/sibling-cache/`,
  `.uf/feedback/`, `.uf/progress/`, and an assertion that
  `.gaze/` is absent from the written block (SC-FR-013).
  Verify the phase with
  `go test -race -count=1 -run 'TestEnsureGitignore|
  TestGitignoreBlock' ./internal/scaffold/`.

## 3. Override stub seeding

All tasks modify `internal/scaffold/scaffold.go` or
`scaffold_test.go` and build on each other — run sequentially,
no `[P]`.

- [x] 3.1 Add the `modelMatrixOverrideStub` and
  `reviewMatrixOverrideStub` Go string constants to
  `internal/scaffold/scaffold.go` (design D1/D2, SC-FR-011):
  comment-only content — a header naming the file and its
  purpose, the whitelisted paths (model-matrix:
  `profiles.*.model`, `profiles.*.variant`,
  `classifier.llm.*`; review-matrix:
  `profiles.<tier>.model`/`variant`), commented-out example
  entries, and a never-commit reminder. No uncommented YAML —
  the content parses to YAML null.

- [x] 3.2 Add `ensureOverrideStubs(opts *Options)
  []subToolResult` to `internal/scaffold/scaffold.go`
  (design D3/D4): for each stub, `MkdirAll` `.uf/` as needed
  and write the file only when it does not exist
  (create-only-if-absent, `Options.ReadFile`/`WriteFile`
  injection), reporting created or skipped per file; existing
  files are never read-modified.

- [x] 3.3 Wire seeding into `Run()` (`internal/scaffold/
  scaffold.go:354-364`): call `ensureOverrideStubs` in the
  non-stealth ensure-branch, after `ensureGitignore`, gated
  on `!opts.DivisorOnly` (SC-FR-011 skip surfaces:
  `--stealth` and `--divisor`).

- [x] 3.4 Add tests in `internal/scaffold/scaffold_test.go`
  (SC-FR-011): exact-content creation in a full `Run()`
  (both files equal their constants and appear in the
  summary); hand-edited override preserved byte-for-byte
  across a plain re-run AND a `--force` re-run (follow the
  `TestRun_ReviewerCapabilitiesPreservedOnRerun` pattern);
  `--stealth` and `--divisor` runs write neither file.
  Verify with
  `go test -race -count=1 -run 'TestOverrideStubs|
  TestRun_Override' ./internal/scaffold/`.

## 4. Overrides-only mode

Tasks 4.1–4.3 are dependency-ordered (the CLI flag needs the
`Options` field; the tests need both) — run sequentially, no
`[P]`.

- [x] 4.1 Add `OverridesOnly bool` to `Options`
  (`internal/scaffold/scaffold.go:56`) with GoDoc, and the
  early-exit path in `Run()` (design D5, SC-FR-012): after
  the Check-mode early return and before stealth setup, when
  `OverridesOnly` is set, run only `ensureGitignore` +
  `ensureOverrideStubs`, print the summary, and return — the
  asset walk, plugin activation, sub-tools, AGENTS.md
  sections, and migrations never run.

- [x] 4.2 Add the `--overrides` flag to
  `cmd/unbound-force/main.go` (`newInitCmd`, `initParams`,
  `runInit`): flag description "Create only the per-developer
  override stubs and the managed .gitignore block"; conflict
  validation failing before any filesystem write with an
  error naming `--overrides` and the conflicting flag for
  `--divisor`, `--force`, `--check`, and `--stealth` (design
  D5 — the issue's three plus `--stealth`); extend the Long
  help text with the mode paragraph and exclusions.

- [x] 4.3 Add tests (SC-FR-012): in
  `cmd/unbound-force/main_test.go`, the conflict matrix (all
  four combinations exit non-zero with the flags named, via
  the `runInit`/`TestInitCmd_*` patterns, asserting the
  target directory is untouched) and help-text coverage
  (follow `TestInitCmd_StealthHelpText`); in
  `internal/scaffold/scaffold_test.go`, `OverridesOnly`
  isolation — an empty temp dir receives exactly
  `.uf/model-matrix.override.yaml`,
  `.uf/review-matrix.override.yaml`, and a `.gitignore` with
  the managed block, and nothing else; re-run is a no-op
  (idempotent); a pre-existing override is preserved.
  Verify with
  `go test -race -count=1 -run 'TestRunInit|
  TestInitCmd|TestOverridesOnly' ./cmd/unbound-force/
  ./internal/scaffold/`.

## 5. OpenCode directory ignore sub-tool

Tasks 5.1–5.3 are dependency-ordered — run sequentially, no
`[P]`.

- [x] 5.1 Add the `opencodeGitignore` constant and its
  ensure-style writer to `internal/scaffold/scaffold.go`
  (design D9, SC-FR-015): content exactly `node_modules`,
  `bun.lock`, `.vitest/`, `coverage/` (one per line); merge
  semantics — file absent → write the canonical lines; file
  present → append only canonical lines missing from it,
  retaining every existing user line; wire it into `Run()`
  as a sub-tool result independent of the plugin-activation
  result, running in full and `--divisor` mode, skipped
  under `--stealth` and in `--overrides` mode.

- [x] 5.2 Update the canonical `.opencode/.gitignore` to the
  four canonical lines and the
  `TestPluginPackageInputs_GitignoreTracksCanonicalInputs`
  want-string (`internal/scaffold/scaffold_test.go:1626`);
  drop the now-redundant root `.opencode/.vitest/`
  (`.gitignore:32`) and `coverage/` (`.gitignore:77`) lines
  from the repo root `.gitignore` (the scoped file covers
  them).

- [x] 5.3 Add tests in `internal/scaffold/scaffold_test.go`
  (SC-FR-015): fresh full-mode `Run()` creates
  `.opencode/.gitignore` with exactly the canonical lines; an
  existing file with a user line and a partial canonical set
  keeps the user line and gains only the missing canonical
  lines; the file is written when plugin activation fails
  (stub a failing activation seam or assert independence
  from the activation result's disposition); DivisorOnly
  deploys it, `--stealth` and `--overrides` do not.
  Verify the phase with
  `go test -race -count=1 ./internal/scaffold/
  ./cmd/unbound-force/`.

## 6. Documentation

- [x] 6.1 [P] Update `docs/cli-reference.md` `uf init`
  section: add the `--overrides` row to the flag table
  (bool, default `false`, "Create only the per-developer
  override stubs and the managed .gitignore block; mutually
  exclusive with --divisor, --force, --check, and --stealth")
  and an example invocation with its expected output shape;
  correct the `--force` row to "Overwrite existing
  scaffold-managed files; per-developer override files are
  never modified" (SC-FR-011); extend the ownership intro
  to three classes (scaffold-managed, project-owned,
  per-developer: seeded once, never modified, never reset
  by `--force`).

- [x] 6.2 [P] Update `docs/configuration.md` (SC-FR-010
  reversal ripple): the four "never seeded" passages —
  the review-matrix table row (line 275), the review-council
  prose (lines 286-288), the model-matrix "deliberately not
  deployed" passage (lines 471-474), and the ownership table
  row (line 487) — now state the create-only-if-absent
  seeding contract: seeded once as a comment-only stub on
  first `uf init` (or `uf init --overrides`), never modified
  afterwards, not even by `--force`. Leave the line-406
  gitignore-entry mention accurate as written.

- [x] 6.3 [P] Add a `CHANGELOG.md` entry under Unreleased →
  Added for init-override-stubs-gitignore-sync: override
  stub seeding with `--overrides` mode and `--force`
  preservation, the gitignore block sync with in-place
  upgrade semantics, and the `.opencode/.gitignore` sub-tool
  (mention the SC-FR-010 contract amendment).

- [x] 6.4 [P] Refresh `docs/architecture.md` "Managed files"
  section (~lines 288-291): the `.gitignore` managed block
  is upgraded in place when it differs (SC-FR-014), not
  merely appended; add the `.opencode/.gitignore` sub-tool
  entry to the managed-files list.

## 7. Bookkeeping and verification

- [x] 7.1 File BOTH documentation-gate tracking issues for
  this user-facing change (`uf init` behavior and flags):
  the repo-local issue against this repo, referencing
  issues #903 and #902, AND the cross-repo tracking issue
  in unbound-force/website per the constitution v1.4.0
  Cross-Repo Documentation rule (what changed -- `uf init
  --overrides`, override stub seeding, managed gitignore
  upgrade; why it matters; which website pages need
  updating). Record both issue numbers for the PR body.

- [x] 7.2 Run the full CI-parity gate (MP-023):
  `make check` — lint, `go test -race -count=1 ./...`,
  build, the scope-based coverage gate, and the TypeScript
  plugin suite (including its 94% statement coverage floor,
  which the scaffold test changes must not disturb — the
  plugin suite itself is untouched). All gates MUST pass
  without modification — any gate conflict is a
  stop-and-report per the AGENTS.md gatekeeping rule.

- [x] 7.3 Verify constitution alignment against the
  proposal's assessments: the stubs deploy no
  credential-shaped content (comment-only), every disposition
  is observable in the run summary, no governance or coverage
  gate was weakened (task 7.2 evidence), and the loaders are
  untouched. Record the verification outcome in the change's
  wrap-up notes.

- [x] 7.4 PR bookkeeping: the PR body MUST close both issues
  — `/uf.finale` emits `Closes #903` from
  `.openspec.yaml`'s `originating_issue`; ensure `Closes
  #902` is also present in the PR body, carrying both task
  7.1 documentation-issue references (repo-local and
  unbound-force/website), per the proposal header note.
