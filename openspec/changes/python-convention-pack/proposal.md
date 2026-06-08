## Why

Unbound Force currently ships convention packs for Go and
TypeScript but not Python. The `scaffold.language` config
already lists `python` as a valid value, and `detectLang()`
recognizes `pyproject.toml`, but no Python convention pack
exists to deploy. Running `uf init --lang python` or
auto-detecting a Python project falls back to the default
pack, losing all language-specific review criteria.

This gap blocks onboarding Python projects (e.g., Django
monoliths, pydantic-ai services) to the UF toolchain.
Divisor persona agents cannot enforce
Python-specific coding standards, and `uf doctor` does not
verify Python prerequisites.

## What Changes

Add first-class Python support to the scaffold and doctor
systems:

1. A canonical `python.md` convention pack with Python-
   specific rules for coding style, architecture, security,
   testing, type annotations, and documentation.

2. An empty `python-custom.md` template for project-specific
   rule extensions (following the `go-custom.md` pattern).

3. Expanded language auto-detection to recognize additional
   Python project markers (`setup.py`, `setup.cfg`,
   `requirements.txt`, `tox.ini`, `Pipfile`).

4. Python-specific `uf doctor` checks for the Python
   toolchain (python3, pip/uv, pytest, black, flake8,
   isort, bandit, mypy, tox).

## Capabilities

### New Capabilities
- `python.md`: Canonical Python convention pack deployable
  by `uf init` when the detected or configured language
  is `python`. Covers 7 sections: Coding Style,
  Architectural Patterns, Security Checks, Testing
  Conventions, Type Annotations, Documentation
  Requirements, and Custom Rules (empty placeholder).
- `python-custom.md`: Empty template for project-specific
  Python conventions, following the established custom pack
  pattern.
- Python doctor checks: `uf doctor` validates Python
  toolchain prerequisites when the project language is
  detected as Python (9 tool category checks with required /
  recommended / optional severity levels). Tool-agnostic
  where alternatives exist (e.g., passes if either `black`
  or `ruff` is found for formatting).
- Extended Python detection: `detectLang()` recognizes
  `setup.py`, `setup.cfg`, `requirements.txt`, `tox.ini`,
  and `Pipfile` in addition to the existing `pyproject.toml`
  marker.

### Modified Capabilities
- `detectLang()`: Additional Python marker files checked
  after `pyproject.toml` (which already exists). Priority
  order preserved -- Go and TypeScript markers still take
  precedence.
- `uf doctor`: New "Python Tools" check group added when
  a Python project is detected in the target directory.

### Removed Capabilities
- None.

## Impact

### Files Modified

**Go source** (production):
- `internal/scaffold/scaffold.go` -- `detectLang()` gains
  5 additional Python marker entries.
- `internal/doctor/checks.go` -- new `pythonToolSpecs`
  slice, `checkPythonTools()` function, and
  `isPythonProject()` detection function.
- `internal/doctor/doctor.go` -- `Run()` conditionally
  includes the Python tools check group when
  `isPythonProject()` returns true.

**Embedded assets** (new files):
- `internal/scaffold/assets/opencode/uf/packs/python.md`
- `internal/scaffold/assets/opencode/uf/packs/python-custom.md`

**Live Markdown** (canonical sources, new files):
- `.opencode/uf/packs/python.md`
- `.opencode/uf/packs/python-custom.md`

**Tests**:
- `internal/scaffold/scaffold_test.go` -- `expectedAssetPaths`
  updated, new test cases in `TestDetectLang`,
  `TestShouldDeployPack`, `TestIsToolOwned`,
  `TestIsDivisorAsset`.
- `internal/doctor/doctor_test.go` -- new tests for Python
  tool checks.

### Cross-Repo Impact
- None. This is additive -- no existing behavior changes.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

Convention packs are self-describing Markdown artifacts with
YAML frontmatter. The Python pack follows the same format as
Go and TypeScript packs. Divisor persona agents consume packs
dynamically at review time without runtime coupling. No
inter-hero communication changes.

### II. Composability First

**Assessment**: PASS

The Python pack is independently deployable. `uf init
--lang python` works regardless of whether Gaze, Dewey, or
any other hero is installed. The pack does not introduce
mandatory dependencies -- it documents recommended tools
(black, flake8, pytest, etc.) but does not require them to
be installed for the pack itself to function. Doctor checks
use the existing required/recommended/optional severity
model.

### III. Observable Quality

**Assessment**: PASS

The Python pack uses the same RFC 2119 severity indicators
([MUST], [SHOULD], [MAY]) and numbered rule identifiers
(CS-NNN, AP-NNN, etc.) as existing packs. Doctor check
results are reported through the existing `CheckGroup` /
`CheckResult` model with structured severity levels. All
output is machine-parseable via the `--format json` flag.

### IV. Testability

**Assessment**: PASS

All new code follows the established injectable-dependency
pattern. `detectLang()` is tested via `t.TempDir()` with
marker files. Pack deployment is tested through the existing
`Run()` integration tests. Doctor checks use injected
`LookPath` and `ExecCmd` functions. No external services
or shared mutable state required.
