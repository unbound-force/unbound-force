# Implementation Plan: Documentation Curation

**Branch**: `026-documentation-curation` | **Date**: 2026-04-11 | **Spec**: `specs/026-documentation-curation/spec.md`
**Input**: Feature specification from `/specs/026-documentation-curation/spec.md`

## Summary

Add a Documentation Curation capability to the Divisor
review council through two changes: (1) a new Curator
agent (`divisor-curator.md`) that detects documentation
gaps, files GitHub issues in the website repo for docs/
blog/tutorial opportunities, and blocks PRs with
unaddressed content needs; (2) a Guard enhancement
adding a "Documentation Completeness" checklist item
to the existing Code Review audit. No Go code changes
— this is Markdown agent definitions, scaffold asset
copies, and test count updates.

## Technical Context

**Language/Version**: Go 1.24+ (scaffold engine, tests only — no new Go logic)
**Primary Dependencies**: Markdown (agent files), `embed.FS` (scaffold engine)
**Storage**: N/A (Markdown files deployed to target directory)
**Testing**: Standard library `testing` package (`go test -race -count=1 ./...`)
**Target Platform**: macOS/Linux (CLI scaffold)
**Project Type**: CLI / meta-repository
**Performance Goals**: N/A (no runtime code)
**Constraints**: N/A
**Scale/Scope**: 1 new agent file (~250 lines), ~10 lines added to Guard, 1 scaffold asset copy, test count update

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Autonomous Collaboration — PASS

The Curator agent operates autonomously through the
review council's existing dynamic discovery mechanism.
It produces findings in the same structured format as
all other Divisor agents. Its outputs (findings, filed
GitHub issues) are self-describing artifacts with
provenance (agent name, severity, file references).
The Curator does not require synchronous interaction
with other agents — it reads the PR diff and produces
findings independently.

### II. Composability First — PASS

The Curator is independently deployable. Adding
`divisor-curator.md` to `.opencode/agents/` is
sufficient — no changes to `/review-council` are needed
because the council uses dynamic `divisor-*.md`
discovery. The Curator delivers value alone (filing
website issues) and adds value when combined with the
Guard (complementary documentation checks at different
scopes). Removing the Curator does not break any other
agent.

### III. Observable Quality — PASS

The Curator produces structured findings in the standard
Divisor output format (`[SEVERITY] Finding Title` with
File, Constraint, Description, Recommendation fields).
GitHub issues filed by the Curator are observable,
traceable artifacts with labels (`docs`, `blog`,
`tutorial`). All quality claims are backed by the
review council's existing fix/re-run verification loop.

### IV. Testability — PASS

The Curator is a Markdown agent file — no runtime code
to unit test. Testability is verified through:
- Scaffold asset drift detection (existing
  `TestAssetPaths_MatchExpected` test)
- The `expectedAssetPaths` count update ensures the new
  file is tracked
- The Curator's behavior is testable via the review
  council's existing integration flow (submit a PR,
  run `/review-council`, verify findings)
- No external services required for scaffold tests

**Coverage strategy**: Unit tests only (scaffold asset
count and drift detection). No new Go logic means no
new unit test functions beyond updating
`expectedAssetPaths`. The existing test suite validates
that all embedded assets are accounted for and that
scaffold output matches embedded content.

## Project Structure

### Documentation (this feature)

```text
specs/026-documentation-curation/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── divisor-curator-contract.md
└── quickstart.md        # Phase 1 output
```

### Source Code (repository root)

```text
.opencode/agents/
├── divisor-curator.md              # NEW: The Curator agent (user-owned)
└── divisor-guard.md                # MODIFIED: +Documentation Completeness checklist item

internal/scaffold/
├── assets/opencode/agents/
│   ├── divisor-curator.md          # NEW: scaffold asset copy
│   └── divisor-guard.md            # MODIFIED: scaffold asset copy synced
├── scaffold.go                     # NO CHANGES (isDivisorAsset already matches divisor-*)
└── scaffold_test.go                # MODIFIED: expectedAssetPaths +1 entry
```

**Structure Decision**: No new directories. The Curator
follows the established `divisor-*.md` pattern in the
existing `.opencode/agents/` directory. The scaffold
engine's `isDivisorAsset()` function already matches
any file with the `opencode/agents/divisor-` prefix,
so no Go logic changes are needed for `--divisor` mode.

## Complexity Tracking

No constitution violations to justify. All four
principles pass cleanly.

## Design Decisions

### DD-001: Curator as Divisor Agent (not standalone)

**Decision**: Implement the Curator as a `divisor-*.md`
agent within the existing review council, not as a
standalone agent or a separate command.

**Rationale**: The review council's dynamic discovery
(`divisor-*.md` glob) means zero changes to the
`/review-council` command. The Curator's findings
integrate into the existing fix/re-run loop. The
severity pack applies automatically. This is the
lowest-cost, highest-integration approach.

**Alternatives rejected**:
- Standalone command: Would require a new slash command,
  separate invocation, and manual integration with the
  review workflow. Higher cost, lower integration.
- Guard extension only: The Guard's scope is intent
  drift and governance. Adding GitHub issue filing and
  content opportunity detection would violate the
  Guard's single-responsibility boundary.

### DD-002: bash: true Exception for Curator

**Decision**: The Curator has `bash: true` in its
frontmatter — an exception to the standard Divisor
pattern where all other personas have `bash: false`.

**Rationale**: The Curator must execute `gh issue create`
and `gh issue list` to file and check for existing
issues in the `unbound-force/website` repo. These are
the only bash operations permitted. The agent file
documents this restriction explicitly in a "Bash Access
Restriction" section. This follows the principle of
least privilege — bash is enabled but scoped.

**Risk mitigation**: The agent file includes a clear
"Bash Access Restriction" section stating that only
`gh issue create` and `gh issue list` against
`unbound-force/website` are permitted. The Adversary
agent's "Gate Tampering" check would flag any attempt
to expand this scope.

### DD-003: Temperature 0.2 (not 0.1)

**Decision**: The Curator uses temperature 0.2, slightly
higher than the standard Divisor review temperature of
0.1.

**Rationale**: The Curator must make judgment calls about
whether a change is "significant enough" for a blog
post or "workflow-changing enough" for a tutorial. These
are inherently subjective assessments that benefit from
slightly more creative latitude than strict code review.
Temperature 0.2 is still very low — it provides minimal
variation while allowing the agent to exercise judgment
on content significance.

### DD-004: Guard Enhancement Scope

**Decision**: Add a single "Documentation Completeness"
checklist item to the Guard's Code Review section, not
a full new section.

**Rationale**: The Guard already checks for intent drift
and cross-component value preservation. Documentation
completeness is a natural extension of "cross-component
value" — if documentation doesn't match the code, value
is lost. A single checklist item (~10 lines) is
proportional to the check's scope. The Curator handles
the heavier cross-repo issue filing work.

### DD-005: No review-council.md Changes

**Decision**: Do not modify the `/review-council`
command file.

**Rationale**: The review council discovers agents via
`divisor-*.md` glob pattern. Adding `divisor-curator.md`
automatically includes the Curator in every review.
The Known Divisor Persona Roles reference table in
`review-council.md` SHOULD be updated to include the
Curator for documentation purposes, but this is
informational — the discovery mechanism does not depend
on the table.

**Correction**: After analysis, the reference table
SHOULD be updated to include the Curator's focus areas
for both Code Review and Spec Review modes. This ensures
the council provides targeted context when delegating
to the Curator. This is a ~3-line addition to the
existing table.
