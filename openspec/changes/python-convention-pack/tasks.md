<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Create Convention Pack Files

- [ ] 1.1 [P] Create `.opencode/uf/packs/python.md` with
  YAML frontmatter (`pack_id: python`, `language: Python`,
  `version: 1.0.0`) and 7 sections: Coding Style,
  Architectural Patterns, Security Checks, Testing
  Conventions, Type Annotations, Documentation
  Requirements, Custom Rules (empty placeholder). Follow
  the format of `go.md` and `typescript.md` exactly. Use
  RFC 2119 severity indicators and numbered rule
  identifiers (CS-NNN, AP-NNN, SC-NNN, TC-NNN, TA-NNN,
  DR-NNN).

- [ ] 1.2 [P] Create `.opencode/uf/packs/python-custom.md`
  with YAML frontmatter (`pack_id: python-custom`,
  `language: Python`, `version: 1.0.0`) and empty template
  matching `go-custom.md` format.

## 2. Create Embedded Asset Copies

Depends on Group 1. The embedded copies must be
byte-identical to the canonical sources.

- [ ] 2.1 Copy `.opencode/uf/packs/python.md` to
  `internal/scaffold/assets/opencode/uf/packs/python.md`.
  Verify byte-identical with diff.

- [ ] 2.2 Copy `.opencode/uf/packs/python-custom.md` to
  `internal/scaffold/assets/opencode/uf/packs/python-custom.md`.
  Verify byte-identical with diff.

## 3. Update Scaffold Language Detection

- [ ] 3.1 Update `detectLang()` in
  `internal/scaffold/scaffold.go`: add 5 new marker entries
  after the existing `pyproject.toml` entry:
  `setup.py` -> python, `setup.cfg` -> python,
  `requirements.txt` -> python, `tox.ini` -> python,
  `Pipfile` -> python. Preserve existing marker order
  (go.mod, tsconfig.json, package.json, pyproject.toml
  remain unchanged; new entries before Cargo.toml).

## 4. Update Scaffold Tests

All changes in `internal/scaffold/scaffold_test.go`.

- [ ] 4.1 Update `expectedAssetPaths`: add 2 new entries
  in sorted position:
  `"opencode/uf/packs/python-custom.md"` and
  `"opencode/uf/packs/python.md"`.

- [ ] 4.2 Update `TestDetectLang`: add test cases for the
  5 new Python markers (`setup.py`, `setup.cfg`,
  `requirements.txt`, `tox.ini`, `Pipfile`). Add a
  priority test case verifying `pyproject.toml` takes
  precedence when multiple Python markers exist. Add a
  test verifying Go still takes priority over Python
  markers.

- [ ] 4.3 Update `TestShouldDeployPack`: add test cases
  for Python pack deployment:
  `python.md` with lang=python -> true,
  `python-custom.md` with lang=python -> true,
  `python.md` with lang=go -> false,
  `python-custom.md` with lang=go -> false,
  `go.md` with lang=python -> false.

- [ ] 4.4 Update `TestIsToolOwned`: add test cases:
  `opencode/uf/packs/python.md` -> true (canonical),
  `opencode/uf/packs/python-custom.md` -> false (custom).

- [ ] 4.5 Update `TestIsDivisorAsset`: add test case:
  `opencode/uf/packs/python.md` -> true (convention packs
  are Divisor assets).

## 5. Add Python Doctor Checks

- [ ] 5.1 Add `checkPythonTools()` function to
  `internal/doctor/checks.go` returning a `CheckGroup`
  named "Python Tools". Checks 9 tool categories:
  python3 (required), pip/uv (recommended, pass if either
  found), pytest (required), formatter (recommended, pass
  if `black` or `ruff` found), linter (recommended, pass
  if `flake8` or `ruff` found), import sorter (recommended,
  pass if `isort` or `ruff` found), security scanner
  (recommended, pass if `bandit` or `ruff` found), mypy
  (optional), tox (optional). For categories with
  alternatives, check all binaries via `LookPath` and pass
  if any is found. Use the existing `checkOneTool()` for
  single-binary checks; implement a small
  `checkAnyTool()` helper for multi-binary categories.

- [ ] 5.2 Add `isPythonProject(targetDir string) bool` to
  `internal/doctor/checks.go` that checks for Python marker
  files (`pyproject.toml`, `setup.py`, `setup.cfg`,
  `requirements.txt`, `tox.ini`, `Pipfile`) via `os.Stat`.
  Returns true if any marker exists. This duplicates the
  marker list from `scaffold.detectLang()` to avoid a
  cross-package import (per design D6).

- [ ] 5.3 Update `Run()` in `internal/doctor/doctor.go` to
  conditionally include the Python tools check group. Call
  `isPythonProject(opts.TargetDir)` and include
  `checkPythonTools()` only when it returns true. Follow
  the DevPod conditional inclusion pattern.

## 6. Add Doctor Tests

- [ ] 6.1 Add `TestCheckPythonTools_AllPresent` to
  `internal/doctor/doctor_test.go`: inject `LookPath` that
  finds all 9 Python tools, verify all results are Pass.

- [ ] 6.2 Add `TestCheckPythonTools_NoneMissing` to
  `internal/doctor/doctor_test.go`: inject `LookPath` that
  finds nothing, verify required tools are Fail,
  recommended are Warn, optional are Pass.

- [ ] 6.3 Add `TestCheckPythonTools_Skipped` to
  `internal/doctor/doctor_test.go`: verify Python tools
  group is NOT included when no Python marker files exist
  in the target directory.

- [ ] 6.4 Add `TestIsPythonProject` to
  `internal/doctor/doctor_test.go`: verify detection for
  each Python marker file individually (`pyproject.toml`,
  `setup.py`, `setup.cfg`, `requirements.txt`, `tox.ini`,
  `Pipfile`) and verify false when none exist.

## 7. Build, Test, and Verify

- [ ] 7.1 Run `go build ./...`. Fix any compilation errors.
- [ ] 7.2 Run `go vet ./...`. Fix any vet warnings.
- [ ] 7.3 Run `go test -race -count=1 ./...`. Fix any
  test failures. Pay specific attention to:
  - `TestAssetPaths_MatchExpected` (asset count)
  - `TestEmbeddedAssets_MatchSource` (drift detection)
  - `TestCanonicalSources_AreEmbedded` (reverse drift)
  - `TestDetectLang` (new Python markers)
  - `TestShouldDeployPack` (Python pack filtering)
- [ ] 7.4 Constitution alignment verification: confirm
  new pack files have YAML frontmatter (Observable
  Quality), detection is testable via TempDir
  (Testability), and no mandatory dependencies introduced
  (Composability First).
