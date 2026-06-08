## Context

The Unbound Force scaffold engine deploys language-specific
convention packs via `uf init`. The `shouldDeployPack()`
function filters packs by the resolved language, and
`collectDeployedPacks()` builds the list of pack filenames
for cross-tool bridge files (AGENTS.md, CLAUDE.md,
.cursorrules). Language detection in `detectLang()` checks
for well-known marker files (`go.mod`, `tsconfig.json`,
`package.json`, `pyproject.toml`, `Cargo.toml`).

Convention packs exist for Go (`go.md`) and TypeScript
(`typescript.md`). Python detection via `pyproject.toml`
already returns `"python"`, but no `python.md` pack exists
in the embedded assets. The result: `uf init` on a Python
project skips language-specific pack deployment and
cross-tool bridge files list only default packs.

The doctor system (`internal/doctor/`) checks tool
prerequisites via `coreToolSpecs` and dedicated check
functions. Tool checks are Go-centric (go, opencode, gaze,
node, gh, replicator, ollama, podman). No Python-specific
tool checks exist.

See `proposal.md` for motivation and constitution alignment.

## Goals / Non-Goals

### Goals
- `uf init --lang python` (or auto-detection) deploys
  `python.md` and `python-custom.md` convention packs
- `uf init --divisor --lang python` deploys the Python
  pack in Divisor-only mode
- Cross-tool bridge files (AGENTS.md, CLAUDE.md,
  .cursorrules) list Python packs when language is Python
- `detectLang()` recognizes additional Python markers
  (`setup.py`, `setup.cfg`, `requirements.txt`, `tox.ini`,
  `Pipfile`) beyond the existing `pyproject.toml`
- `uf doctor` checks Python toolchain prerequisites when
  the project language is Python
- All existing Go and TypeScript behavior is unchanged

### Non-Goals
- Python-specific Gaze integration (Snake Eyes -- separate
  project in zero-dot-force labs)
- Python project scaffolding beyond convention packs (e.g.,
  pyproject.toml generation, virtualenv setup)
- Rust convention pack (placeholder detection exists but
  pack creation is a separate change)
- SDEngine-specific custom rules (those go in
  `python-custom.md` in the SDEngine repo, not here)

## Decisions

### D1: Pack format matches Go and TypeScript exactly

The Python pack uses the same structure as `go.md` and
`typescript.md`: YAML frontmatter (`pack_id`, `language`,
`version`), followed by sections for Coding Style,
Architectural Patterns, Security Checks, Testing
Conventions, Documentation Requirements, and Custom Rules.

Additionally, a **Type Annotations** section is added
between Testing Conventions and Documentation Requirements.
This is Python-specific -- Go has type safety built in,
and TypeScript's type system is covered in Coding Style.

**Rationale**: Consistency with existing packs. Divisor
persona agents parse packs by section heading. A new
section heading is safe because agents match on content,
not on a hardcoded section list.

### D2: Rule numbering follows established convention

Rules use the same prefix scheme as Go and TypeScript:
CS-NNN (Coding Style), AP-NNN (Architectural Patterns),
SC-NNN (Security Checks), TC-NNN (Testing Conventions),
TA-NNN (Type Annotations -- new prefix), DR-NNN
(Documentation Requirements).

**Rationale**: Consistent numbering enables cross-pack
references (e.g., "per CS-001") and machine parsing of
rule identifiers.

### D3: Tool rules are outcome-focused, not tool-specific

The Python ecosystem is undergoing a tooling consolidation.
Ruff is rapidly replacing black + flake8 + isort as a
unified formatter and linter. Projects like aegis-ai use
ruff exclusively; projects like SDEngine use the traditional
trio. A convention pack that mandates specific tools (e.g.,
"MUST use black") would create false negatives for projects
using equivalent modern alternatives.

Rules that reference tooling are written as outcome
requirements with tool examples:

- CS-001: "MUST be auto-formatted consistently" (black,
  ruff format, or equivalent)
- CS-002: "imports MUST be auto-sorted" (isort, ruff with
  isort rules, or equivalent)
- CS-003: "MUST pass a linter" (flake8, ruff check, or
  equivalent)
- SC-005: "MUST run security static analysis" (bandit,
  ruff S rules, or equivalent)

**Rationale**: Convention packs should be durable. Coupling
rules to specific tools creates maintenance burden as the
ecosystem evolves. Outcome-focused rules evaluate code
quality, not tool preferences. Projects that need to
mandate a specific tool (e.g., "MUST use ruff, not flake8")
can do so in `python-custom.md`.

This also affects doctor checks (D5): the doctor should
check for *any* of the equivalent tools, not mandate a
single one.

### D4: Detection priority preserves existing order

New Python markers are added after `pyproject.toml` in the
`detectLang()` marker list. The full priority order becomes:

1. `go.mod` -> go
2. `tsconfig.json` -> typescript
3. `package.json` -> typescript
4. `pyproject.toml` -> python
5. `setup.py` -> python
6. `setup.cfg` -> python
7. `requirements.txt` -> python
8. `tox.ini` -> python
9. `Pipfile` -> python
10. `Cargo.toml` -> rust

**Rationale**: `pyproject.toml` is the modern standard and
should be checked first. Legacy markers follow in decreasing
specificity. Existing Go/TypeScript/Rust order is unchanged.

### D5: Doctor checks use a separate check group

Python tool checks are implemented as a new `CheckGroup`
named "Python Tools", conditionally included when the
detected language is `python`. This parallels how the
DevPod check group is conditionally included.

Per D3, doctor checks for formatting/linting/security tools
check for *any* of the equivalent tools. If the project has
ruff OR (black + flake8 + isort), the check passes. The
check only warns when *none* of the equivalent tools are
found.

Tool checks and their severity:

| Check | Severity | Binaries checked |
|-------|----------|------------------|
| python3 | required | `python3` |
| pip/uv | recommended | `pip` or `uv` |
| pytest | required | `pytest` |
| formatter | recommended | `black` or `ruff` |
| linter | recommended | `flake8` or `ruff` |
| import sorter | recommended | `isort` or `ruff` |
| security scanner | recommended | `bandit` or `ruff` |
| mypy | optional | `mypy` |
| tox | optional | `tox` |

**Rationale**: `python3` and `pytest` are required because
the convention pack mandates them ([MUST] rules). Other
tools are recommended or optional because they enforce
[SHOULD] rules. The severity model matches the existing
`required` / `recommended` / `optional` pattern in
`checkOneTool()`. Checking for tool alternatives (e.g.,
ruff as a substitute for flake8) avoids false warnings on
modern toolchains per D3.

### D6: Doctor duplicates language detection inline

The doctor system does not currently detect language. To
gate the Python tools check group, `checkPythonTools()`
checks for the same Python marker files that
`scaffold.detectLang()` uses (`pyproject.toml`, `setup.py`,
`setup.cfg`, `requirements.txt`, `tox.ini`, `Pipfile`).
The detection is duplicated rather than importing
`internal/scaffold` -- the doctor package should not
depend on the scaffold package for a 10-line file-existence
check.

The function `isPythonProject(targetDir string) bool`
returns true if any Python marker file exists. The check
group is included only when `isPythonProject()` returns
true.

**Rationale**: Avoids a cross-package dependency for
trivial logic. The marker file list is small and stable.
If the lists diverge, the scaffold drift tests will
surface the inconsistency because `uf init` will deploy
Python packs but `uf doctor` will not check Python tools
(or vice versa). Users can also suppress checks via
`doctor.skip` config.

### D7: SC-008 subprocess rule scoped to call sites

The original draft of SC-008 said "Use `subprocess.run()`
with `shell=False`. Never pass user input to `shell=True`
subprocess calls." This is too absolute -- real-world
codebases (e.g., SDEngine) have command execution utilities
that use `shell=True` internally with controlled, non-user-
supplied command strings. Requiring `shell=False` everywhere
would force rewrites of foundational utilities with no
security benefit when the inputs are developer-controlled.

The revised rule narrows the prohibition:
- [MUST] Never pass **user-supplied or external input** to
  `shell=True` subprocess calls.
- [SHOULD] Prefer `subprocess.run()` with list arguments
  and `shell=False` for new code.
- Internal utility wrappers that use `shell=True` with
  hardcoded or developer-controlled commands are acceptable
  when documented.

**Rationale**: The security risk is shell injection from
untrusted input, not the `shell=True` flag itself. A blanket
prohibition would create friction for onboarding without
improving security posture. SDEngine's `Cmd.run(shell=True)`
is used with fixed command strings -- the risk is low.
Projects that need stricter enforcement can override via
`python-custom.md`.

### D8: python-custom.md is an empty template

The custom pack template contains only the YAML frontmatter
and placeholder heading, identical in structure to
`go-custom.md`. No default custom rules are included.

**Rationale**: Custom rules are project-specific. Shipping
default custom rules would violate the pack ownership model
(canonical packs are tool-owned, custom packs are
user-owned).

## Risks / Trade-offs

### R1: Convention pack rules may not fit all Python projects

The Python pack mandates pytest (not unittest), requires
auto-formatting and linting, and recommends specific
architectural patterns. Projects using niche frameworks
(e.g., Twisted, Tornado) or unconventional test runners
may need to override rules in `python-custom.md`.

**Mitigation**: Tool-specific rules use outcome-focused
language per D3, so the pack is compatible with both
traditional (black+flake8+isort) and modern (ruff)
toolchains. Custom packs can override any rule. This
matches how Go projects can customize the Go pack.

### R2: Doctor checks may produce false warnings

A valid Python project may not have all recommended tools
installed globally (e.g., they live in a virtualenv).

**Mitigation**: Doctor checks use `LookPath` which
searches PATH. Activated virtualenvs add their bin to
PATH. Users can suppress checks via `doctor.skip` or
`doctor.tools` severity overrides in `.uf/config.yaml`.

### R3: Additional Python markers may cause false detection

A project with both `go.mod` and `requirements.txt` (e.g.,
a Go project with Python tooling scripts) will be detected
as Go due to priority order, which is correct. But a
project with only `requirements.txt` and no clear primary
language will be detected as Python.

**Mitigation**: Users can set `scaffold.language` explicitly
in `.uf/config.yaml` to override auto-detection. The
detection priority order is documented in this design.
