## Why

Pre-flight only runs tools it can discover from tracked files:
`.pre-commit-config.yaml`, `Makefile` targets, and
`.github/workflows` steps (pre-flight SKILL.md Phases 1-2). A
project has no way to add a deterministic check of its own
without editing one of those files or the tool-owned skill.
Both routes fail for real consumers:

1. Editing `.opencode/skills/pre-flight/SKILL.md` or
   `uf.review-council.md` is lost on the next `uf init`
   (both are tool-owned, `isToolOwned` in
   `internal/scaffold/scaffold.go`).
2. Adding a `Makefile` target or workflow step means changing
   tracked project files. A stealth install (`uf init
   --stealth`, git-excluded `.uf/` and `.opencode/`) cannot do
   that at all. Even where it can, `ci-aware` mode skips any
   tool it maps to a CI check, so a check that also runs in CI
   never gives local evidence before the council dispatches.
3. A new `divisor-*.md` agent is not a substitute. Divisors run
   with `bash: deny` and only see command output through
   `pre_flight_results`, so the evidence has to come from
   pre-flight.

The motivating case is the fullsend-ai org. Its `agents` repo
holds bash scripts whose `${VAR:?}` / `${VAR:-}` reads must be
supplied by the `fullsend` repo's workflow `env:`, setup script
and Go runner. Breakages of that contract (an env var nobody
supplies, an optional operator setting made required) recur, and
the only review today is LLM prose in a divisor. A project-owned
contract-checker script, run deterministically by pre-flight,
would have caught each one.

A first draft added the procedure as prose in a new pre-flight
Phase 2b (parse YAML, validate, probe PATH, filter by diff, run,
map gates, and for `soft-gate` build a merge-base worktree and
re-run failures). Reviewed against the determinism pack, that
draft violated DE-001 (procedure in prose), DE-003 (output
assembled by the model), DE-005 (a judgment-free chain spread
over steps) and DE-006 (routing in prose). It also had a defect:
a stealth `.uf/scripts/` tree does not exist in a merge-base
worktree, so every baseline re-run would fail and every failure
would classify as `unknown`, which blocks. This change records
the tool-based design that replaced it.

## What Changes

- New project-owned asset `.uf/preflight.yaml` (embedded at
  `internal/scaffold/assets/uf/preflight.yaml`): `version: 1`
  plus a list of declared checks (`name`, `run`, `args`,
  `requires`, `paths`, `gate: required|advisory`,
  `timeout_seconds`). It is seeded as a commented, empty
  template, survives plain `uf init` re-runs, and is reset only
  by `--force`.
- New shared library `.opencode/lib/preflight-declared.ts` that
  implements the whole procedure with injected side effects:
  - zod validation; `run` must be one repo-relative executable
    path, never a shell line;
  - PATH and executable probes;
  - diff-scope filtering with the Phase 2a fallbacks;
  - argv execution with no shell, a per-check timeout, and
    output tail truncation;
  - `required`/`advisory` status mapping;
  - in `soft-gate`, merge-base worktree baseline
    classification;
  - rendered Execution Results rows.
- New `uf_preflight_declared_checks` tool in the uf-workflow
  plugin. Its only argument is `mode`
  (`hard-gate|ci-aware|soft-gate`). It returns the standard
  envelope with structured results, a verdict and the rendered
  rows. Failure codes are `INVALID_PREFLIGHT_CONFIG` and
  `PREFLIGHT_DECLARED_FAILED`.
- Pre-flight SKILL.md gains Phase 2b, which is a single call to
  the tool plus how to fold its result into the report. It also
  gains one-line notes in Phase 2a (declared checks are scoped by
  the tool) and Phase 5 (insert `rendered` unchanged).
- Scaffold registration:
  - `uf/preflight.yaml` is project-owned in `isToolOwned`.
  - `opencode/lib/preflight-declared.ts` is tool-owned, part of
    the Divisor asset subset, and staged with the review plugin
    sources.
- Seam tests (DE-007): unit and mode tests for the library, a
  registration pin, two golden output-shape cases, and a
  round-trip test with a real filesystem and real process
  execution. A Go regression test covers re-run survival and the
  `--force` reset.
- CHANGELOG (Unreleased -> Added) and a `docs/configuration.md`
  row.

## Capabilities

### New Capabilities

- `pre-flight`: Project-declared checks. A project lists its own
  deterministic checks in `.uf/preflight.yaml`, and pre-flight
  runs them in every mode through one tool call, with the same
  file-scope and soft-gate semantics as detected tools.
- `scaffold-ownership`: `.uf/preflight.yaml` is a project-owned
  extension point.

### Modified Capabilities

- None. Detected-tool behavior (Phases 1-2a, 3-4b) is
  unchanged, and declared checks are kept out of CI mapping.

### Removed Capabilities

- None.

## Impact

- `.opencode/lib/preflight-declared.ts` (new, plus its embedded
  asset mirror).
- `.opencode/plugins/uf-workflow/index.ts` (plus mirror): new
  tool, `runProcessNoShell`, dependency factory and
  registration.
- `.opencode/skills/pre-flight/SKILL.md` (plus mirror): Phase 2b
  and two one-line notes. The file grows from 551 to 589 lines
  (+38); the prose draft was 652.
- `.uf/preflight.yaml` and `internal/scaffold/assets/uf/
  preflight.yaml` (new, byte-identical).
- `internal/scaffold/scaffold.go`: `isToolOwned` and
  `isDivisorAsset` entries.
- `internal/scaffold/plugin_activation.go`: the
  `stageReviewPluginSources` entry.
- `internal/scaffold/scaffold_test.go`: asset lists, ownership,
  mapping, Divisor-subset tables, and
  `TestRun_PreflightChecksPreservedOnRerun`.
- `cmd/unbound-force/main_test.go`: the deployed-file count
  range moves to 82-86 for the two new assets.
- `.opencode/test/`: `preflight-declared.test.ts` and
  `preflight-declared-roundtrip.test.ts` (new); the golden,
  uf-workflow and scratch-smoke suites gain the new tool.
- Generated repositories get one new seeded file. Pre-flight
  behavior changes only when that file declares checks; an
  absent or empty file reports "No project-declared checks".

Out of scope:
- Porting Phases 1-5 of the pre-flight skill to tools.
- Mapping declared checks to CI checks.
- Any change to divisor permissions or to how
  `pre_flight_results` reaches divisors.

## Constitution Alignment

Assessed against the Unbound Force org constitution
(`.specify/memory/constitution.md`).

### I. Autonomous Collaboration

**Assessment**: PASS

Projects and the review council meet through a file artifact,
`.uf/preflight.yaml`, and its results reach Divisors through the
existing `pre_flight_results` artifact. There is no new runtime
coupling. Each check is an independent executable with an
exit-code contract.

### II. Composability First

**Assessment**: PASS

This adds a documented extension point, so projects stop needing
to modify the skill or tracked build files. An absent or empty
file changes nothing, and uf has no new mandatory dependency.

### III. Observable Quality

**Assessment**: PASS

The tool returns the standard JSON envelope with per-check
status, exit code, timeout flag, unavailable binaries, scope
fallback reason, causality and baseline method, plus rendered
rows. A golden test pins the output shape.

### IV. Testability

**Assessment**: PASS

Every side effect (filesystem, PATH, git, process execution,
temp dirs) is injected. Unit and mode tests run with fakes.
The round-trip test uses only a temp dir and local executables,
with no external services.

### V. Security by Default

**Assessment**: PASS

`run` is validated as one repo-relative path with no `..`
segments and no shell metacharacters, and it executes through
`execFile` with `shell: false`. Arguments pass literally (the
round-trip test pins `$HOME` unexpanded and `a b` as one arg).
Each check has a capped timeout. An invalid config fails
pre-flight closed instead of being skipped. Declared checks only
add evidence; no existing gate is weakened.

### VI. Deterministic Execution

**Assessment**: PASS

The determinism-pack review is the reason this change took its
final shape:
- DE-001 and DE-005: the procedure is code.
- DE-003: the skill inserts tool-rendered rows.
- DE-006: gate and causality routing lives in the tool.
- DE-007: seam tests are listed above.
- DE-004: the line delta is +38.
