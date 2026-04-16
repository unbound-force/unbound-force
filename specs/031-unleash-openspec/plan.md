# Implementation Plan: Unleash OpenSpec Support

**Branch**: `031-unleash-openspec` | **Date**: 2026-04-16 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/031-unleash-openspec/spec.md`

## Summary

Extend `/unleash` to support `opsx/*` branches by
removing the hard stop, adding OpenSpec detection with
change name extraction, skipping Steps 1-3 (clarify/
plan/tasks), passing the correct feature directory and
workflow tier to the review council, and using a
unified `FEATURE_DIR` variable for Steps 4-8. The
change is ~50 lines in 1 Markdown command file
(`.opencode/command/unleash.md`) plus 1 scaffold asset
copy (`internal/scaffold/assets/opencode/command/
unleash.md`). No Go code changes.

## Technical Context

**Language/Version**: Markdown (OpenCode command files)
**Primary Dependencies**: OpenCode runtime, `/review-council` command, `/opsx-propose` artifacts
**Storage**: N/A (Markdown files deployed to target directory)
**Testing**: Manual verification + existing scaffold drift detection test (`TestEmbeddedAssets_MatchSource`)
**Target Platform**: Any platform running OpenCode
**Project Type**: CLI (meta-repo command file)
**Performance Goals**: N/A (command file, not runtime code)
**Constraints**: Backward compatible — `NNN-*` branches must behave identically to current behavior
**Scale/Scope**: 1 live file + 1 scaffold asset copy; ~50 lines changed

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Autonomous Collaboration — PASS

The change modifies a single command file that
orchestrates existing heroes (review council, Cobalt-
Crush) through well-defined artifacts. OpenSpec
artifacts (proposal.md, design.md, specs/, tasks.md)
are self-describing and consumed without synchronous
interaction. No new inter-hero coupling is introduced.

### II. Composability First — PASS

`/unleash` already works standalone. This change adds
a second mode (OpenSpec) that uses the same pipeline
steps (spec review, implement, code review,
retrospective, demo). The review council already
supports OpenSpec workflow tier detection (per
`review-council.md` line 61). No new mandatory
dependencies are introduced — if OpenSpec artifacts
don't exist, the command stops with an actionable
error.

### III. Observable Quality — PASS

The same quality gates apply to both modes: spec review
produces `<!-- spec-review: passed -->` marker, code
review produces `<!-- code-review: passed -->` marker.
Both markers are machine-parseable HTML comments. The
review council's structured output (APPROVE/REQUEST
CHANGES with findings) is unchanged.

### IV. Testability — PASS

This change modifies a Markdown command file, not Go
source code. Testability is verified through:
1. Scaffold drift detection test ensures the live file
   and scaffold asset copy remain synchronized.
2. Manual verification via the acceptance scenarios in
   the spec (run `/unleash` on `opsx/*` branch, verify
   pipeline executes).
3. Backward compatibility verified by running `/unleash`
   on a `NNN-*` branch (unchanged behavior).

No new Go code means no new coverage strategy is
needed. The existing `TestEmbeddedAssets_MatchSource`
test provides the automated regression gate.

**Coverage Strategy**: No new Go code — the scaffold
drift detection test (`TestEmbeddedAssets_MatchSource`)
is the automated regression gate. Behavioral
correctness is verified through manual acceptance
testing per `quickstart.md` (4 user stories, 10
acceptance scenarios). Automated behavioral testing
of OpenCode command files is not currently feasible —
this is a known limitation of the Markdown command
file architecture. The trade-off is acceptable because
the change is a conditional branch extension (~50
lines) in a single file with well-defined before/after
contracts.

## Project Structure

### Documentation (this feature)

```text
specs/031-unleash-openspec/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Phase 0: existing structure analysis
├── data-model.md        # Phase 1: change mapping
├── quickstart.md        # Phase 1: verification guide
├── contracts/           # Phase 1: before/after contract
│   └── unleash-command-contract.md
├── checklists/
│   └── requirements.md  # Pre-existing checklist
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
.opencode/command/
└── unleash.md                          # Live command file (MODIFIED)

internal/scaffold/assets/opencode/command/
└── unleash.md                          # Scaffold asset copy (SYNCED)
```

**Structure Decision**: No new files or directories are
created. Two existing files are modified: the live
`unleash.md` command file and its scaffold asset copy.
This follows the established pattern from Specs 018,
019, 022, and the `workflow-phase-boundaries` OpenSpec
change.

## Implementation Approach

### Change Strategy: Surgical Insertion

The change uses a surgical insertion strategy — adding
conditional branches to existing steps rather than
restructuring the command file. This minimizes diff
size and preserves the existing step numbering that
other documentation references.

### Key Design Decisions

1. **Branch detection before prerequisite check**:
   The `opsx/*` detection happens in Step 1 (Branch
   Safety Gate) before the `check-prerequisites.sh`
   call. This is because `check-prerequisites.sh`
   validates Speckit-specific prerequisites (spec.md
   in `specs/NNN-*/`), which don't apply to OpenSpec.
   For OpenSpec, we check `openspec/changes/<name>/
   tasks.md` directly.

   *Alternative rejected*: Modifying
   `check-prerequisites.sh` to support OpenSpec. This
   would require Go code changes and is unnecessary —
   the branch name itself provides sufficient routing
   information.

2. **Unified FEATURE_DIR variable**: Steps 4-8 already
   use a feature directory concept. The change
   introduces a `FEATURE_DIR` variable set early in
   Step 1 that works for both modes:
   - Speckit: `specs/NNN-feature-name/` (from
     `check-prerequisites.sh` JSON output)
   - OpenSpec: `openspec/changes/<name>/` (from branch
     name extraction)

   Steps 4-8 reference `FEATURE_DIR` instead of
   hardcoded `specs/` paths.

3. **Skip Steps 1-3 for OpenSpec**: Rather than adding
   OpenSpec-specific logic to the clarify, plan, and
   tasks steps, we skip them entirely with an
   announcement. This is correct because `/opsx-propose`
   creates all artifacts in one step — there is no
   separate clarify/plan/tasks phase in the OpenSpec
   workflow.

4. **Review council already supports OpenSpec**: The
   review council (`.opencode/command/review-council.md`)
   already detects `opsx/*` branches and adjusts its
   review scope to `openspec/changes/<name>/`. No
   changes to the review council are needed — we just
   need to pass the correct feature directory and
   remove the explicit "Spec Review Mode" override
   that assumes Speckit artifacts.

5. **Resumability markers are format-agnostic**: The
   `<!-- spec-review: passed -->` and
   `<!-- code-review: passed -->` markers are appended
   to `tasks.md` regardless of workflow tier. Both
   Speckit and OpenSpec `tasks.md` files support HTML
   comments. The resumability detection in Step 2 reads
   `tasks.md` from `FEATURE_DIR`, which works for both
   modes.

### Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Backward compatibility regression | Low | High | Speckit path is unchanged; only the `opsx/*` STOP is removed and replaced with detection logic |
| Review council receives wrong artifacts | Low | Medium | Review council already auto-detects workflow tier from branch name; no override needed |
| Scaffold asset drift | Low | Low | `TestEmbeddedAssets_MatchSource` catches drift automatically |
| OpenSpec tasks.md format incompatibility | Very Low | Medium | Both formats use `- [ ]`/`- [x]` checkboxes; the task loop is format-agnostic |

## Scaffold Asset Sync

Each modified `.opencode/` file has a corresponding
scaffold asset copy under `internal/scaffold/assets/
opencode/`. Both copies MUST be updated in sync. The
drift detection test (`TestEmbeddedAssets_MatchSource`)
enforces this.

```text
# Scaffold asset copies (synced from live files):
internal/scaffold/assets/opencode/command/unleash.md
```

## Documentation Impact

Upon implementation, the following documentation updates
are required:

- **AGENTS.md Recent Changes**: Add entry documenting
  OpenSpec support in `/unleash` (files modified, user
  stories completed, task count).
- **AGENTS.md Active Technologies**: No change — no new
  technologies introduced.
- **Website issue**: Assess whether a `docs:` issue is
  needed in `unbound-force/website` for the `/unleash`
  dual-mode workflow. A tutorial opportunity exists for
  engineers learning the new `opsx/*` branch support.

## Complexity Tracking

> No constitution violations to justify. All four
> principles pass cleanly.
