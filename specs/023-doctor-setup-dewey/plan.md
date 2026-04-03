# Implementation Plan: Doctor & Setup Dewey Alignment

**Branch**: `023-doctor-setup-dewey` | **Date**: 2026-04-03 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/023-doctor-setup-dewey/spec.md`

## Summary

Add a Dewey embedding capability health check to
`uf doctor` that verifies end-to-end embedding generation
via the Ollama HTTP API, demote the direct Ollama serving
check to informational status (Dewey manages Ollama's
lifecycle), and change `uf setup` to install the Swarm
plugin from the organization's fork
(`unbound-force/swarm-tools`) instead of upstream
(`joelhooks/swarm-tools`). Three user stories, two Go
packages modified, no new dependencies beyond the Go
standard library.

## Technical Context

**Language/Version**: Go 1.24+  
**Primary Dependencies**: `github.com/spf13/cobra` (CLI), `github.com/charmbracelet/lipgloss` (terminal styling), `gopkg.in/yaml.v3` (frontmatter parsing), `net/http` (standard library — new import in `checks.go`)  
**Storage**: N/A (reads environment state, writes no persistent data)  
**Testing**: Standard library `testing` package, `t.TempDir()` for isolation, injected mocks via `Options` struct  
**Target Platform**: macOS (darwin/arm64, darwin/amd64), Linux (amd64, arm64)  
**Project Type**: CLI tool (Go binary)  
**Performance Goals**: Embedding check completes in <5 seconds (timeout). Doctor full run <15 seconds.  
**Constraints**: No new external Go dependencies. All external calls injectable for testability.  
**Scale/Scope**: 2 packages, ~6 files modified, ~150 lines of new code, ~50 lines of modified code

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Pre-Research Check (Gate 1)

| Principle | Status | Evidence |
|-----------|--------|----------|
| I. Autonomous Collaboration | PASS | Doctor and setup operate independently. The embedding check reads Ollama state via HTTP — no runtime coupling to Dewey. Results are self-describing `CheckResult` structs with full metadata. |
| II. Composability First | PASS | The embedding check is skipped when Dewey is not installed (FR-003). Ollama check demotion uses informational messaging, not removal. No new mandatory dependencies introduced. |
| III. Observable Quality | PASS | All new checks produce `CheckResult` structs that serialize to JSON via `--format=json` (FR-007). Severity values use the existing `Pass`/`Warn`/`Fail` enum with stable serialization. |
| IV. Testability | PASS | The `EmbedCheck` function is injected via `Options` struct, following the established pattern (`LookPath`, `ExecCmd`, `ReadFile`). All new code is testable in isolation without network access or external services. Coverage strategy defined in research.md R6. |

**Gate 1 Result**: PASS — all four principles satisfied. Proceeding to Phase 0 research.

### Post-Design Check (Gate 2)

| Principle | Status | Evidence |
|-----------|--------|----------|
| I. Autonomous Collaboration | PASS | No inter-hero runtime coupling added. The embedding check calls the Ollama HTTP API directly — it does not depend on Dewey being running. |
| II. Composability First | PASS | The `EmbedCheck` field on `Options` defaults to a production implementation but can be overridden. When Dewey is absent, the check is skipped. When Ollama is absent, the check reports `Warn` (not `Fail`). |
| III. Observable Quality | PASS | The `CheckResult` struct is unchanged. JSON output includes the new check with the same schema. No format breaking changes. |
| IV. Testability | PASS | `defaultEmbedCheck` is a pure function that returns a closure. Tests inject mock `EmbedCheck` functions. No `net/http` calls in test code. Coverage strategy: unit tests for all new functions, update existing tests for changed install commands. |

**Gate 2 Result**: PASS — design satisfies all four principles.

## Project Structure

### Documentation (this feature)

```text
specs/023-doctor-setup-dewey/
├── plan.md                      # This file
├── research.md                  # Phase 0: research findings
├── data-model.md                # Phase 1: type changes
├── quickstart.md                # Phase 1: verification guide
├── checklists/
│   └── requirements.md          # Pre-existing checklist
├── contracts/
│   ├── doctor-checks.md         # Embedding check contract
│   └── setup-swarm.md           # Swarm fork install contract
└── tasks.md                     # Phase 2 output (created by /speckit.tasks)
```

### Source Code (repository root)

```text
internal/
├── doctor/
│   ├── doctor.go                # Options struct (add EmbedCheck field + default)
│   ├── checks.go                # checkDewey() modifications, new checkEmbeddingCapability()
│   ├── environ.go               # Update swarm install hints to fork source
│   ├── doctor_test.go           # New embedding capability tests, updated assertions
│   ├── models.go                # Unchanged
│   └── format.go                # Unchanged
└── setup/
    ├── setup.go                 # installSwarmPlugin() source change
    └── setup_test.go            # Updated swarm plugin install assertions
```

**Structure Decision**: No new packages or directories.
All changes are modifications to existing files in the
two established packages (`internal/doctor/` and
`internal/setup/`). This follows the existing project
structure and avoids unnecessary complexity.

## Coverage Strategy

### Unit Test Coverage

| Component | Test Approach | Target |
|-----------|--------------|--------|
| `checkEmbeddingCapability()` | Inject mock `EmbedCheck` returning nil/error | 100% branch coverage |
| `defaultEmbedCheck()` | Not unit-tested directly (makes HTTP call). Error categorization verified via manual quickstart.md steps. Production behavior validated by `checkEmbeddingCapability` mock tests (T009-T012) which exercise the contract surface. | N/A (injected in production, verified manually) |
| `checkDewey()` modifications | Extend existing `TestCheckDewey_*` tests | 100% of new branches |
| `installSwarmPlugin()` changes | Update existing mock assertions | 100% of changed paths |
| Install hint updates | Update existing `TestInstallHint_*` tests | 100% of changed strings |

### Coverage Ratchet

All existing tests must continue to pass (FR-008).
New tests must cover all new branches. No coverage
regression permitted per Constitution Principle IV.

## Complexity Tracking

No constitution violations to justify. All changes
follow established patterns and use existing types.

| Aspect | Complexity | Justification |
|--------|-----------|---------------|
| New `Options` field | Low | Follows existing injection pattern |
| HTTP call in `defaultEmbedCheck` | Low | Standard library only, 5-second timeout |
| Swarm install source change | Trivial | String replacement |
| Ollama demotion | Trivial | Message text change |

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Ollama API format changes | Low | Medium | The `/api/embed` endpoint is stable and documented. Version-pin the expected response shape. |
| Fork package name differs from upstream | Low | High | Verify `github:unbound-force/swarm-tools` installs correctly before merging. |
| Existing test breakage from install hint changes | Medium | Low | Search-and-replace all `opencode-swarm-plugin@latest` references in test files. |
| `net/http` import adds weight to doctor package | Low | Low | Standard library — no binary size impact. |
<!-- scaffolded by uf vdev -->
