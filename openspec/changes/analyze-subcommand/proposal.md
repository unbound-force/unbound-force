## Why

Adopting unbound-force requires knowing what a repository
already has and what it needs. Today, maintainers must
manually check for a constitution, lint configs, convention
packs, review agents, CI workflows, and AGENTS.md — then
figure out which `uf init` flags to use. This is error-
prone and discourages adoption.

A diagnostic subcommand that inspects the repository and
produces actionable recommendations would eliminate
guesswork and standardize the onboarding process.

## What Changes

### New `uf analyze` subcommand

Scans the current repository and reports adoption status
with recommendations.

**Detection checks:**

| Check | How | Status |
|---|---|---|
| Language | `go.mod`, `pyproject.toml`, `tsconfig.json`, `Cargo.toml` | Detected / Unknown |
| Constitution | `.specify/memory/constitution.md` exists | Found (version) / Missing |
| Convention packs | `.agents/packs/*.md` exists | Found (count) / Missing |
| Review agents | `.opencode/agents/divisor-*` exists | Found (count) / Missing |
| Lint config | `.golangci.yml`, `ruff.toml`, `.yamllint.yml` | Found / Missing |
| CI workflows | `.github/workflows/ci_*.yml` | Found (count) / Missing |
| AGENTS.md | Root file exists | Found / Missing |
| PR review command | `.opencode/command/review_pr.md` | Found / Missing |
| Dewey config | `.uf/dewey/config.yaml` | Found / Not configured |

**Output format (terminal):**

```
$ uf analyze

Repository: complyctl
Language:   Go (go.mod)

  OK  Constitution       .specify/memory/constitution.md (v1.2.0)
  OK  Lint config        .golangci.yml
  OK  CI workflows       3 found
  OK  AGENTS.md          found
  OK  PR review command  .opencode/command/review_pr.md
  --  Convention packs   not found
  --  Review agents      not found
  --  Dewey              not configured

Recommendations:
  1. Run: uf init --packs-only --lang go
  2. Commit: .agents/packs/
  3. (Optional) Run: uf init --divisor --lang go
     for review council agents

Run 'uf analyze --json' for machine-readable output.
Run 'uf analyze --apply' to execute recommendations.
```

**JSON output** (`--json` flag): same data as structured
JSON for CI consumption or tooling integration.

**Apply mode** (`--apply` flag): executes the
recommendations automatically (runs `uf init` with the
appropriate flags). Requires confirmation unless
`--yes` is also provided.

## Capabilities

### New Capabilities

- `uf analyze`: Repository diagnostic and recommendation
  engine.
- `--json` flag: Machine-readable output for tooling.
- `--apply` flag: Execute recommendations automatically.

### Modified Capabilities

None. This is a new subcommand with no changes to
existing functionality.

### Removed Capabilities

None.

## Impact

- New Go source files in `cmd/` or `internal/analyze/`.
- New CLI subcommand registered in the root command.
- Reuses existing `detectLang` function from scaffold.
- No changes to existing commands or scaffold behavior.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

`uf analyze` produces a self-describing diagnostic
report. The `--json` flag ensures machine-parseable
output for tooling consumption.

### II. Composability First

**Assessment**: PASS

`uf analyze` works independently of `uf init`. It
inspects without modifying. The `--apply` flag is
optional and requires confirmation.

### III. Observable Quality

**Assessment**: PASS

The output includes specific file paths, versions, and
counts — all verifiable. JSON output enables programmatic
validation.

### IV. Testability

**Assessment**: PASS

Each detection check is a pure function (file existence,
content parsing). Testable with filesystem fixtures.
