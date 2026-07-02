# AGENTS.md

## Project Overview

Unbound Force is an organization of AI agent personas and roles
for a software agent swarm, themed as a superhero team. Each
hero is a repository in the [unbound-force](https://github.com/unbound-force)
GitHub organization. This repo (`unbound-force/unbound-force`)
is the meta/organizational repository -- it defines the team,
the org-level constitution, the architectural specs for all
heroes, and the shared standards that every hero repo must follow.

- **Type**: Meta repository (specifications, governance, standards)
- **Heroes**: Muti-Mind (PO), Cobalt-Crush (Dev), Gaze (Tester),
  The Divisor (Reviewer), Mx F (Manager) -- all implemented
- **Tooling**: [Speckit](https://github.com/github/spec-kit) +
  [OpenCode](https://opencode.ai) +
  [Replicator](https://github.com/unbound-force/replicator)
- **License**: Apache 2.0
- **Mission**: Engineers shift from manual coding to directing
  AI agents via specs and rules as the medium through which
  human intent is manifested into code.

Agent files do not hardcode a `model:` field. They inherit from
OpenCode's configuration hierarchy: project `opencode.json` >
user `~/.config/opencode/opencode.json` > built-in default.

## Build & Test Commands

```bash
# Build
make build
# or: go build ./...

# Run tests
make test
# or: go test -race -count=1 ./...

# Run all checks (lint, test, build)
make check

# Lint (vet + golangci-lint)
make lint
# or: go vet ./... && golangci-lint run
```

Always run tests with `-race -count=1`. CI enforces this.

### CI Workflow Structure

| Workflow | File | Purpose |
|---|---|---|
| Local CI | `ci_local.yml` | Build, test, coverage ratchets |
| CI Checks | `ci_checks.yml` | MegaLinter + commitlint |
| Security | `ci_security.yml` | OSV-Scanner, Trivy, Scorecards |
| Dependencies | `ci_dependencies.yml` | Dependency review + dependabot |
| CRAP Load | `ci_crapload.yml` | CRAP regression analysis |
| Release | `release.yml` | workflow_dispatch, GoReleaser + Cosign + Syft + Homebrew tap |
| Scheduled | `ci_scheduled.yml` | Daily OSV-Scanner + Scorecards |

## Project Structure

```text
unbound-force/
├── .specify/                         # Speckit framework (templates, scripts, memory)
├── .opencode/
│   ├── agents/                       # Hero persona agents (18 active)
│   ├── commands/                     # Slash commands (48 files)
│   ├── skill/                        # Swarm skills packages
│   ├── skills/                       # Additional skills packages
│   └── uf/packs/                     # Convention packs
├── cmd/unbound-force/                # Cobra CLI entry point
├── cmd/mutimind/                     # Muti-Mind backend CLI
├── cmd/mxf/                          # Mx F backend CLI
├── internal/
│   ├── artifacts/                    # Artifact envelope I/O
│   ├── backlog/                      # Muti-Mind backlog parsing
│   ├── coaching/                     # Mx F coaching and retrospective data
│   ├── config/                       # Unified config loading
│   ├── dashboard/                    # Mx F dashboard rendering
│   ├── doctor/                       # Environment health checks
│   ├── gateway/                      # LLM reverse proxy (Vertex/Bedrock/Anthropic)
│   ├── impediment/                   # Impediment tracking and detection
│   ├── metrics/                      # Metrics collection and health analysis
│   ├── orchestration/                # Swarm orchestration engine
│   ├── sandbox/                      # Containerized sessions (Podman/DevPod)
│   ├── scaffold/                     # Core scaffold engine (embed.FS)
│   ├── schemas/                      # JSON Schema generation/validation
│   ├── setup/                        # Automated tool installation
│   ├── sprint/                       # Sprint lifecycle management
│   ├── sync/                         # GitHub issue sync
│   └── textutil/                     # Shared text formatting helpers
├── docs/                             # User-facing documentation
├── specs/                            # Architectural specs (001-035)
├── openspec/                         # OpenSpec tactical workflow
├── schemas/                          # JSON Schema registry
│   ├── feedback-triage/              # Feedback triage schemas
│   └── issue-triage/                 # Issue triage schemas
├── go.mod                            # Go module (1.25+)
├── opencode.json                     # MCP server configuration
├── .goreleaser.yaml                  # Release configuration
├── .packit.yaml                      # Fedora packaging (Packit)
├── unbound-force.spec                # Fedora RPM spec
└── .fmf/                            # Testing Farm metadata
```

All business logic lives under `internal/` and MUST NOT be
imported externally.

## Coding Conventions

- **Formatting**: `gofmt` and `goimports` (enforced by golangci-lint)
- **Naming**: PascalCase exported, camelCase unexported
- **Comments**: GoDoc-style on all exported functions and types
- **Error handling**: Return `error`, wrap with
  `fmt.Errorf("context: %w", err)`
- **Import grouping**: stdlib, third-party, internal (blank lines)
- **No global state**: Prefer functional style and DI
- **Logging**: `github.com/charmbracelet/log` (not stdlib `log`)
- **CLI**: `github.com/spf13/cobra` for commands and flags
- **Spec writing**: RFC 2119 language (MUST/SHOULD/MAY),
  Given/When/Then scenarios, FR-NNN numbering, line length < 72

## Testing Conventions

- **Framework**: Standard library `testing` only (no testify)
- **Assertions**: `t.Errorf` / `t.Fatalf` directly
- **Naming**: `TestXxx_Description`
- **Isolation**: `t.TempDir()` for filesystem tests
- **Drift detection**: Tests MUST verify embedded assets match
  canonical sources

## Technology Stack

- **Language**: Go 1.25+ (module: `github.com/unbound-force/unbound-force`)
- **CLI framework**: `github.com/spf13/cobra`
- **Logging**: `github.com/charmbracelet/log`
- **Terminal styling**: `github.com/charmbracelet/lipgloss`
- **YAML**: `gopkg.in/yaml.v3`, `github.com/goccy/go-yaml`
- **JSON Schema**: `github.com/invopop/jsonschema`,
  `github.com/santhosh-tekuri/jsonschema/v6`
- **Container runtime**: Podman (>= 4.3)
- **Workspace manager**: DevPod (>= 0.5.0, optional)
- **Embedding model**: `granite-embedding:30m` via Ollama

## Behavioral Rules

These rules are non-negotiable. Violations are CRITICAL severity.

- **Gatekeeping**: MUST NOT modify quality/governance gates
  (coverage thresholds, CRAP scores, severity definitions,
  CI flags, agent settings, constitution MUST rules, review
  limits, workflow markers). Stop and report instead.
- **Phase boundaries**: MUST NOT cross workflow phase boundaries.
  Spec phases: spec artifacts only. Implement: source code.
  Review: fixes only. Violation = process error, stop immediately.
- **CI parity**: MUST replicate CI checks locally before marking
  tasks complete. Derive commands from `.github/workflows/`.
- **Review council**: MUST run `/review-council` before PR
  submission. Resolve all REQUEST CHANGES. No code changes
  between APPROVE and PR. Exempt: constitution amendments,
  docs-only, emergency hotfixes.
- **Branch protection**: MUST NOT commit directly to `main`.
  All changes via feature branches and PRs.
- **Documentation gate**: Before marking a task complete,
  assess documentation impact: `CHANGELOG.md` for change
  entries, `AGENTS.md` for structural updates (project
  structure, conventions, build commands), `README.md` for
  description changes.
- **Documentation gate**: MUST file a documentation issue
  against the current repo for user-facing changes before
  PR merge. Exempt: internal refactoring, test-only,
  CI-only, spec artifacts.
- **Zero-waste**: No orphaned specs, unused standards, or
  aspirational documents that do not map to actionable work.
- **Commit scope**: Only commit files directly related to the
  active spec or change. Tooling scaffolds (`uf init`,
  convention pack updates, command directory renames, schema
  template updates) MUST be committed on a separate branch
  (e.g., `chore/uf-init-sync`), not mixed into feature
  branches. Never use `git add -A` or `git add .` on feature
  branches — stage files explicitly.

### PR Review Commands

| Command | When | Scope |
|---------|------|-------|
| `/review-council [N]` | Pre-PR + optional PR posting | 5+ Divisor agents |
| `/review-pr [N]` | Post-PR (GitHub) | Single agent, CI analysis |
| `/address-feedback [N]` | Post-PR (GitHub) | Triage + address reviewer feedback |

### Issue Triage Commands

| Command | When | Scope |
|---------|------|-------|
| `/triage-issue <N>` | Issue triage | 5 Divisor agents, classify + comment |

`/review-pr` key capabilities: PR walkthrough, issue
linking (`Fixes #N`), suggestion blocks, verdict-aligned
posting (APPROVE/REQUEST_CHANGES/COMMENT), path-based
focus heuristics, and review state awareness (fetches
existing reviews to prevent duplicate findings, warns
about stale review dismissal, detects CODEOWNER
requirements).

#### GitHub Review Lifecycle

- **Review states**: `APPROVED`, `CHANGES_REQUESTED`,
  `COMMENTED`, `DISMISSED`
- **Stale dismissal**: When `dismiss_stale_reviews` is
  enabled, APPROVE is auto-invalidated on new commits.
  `/review-pr` warns before posting APPROVE.
- **Review requests**: A user can appear in
  `requestedReviewers` even with a prior APPROVE (it was
  dismissed). `/review-pr` detects this.
- **CODEOWNER checks**: `/review-pr` warns when APPROVE
  may not satisfy `require_code_owner_reviews`.
- **Duplicate detection**: `/review-pr` warns before
  posting a second review from the same account.
- **Dependabot**: `ci_dependencies.yml` respects human
  `CHANGES_REQUESTED` before auto-approving.

## Specification Workflow

All non-trivial changes MUST be preceded by a spec workflow.

| Tier | Tool | When | Artifacts |
|------|------|------|-----------|
| Strategic | Speckit | >= 3 stories, cross-repo | `specs/NNN-*/` |
| Tactical | OpenSpec | < 3 stories, single-repo | `openspec/changes/*/` |

Pipeline: `constitution → specify → clarify → plan → tasks →
analyze → checklist → implement`

**Ordering**: Constitution before specs. Spec before plan. Plan
before tasks. Tasks before implementation. Spec artifacts MUST
be committed/pushed before implementation begins.

**Branches**: Speckit: `NNN-<name>`. OpenSpec: `opsx/<name>`.

**Task bookkeeping**: Mark checkboxes `[x]` immediately on
completion. `[P]` marks parallel-eligible tasks.

**When in doubt**: Start with OpenSpec. Escalate to Speckit if
scope grows beyond 3 stories or crosses repo boundaries.

**What requires a spec**: New features, refactoring that changes
signatures, test additions across multiple functions, agent
changes, CI changes, data model changes.

**Exempt**: Constitution amendments, typo fixes, emergency
hotfixes (retroactively documented).

## Knowledge Retrieval

Prefer Dewey MCP tools over grep/glob/read for cross-repo
context and architectural patterns.

| Intent | Tool |
|--------|------|
| Conceptual | `dewey_semantic_search` |
| Keyword | `dewey_search` |
| Navigation | `dewey_traverse`, `dewey_get_page` |
| Discovery | `dewey_find_connections`, `dewey_similar` |

**Fallback**: Use Read/Grep/Glob when Dewey is unavailable,
for exact string matching, known file paths, or non-Markdown
content (Go source, JSON, YAML).

## Architecture

Single binary CLI with layered internal packages:

- **Scaffold pattern**: `Options`/`Result` structs, `Run()`,
  file ownership (`isToolOwned`), version markers
- **Testable CLI**: Commands delegate to `runXxx(params)` with
  `io.Writer` injection for testing

### Embedding Model

Dewey and Replicator use `granite-embedding:30m` (Apache 2.0).
Override in `.uf/config.yaml` under `embedding.model`.

## Convention Packs

This repository uses convention packs scaffolded by
unbound-force. Agents MUST read the applicable pack(s)
before writing or reviewing code.

- `.opencode/uf/packs/default.md`
- `.opencode/uf/packs/default-custom.md`
- `.opencode/uf/packs/severity.md`
- `.opencode/uf/packs/content.md`
- `.opencode/uf/packs/content-custom.md`
- `.opencode/uf/packs/go.md`
- `.opencode/uf/packs/go-custom.md`
- `.opencode/uf/packs/python.md`
- `.opencode/uf/packs/python-custom.md`
- `.opencode/uf/packs/typescript.md`
- `.opencode/uf/packs/typescript-custom.md`
