# Research: Unleash OpenSpec Support

**Branch**: `031-unleash-openspec` | **Date**: 2026-04-16

## Existing `/unleash` Structure

The current `/unleash` command file
(`.opencode/command/unleash.md`, 603 lines) implements
an 8-step autonomous Speckit pipeline:

| Step | Section | Lines | Purpose |
|------|---------|-------|---------|
| 0 | Startup Cleanup | 31-44 | Stale worktree removal |
| 1 | Branch Safety Gate | 46-82 | Branch validation + spec.md check |
| 2 | Resumability Detection | 84-124 | Probe filesystem for completed steps |
| 3 | Step 1 -- Clarify | 126-195 | Dewey-powered clarification |
| 4 | Step 2 -- Plan | 197-213 | Delegate to cobalt-crush-dev |
| 5 | Step 3 -- Tasks | 215-228 | Delegate to cobalt-crush-dev |
| 6 | Step 4 -- Spec Review | 230-279 | Review council in spec review mode |
| 7 | Step 5 -- Implement | 281-424 | Task execution with parallel workers |
| 8 | Step 6 -- Code Review | 426-487 | Review council in code review mode |
| 9 | Step 7 -- Retrospective | 489-523 | Dewey learning storage |
| 10 | Step 8 -- Demo | 525-577 | Structured demo output |

### Branch Safety Gate (Step 1, lines 46-82)

The current gate has three branches:

1. `main` → STOP (unchanged)
2. `opsx/*` → STOP with "Use `/opsx:apply`" (to be
   replaced with OpenSpec detection)
3. Non-`NNN-*` → STOP (unchanged)

After the branch check, `check-prerequisites.sh
--json --require-spec` validates that `spec.md` exists
and extracts `FEATURE_DIR`. This script is
Speckit-specific — it looks for `specs/NNN-*/spec.md`.

### Resumability Detection (Step 2, lines 84-124)

Checks 6 conditions in order:
1. Clarify done? (no `[NEEDS CLARIFICATION]` markers)
2. Plan done? (plan.md exists)
3. Tasks done? (tasks.md exists)
4. Spec review done? (`<!-- spec-review: passed -->`)
5. Implementation done? (all `[x]`, no `[ ]`)
6. Code review done? (`<!-- code-review: passed -->`)

For OpenSpec, conditions 1-3 are always "done" (skip),
and conditions 4-6 work identically.

### Steps 4-8 (lines 230-577)

These steps reference the feature directory implicitly
through the `FEATURE_DIR` established in Step 1. They
read `tasks.md`, `spec.md`, and `quickstart.md` from
this directory. The review council auto-detects the
workflow tier from the branch name.

## Review Council OpenSpec Support

The review council (`.opencode/command/review-council.md`)
already supports OpenSpec:

- **Auto-detection** (line 61): Branch `opsx/*` →
  OpenSpec (tactical) workflow tier
- **Spec Review Mode** (lines 227-235): For OpenSpec,
  reviews `openspec/changes/<name>/` including
  proposal.md, design.md, specs/, tasks.md, plus
  referenced main specs at `openspec/specs/`
- **Announcement** (line 81-83): "Detected Spec Review
  Mode (OpenSpec)"

No changes to the review council are needed.

## OpenSpec Change Directory Structure

From examining `openspec/changes/finale-command/`:

```text
openspec/changes/<name>/
├── .openspec.yaml       # OpenSpec metadata
├── proposal.md          # Change proposal
├── design.md            # Design document
├── specs/               # Delta specs
│   └── *.md
└── tasks.md             # Task list (checkbox format)
```

The `tasks.md` file uses the same `- [ ]`/`- [x]`
checkbox format as Speckit tasks.md. The task execution
loop in Step 5 is format-agnostic.

## Scaffold Asset Synchronization Pattern

From Dewey learnings and prior specs (019, 022, 026):

- Each `.opencode/` file has a corresponding copy under
  `internal/scaffold/assets/opencode/`
- Both copies MUST be updated in sync
- `TestEmbeddedAssets_MatchSource` enforces drift
  detection
- No changes to `expectedAssetPaths` are needed (the
  file already exists in both locations)

## Branch Name Extraction

The branch name `opsx/<name>` maps directly to the
change directory `openspec/changes/<name>/`. Extraction
is a simple string operation:

```
Branch: opsx/finale-command
         ↓ strip "opsx/" prefix
Name:   finale-command
         ↓ construct path
Dir:    openspec/changes/finale-command/
```

## Guardrails Analysis

The current guardrails section (lines 579-603) includes:

> "NEVER run on `main`" — unchanged
> "NEVER skip spec review exit on HIGH/CRITICAL" — unchanged

The guardrail "NEVER run on `main` -- the command is
for Speckit feature branches only" needs updating to
reflect that `/unleash` now supports both Speckit
(`NNN-*`) and OpenSpec (`opsx/*`) branches.
