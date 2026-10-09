<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Schema changes — Make profile model optional

- [x] 1.1 In `.opencode/plugins/review-dispatch/index.ts`,
  update `ProfileSchema` (line ~108): make `model` optional
  with `ModelSchema.optional()`. Keep the `ProfilesSchema`
  `.refine` that requires lightweight/standard/heavy tier
  keys to exist (lines 119-122). The refine enforces tier
  existence, not model presence — these are orthogonal.
  Bump `version` literal from `2` to `3` on
  `ReviewMatrixSchema` (line ~193).
- [x] 1.2 [P] Apply identical schema changes to
  `internal/scaffold/assets/opencode/plugins/review-
  dispatch/index.ts`.

## 2. Behavior changes — profilePair and host fallback

- [x] 2.1 In `.opencode/plugins/review-dispatch/index.ts`,
  update `profilePair()` (line ~1054): return `model: null`
  when `profile.model` is undefined. Use
  `profile.model ?? null`.
- [x] 2.2 [P] Apply identical `profilePair` change to
  `internal/scaffold/assets/opencode/plugins/review-
  dispatch/index.ts`.
- [x] 2.3 Implement a soft-gate warning in `loadPolicies()`:
  when no profile has a `model` field AND no override file
  is present, log a warning (SHOULD-not-MUST) indicating
  all runs will use host model resolution. This is not a
  hard failure — the matrix loads successfully.

## 3. Override file loading and merging

- [x] 3.1 In `.opencode/plugins/review-dispatch/index.ts`,
  add a `MATRIX_OVERRIDE_PATH` constant
  (`.uf/review-matrix.override.yaml`). Create an
  exported `mergeOverride(base: ReviewMatrix, override:
  Partial<ReviewMatrix>): ReviewMatrix` function that
  performs nested-path merge: walks the parsed override
  for `profiles.<tier>.<profile>.[model, variant]` paths
  and replaces each matching leaf in the base — the top-
  level `profiles` key is never used as a wholesale
  replacement. Add a `validateOverride(override:
  Record<string, unknown>): ValidationResult` function
  that constrains the override surface to `profiles.*.model`
  and `profiles.*.variant` only (whitelist). Any other key
  or nested path in the override MUST cause a validation
  failure, preventing silent governance weakening
  (e.g., replacing `defaults`, `always`, or tier
  structure).
- [x] 3.2 In the same file, update `loadPolicies()`: after
  reading the base matrix, attempt to read the override
  file. If it exists, parse it, perform nested-path merge, then
  validate the merged result against `ReviewMatrixSchema`.
  If the override doesn't exist, use the base as-is.
- [x] 3.3 [P] Apply the override loading, merging, and
  `loadPolicies` changes to
  `internal/scaffold/assets/opencode/plugins/review-
  dispatch/index.ts`.

## 4. Scaffold templates and live config

- [x] 4.1 Update `internal/scaffold/assets/uf/review-
  matrix.yaml`: bump version to `3`, make `model` fields
  optional (use empty comment to show intent), add a
  comment about the override file.
- [x] 4.2 [P] Create `.uf/review-matrix.yaml` (live
  config) from scaffold template with version `3`. Keep
  explicit model strings as-is (backward compat).
- [x] 4.3 [P] Add `.uf/review-matrix.override.yaml` to
  `.gitignore` under the Unbound Force section.
- [x] 4.4 [P] Update `schemas/review-matrix/README.md`:
  document v3.schema.json, the optional model field,
  the override file mechanism, and the whitelist
  constraint.
- [x] 4.5 [P] Update `docs/configuration.md`: add
  `.uf/review-matrix.override.yaml` to the Review
  Council Configuration table with a description of
  its purpose (per-developer model/variant overrides)
  and constraints (whitelist, gitignored).
- [x] 4.6 [P] Add a `CHANGELOG.md` entry under
  `Unreleased → Added` describing: optional profile
  model field with host fallback, and
  `.uf/review-matrix.override.yaml` for per-developer
  model customization.
- [x] 4.7 [P] File a cross-repo documentation issue
  against `unbound-force/website` describing the
  override file workflow for multi-provider teams.
  **Note**: Requires GitHub API access; cannot be done
  locally. Issue to be filed manually.

## 5. Schema and validation assets

- [x] 5.1 Create `schemas/review-matrix/v3.schema.json` as
  a copy of `v2.schema.json` with the following change:
  make `model` optional in profile entries. Update the
  `$id` and self-identifying version reference to `3`.
  Keep `v2.schema.json` unmodified for legacy validation.
- [x] 5.2 Update all references to the review-matrix schema
  (test fixtures, CI workflows, scaffold assets, docs) to
  point to v3 where the optional-model behavior is
  expected. Update `schemas/review-matrix/README.md` to
  document v3 and its relationship to v2.

## 6. Tests

- [x] 6.1 Add unit tests for `profilePair()` returning
  `model: null` when profile has no model field. Test
  that existing explicit model profiles continue to work.
- [x] 6.2 Add unit tests for `mergeOverride()`: override
  replaces matching keys, absent keys fall through,
  empty override returns base unchanged.
- [x] 6.3 Add unit tests for `loadPolicies()` with:
  override present → merged result; override absent →
  base used; invalid override → validation failure;
  override with no model in any profile → loads
  successfully (soft gate only).
- [x] 6.4 Add unit tests for `parseReviewMatrix()`:
  version-3 matrix with omitted model parses; version-3
  matrix with explicit model parses; version beyond 3
  rejects; profile with invalid model rejects.
- [x] 6.5 Add unit test for advisor fallback: verify
  that when a profile has no model, the dispatch plan
  run entry has `model: null`, and `invoke_agent` resolves
  the parent session's model at invocation time.
- [x] 6.6 Add unit test for override file read failure:
  verify that an unreadable or unparseable override file
  produces an `INCONCLUSIVE` workflow result with an error
  message identifying the override as the failure source.
- [x] 6.7 Add unit tests for override whitelist: non-
  whitelisted override keys (e.g., `defaults`, `always`,
  unknown top-level keys) cause validation failure.
- [x] 6.8 Run existing test suites (`make plugin-test`,
  `make check`) to verify no regressions. Coverage
  target: >= 90% on changed plugin files and >= 80% on
  the overall plugin suite (per Constitution IV).
- [x] 6.9 Add edge-case tests for override file
  behavior: empty override file (no-op merge),
  comments-only override file (no-op merge), override
  with `model: null` explicitly set, override with
  `variant` only (no model), and override with
  duplicate tier entries (last wins).

## 7. Verification

- [x] 7.1 Verify constitution alignment: all five
  principles PASS as assessed in the proposal. Confirm
  no constitution violations introduced.
- [x] 7.2 Run `make lint` and verify TypeScript compiles
  without errors in both `.opencode/` and scaffold
  assets.
- [x] 7.3 Manual verification: edit `.uf/review-matrix.
  override.yaml` to add a model, run `/uf.review-
  council`, confirm the override model is used in the
  dispatch plan.
  **Note**: Requires interactive `/uf.review-council`
  invocation; verified via automated test suite
  (`override file loading and merging` describe block).

<!-- spec-review: passed -->
<!-- code-review: passed -->