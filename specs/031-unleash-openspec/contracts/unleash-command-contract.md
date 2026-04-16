# Contract: `/unleash` Command — Branch Routing

**Branch**: `031-unleash-openspec` | **Date**: 2026-04-16

## Before (Current Behavior)

```
Branch          → Outcome
─────────────────────────────────────────────
main            → STOP: "Cannot run on main"
opsx/*          → STOP: "Use /opsx:apply"
NNN-*           → Run Speckit pipeline (8 steps)
other           → STOP: "Unrecognized branch"
```

## After (New Behavior)

```
Branch          → Outcome
─────────────────────────────────────────────
main            → STOP: "Cannot run on main" (unchanged)
opsx/<name>     → Run OpenSpec pipeline:
                    FEATURE_DIR = openspec/changes/<name>/
                    Gate: tasks.md must exist
                    Skip: clarify, plan, tasks
                    Run: spec review → implement →
                         code review → retrospective → demo
NNN-*           → Run Speckit pipeline (unchanged)
other           → STOP: "Unrecognized branch" (unchanged)
```

## OpenSpec Pipeline Flow

```
Step 0: Startup Cleanup (unchanged)
Step 1: Branch Safety Gate
  ├── Detect opsx/* branch
  ├── Extract <name> from opsx/<name>
  ├── Set FEATURE_DIR = openspec/changes/<name>/
  ├── Verify FEATURE_DIR/tasks.md exists
  └── Announce "Detected OpenSpec change: <name>"
Step 2: Resumability Detection
  ├── Clarify: always done (skip)
  ├── Plan: always done (skip)
  ├── Tasks: always done (skip)
  ├── Spec review: check marker in tasks.md
  ├── Implementation: check all [x] in tasks.md
  └── Code review: check marker in tasks.md
Step 3: Skip announcement
  └── "OpenSpec mode — artifacts from /opsx-propose,
       skipping clarify/plan/tasks"
Step 4: Spec Review (using FEATURE_DIR)
Step 5: Implement (using FEATURE_DIR/tasks.md)
Step 6: Code Review (using FEATURE_DIR)
Step 7: Retrospective (unchanged)
Step 8: Demo (using FEATURE_DIR; proposal.md not spec.md)
```

## Speckit Pipeline Flow (Unchanged)

> Note: Steps 0-2 are infrastructure (cleanup, gate,
> resumability). Steps 3-10 correspond to the 8
> user-facing pipeline steps (clarify through demo).

```
Step 0: Startup Cleanup
Step 1: Branch Safety Gate (NNN-* detection)
Step 2: Resumability Detection (all 6 checks)
Step 3: Clarify (Dewey-powered)
Step 4: Plan (delegate to cobalt-crush-dev)
Step 5: Tasks (delegate to cobalt-crush-dev)
Step 6: Spec Review (using FEATURE_DIR)
Step 7: Implement (using FEATURE_DIR/tasks.md)
Step 8: Code Review (using FEATURE_DIR)
Step 9: Retrospective
Step 10: Demo
```

## Resumability Markers

Both workflows use identical markers in `tasks.md`:

| Marker | Written By | Read By |
|--------|-----------|---------|
| `<!-- spec-review: passed -->` | Step 4 (spec review) | Step 2 (resumability) |
| `<!-- code-review: passed -->` | Step 6 (code review) | Step 2 (resumability) |

## Error Messages

| Condition | Message |
|-----------|---------|
| `opsx/*` branch, no tasks.md | "No tasks.md found for change `<name>`. Run `/opsx-propose` first." |
| `opsx/*` branch, change archived | Same as above (archived changes have no `openspec/changes/<name>/tasks.md`) |
| `main` branch | Unchanged |
| Non-matching branch | Unchanged |

## Backward Compatibility Guarantee

The following behaviors are guaranteed unchanged:

1. `NNN-*` branches execute the full 8-step Speckit
   pipeline with no behavioral difference.
2. `main` branch produces the same STOP error.
3. Non-matching branches produce the same STOP error.
4. All existing guardrails remain in effect.
5. All existing exit points (clarify, spec review,
   worker failure, merge conflict, build checkpoint,
   code review iterations) remain unchanged.
6. The scaffold asset copy mechanism is unchanged —
   only the file content is updated.
