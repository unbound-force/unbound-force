# Changelog

All notable changes to this project are documented in this file.
Each entry follows the format: `- <change-name>: <summary>`.

## Unreleased

### Added
- model-matrix-delegation: Capability-driven model selection for
  `uf_delegate_command` -- delegated content-agent work (the tool
  behind `/uf.unleash` and `/uf.run-pipeline-plan`) now runs on
  the model best suited to it instead of always replaying the
  parent session's model. New configuration pair mirroring the
  review matrix: tracked `.uf/model-matrix.yaml`
  (capability-named profiles, a `packs:` mapping from UF
  convention packs to capabilities, task-kind `defaults`,
  optional `agents:` pins, and `classifier:` structure -- model
  strings optional everywhere, host fallback) plus gitignored
  `.uf/model-matrix.override.yaml` carrying per-developer model
  identities through a fail-closed whitelist
  (`profiles.*.model`, `profiles.*.variant`, `classifier.llm.*`
  only; anything else fails validation naming the whitelist).
  `uf_delegate_command` gains `model`/`variant`, `profile`, and
  `pack` parameters resolved through a fixed precedence chain --
  explicit model, explicit profile, explicit pack (validated:
  pack file exists AND has a matrix entry), agents pin, static
  defaults by task kind, classifier cascade, host model -- and a
  resolved matrix model skips host introspection entirely
  (delegation no longer fails on `HOST_MODEL_UNAVAILABLE`;
  zero-config behavior with no matrix present is unchanged). For
  `implement` and unknown task kinds a two-engine classifier
  cascade derives the pack: a hidden `model-matrix-classifier`
  subagent (investigate-then-decide, JSON-only response, bounded
  by `agent_timeout_ms`, default 45 s) then an LLM engine
  (chat-completions POST, `temperature: 0`, JSON response format
  with one 400-retry without it, AbortSignal request bound
  default 10 s, apiKey interpolated from `{env:VAR}` with
  literal-key rejection, endpoint/model/apiKey configured only
  in the gitignored override), confidence-threshold gated
  (default 0.6), matrix-validity checked, first valid
  above-threshold result wins, results cached in-process by
  SHA-256 of feature dir + prompt + agent (bounded 128-entry
  per-instance cache), and every failure mode -- timeout,
  malformed or fences-wrapped JSON, oversized (above 50 MiB)
  response body, low confidence, unconfigured engines -- falls
  back to the static default; classifier failure never blocks
  delegation. The `pack` parameter does double duty: it selects
  the model AND renders a tool-generated pack directive (fixed
  template at the fixed prompt position after the skill
  directive, golden-shape tested -- never agent-reconstructed)
  instructing the child to load `.opencode/uf/packs/<pack>.md`
  before writing. Every delegation records its selection
  provenance -- `selection_source` (one of `explicit-model`,
  `explicit-profile`, `explicit-pack`, `agent-pin`,
  `default-kind`, `classifier-agent`, `classifier-llm`, `host`),
  the selected pack, resolved capability and model, and
  classifier confidence -- in the success envelope and checkpoint
  data. New meta-prompt convention pack
  `.opencode/uf/packs/meta-prompt.md` (MP-NNN rule ids, #888)
  codifies spec/plan/tasks/command authoring conventions and
  anchors the starter `packs:` mapping. The selection layer
  lives in a standalone `.opencode/lib/model-matrix.ts` (closed
  Zod schema, loader with injectable `readText`/`fileExists`
  seams, override merge, precedence resolution, classifier
  engines) mirrored byte-identical into
  `internal/scaffold/assets/opencode/lib/` alongside the synced
  uf-workflow plugin mirror; a JSON Schema registry lands at
  `schemas/model-matrix/` (v1.schema.json, README, sample
  fixtures cross-validated against the Zod surface with pinned
  per-family verdicts) with byte-identical
  `.uf/schemas/model-matrix/` and scaffold mirrors; the embedded
  template `internal/scaffold/assets/uf/model-matrix.yaml`
  deploys the starter matrix; `.uf/model-matrix.override.yaml`
  is appended to the root and scaffold-managed gitignore blocks;
  and `.opencode/commands/uf.unleash.md` (canonical plus mirror)
  recommends passing `pack` on implement delegation when the task
  clearly names one. Constitution alignment verified: standalone
  injectable lib module (II, IV); `selection_source` provenance
  on every delegation (I, III); credentials only in the
  gitignored override with `{env:VAR}` interpolation and a
  fail-closed whitelist (V); deterministic precedence with the
  classifier bounded, cached, and overridable and the pack
  directive tool-rendered (VI); every behavior testable without
  external services (IV). Documentation gates filed as in-repo
  issue #893 and website issue unbound-force/website#358.
  Plugin suite: 50 files / 1645 tests passing at 95.76 percent
  statements under the unchanged governance-controlled 94
  percent floor (model-matrix lib at 99.76 percent). (Spec:
  openspec/changes/model-matrix-delegation/, Fixes: #887,
  Fixes: #888, Fixes: #893)
- review-dispatch-diff-manifest: Large changesets now review via a
  diff manifest plus on-demand retrieval instead of inlining the full
  diff into every reviewer prompt. `build_review_prompts` selects the
  delivery mode from the diff's line count using the single shared
  `DIFF_THRESHOLD` (1500 lines) and counting function (`.opencode/lib/
  review-diff-manifest.ts`): inline delivery at or below the
  threshold — byte-for-byte identical to the previous assembly,
  including the 2 MiB inline diff cap — and manifest delivery strictly
  above it. In manifest mode no prompt inlines the diff; each carries a
  `## Diff Manifest` section (rendered entry table with paths, +/-
  counts, flags, chunk ids, and persona-relevance hints inside
  untrusted-content delimiters) plus retrieval instructions. The new
  `read_diff_chunk` tool serves chunks on demand, confined to
  manifest-declared byte coordinates from the path-validated diff,
  bounded to 256 KiB per response in untrusted-evidence delimiters,
  and fails closed on any validation or manifest-to-diff SHA-256
  integrity mismatch. The manifest-mode response contract adds a
  `**Files Reviewed**:` attestation field (one repo-relative path per
  line); `consolidate_dispatch` emits an advisory
  `empty_coverage_attestation` warning for successful runs whose
  attestation is missing or empty (quality signal only — no finding,
  no verdict change). Manifest-mode runs gain a fixed additive
  retrieval timeout allowance (default 300 s, recorded in timeout
  provenance) that `redispatch_failed_runs` also grants on the
  recovery path. Per the build-review-prompts FR-013 delta, an
  over-2 MiB over-threshold diff now proceeds through manifest
  delivery instead of failing `prompt_size_exceeded` (the 4 MiB
  prompt cap still applies to every assembled prompt in both modes);
  `fetch_pr_review_context`'s threshold warning now states that
  manifest delivery is engaged, preserving the pinned "Diff exceeds
  size advisory threshold" substring. Triage mode (no diff) is
  unaffected. Scaffold mirrors synced byte-identical
  (`review-diff-manifest.ts` registered across all four asset-surface
  enumerations; 79 embedded assets) with drift tests green; delivery-
  mode behavior documented in `uf.review-council.md`,
  `uf.unleash.md`, `uf.review-pr.md`, and `uf.address-feedback.md`
  (canonical plus mirrors). Plugin suite: 44 files / 1462 tests
  passing under the unchanged 94 percent statements coverage floor.
- sibling-evidence-agent-guidance: Pointed agents at declared,
  in-repo, permission-safe sibling-repo resources after a
  deployed /uf.unleash pipeline wedged 26 minutes on an
  out-of-workspace sibling read (parented child sessions
  cannot answer permission asks). Added MUST rule FA-005 to
  the content convention pack (consult `.uf/sibling-repos.yaml`,
  use `acquire_sibling_evidence` or the in-repo
  `.uf/sibling-clones/<name>/` / `.uf/sibling-cache/<name>/`
  copies, never read sibling repos from outside-workspace
  paths); added "Sibling repository references" research steps
  to `speckit.specify`, `speckit.plan`, `opsx-propose`, and
  `uf.run-pipeline-plan`; added a "Sibling Repositories"
  section to AGENTS.md plus a new `uf init` behavior
  (`ensureAGENTSmdSiblingSection`) that appends the same
  section to deployed repos' AGENTS.md when
  `.uf/sibling-repos.yaml` exists (stealth-gated,
  pattern-matched to the Convention Packs append, with nine
  new scaffold tests); and threaded `child_session_id` into
  `uf_delegate_command` DELEGATE_FAILED / DELEGATE_CANCELLED
  envelopes (declared on `ToolFailure`, omitted when no child
  session existed) so failed delegations name their child
  session for direct forensics. (Docs-issue exemption:
  jflowers recorded the human decision at the review-council
  curation gate on 2026-10-30 to defer the in-repo docs,
  website, and tutorial issues rather than file them.)
- extend-fetch-pr-review-context: Extended the review-dispatch
  plugin tool `fetch_pr_review_context` with five PR-repair
  output fields — `state`, `mergeable`, `merge_state_status`,
  `head_repository`, `base_repository` (the last derived from
  the PR `url`; `gh pr view --json` rejects `baseRepository`
  as an unknown field, verified 2026-10-28 and re-verified
  2026-10-29 against PR #868) — and made `pr_number` optional,
  resolving the current branch's open PR when omitted
  (`pr_not_found` when none exists); every existing caller
  passes `pr_number` explicitly and is unaffected;
  `.strict()` schema validation retained. Added two tools on
  the same intent-level path: `fetch_pr_checks` (`gh pr checks
  <N> --json name,state,bucket,link` with exit-code-tolerant
  parsing — stdout JSON is check-status data on any exit code;
  a zero-checks PR reports passing eligibility with an empty
  array, the verified gh shape being empty stdout plus the
  "no checks reported" stderr marker at exit 1 — and a derived
  `eligibility` classification `failing`/`pending`/`passing`,
  unknown buckets fail-closed to `pending` plus a warning) and
  `fetch_failed_run_logs` (`gh run view <run-id> [--repo
  owner/repo] --log-failed`, persisted to
  `$TMPDIR/opencode/run-<run_id>-failed-logs.txt` with `0o600`
  permissions and `O_EXCL` semantics under a `0o700` parent,
  returning `log_path` plus byte/line counts — never the log
  text inline; `run_not_found` for missing/expired runs;
  zero-line fetch succeeds with a warning). Re-routed
  `/uf.fix-ci` (canonical `.opencode/commands/uf.fix-ci.md`
  plus a byte-identical scaffold mirror) Phase 1.0/1.1/1.2 and
  Phase 2.1 fetch mechanics through the three tools with
  observable behavior preserved (STOP messages, gates,
  eligibility outcomes, never-merge rule unchanged); the
  uf-fix-ci change's raw-gh deviation records (design D1
  inventory entries, spec FR-001/FR-002/FR-003/FR-010) retired
  with closed-with-reference notes; and
  `TestCommandContracts_FixCiCommandStructure` re-pinned from
  raw-gh fetch assertions to tool-routed ones (the `gh run
  rerun --failed` pin kept — rerun stays prose). Measured
  instruction-text delta (design D8): across tasks 4.1-4.2 the
  re-route replaced 106 lines of raw-gh fetch prose with 100
  lines of tool-call instructions in `uf.fix-ci.md` (net -6;
  the command file measured 1262 → 1256 lines), collapsing
  the Phase 1.0 availability blocks, the Phase 1.1
  auto-detect invocation, the Phase 1.2 bash blocks with
  the non-zero-exit caveat and bucket-categorization
  table, and the Phase 2.1 `mktemp` + fetch block.
  Plugin suite extended: new suites `fetch-pr-checks.test.ts`,
  `fetch-failed-run-logs.test.ts`, and the shared
  `test/helpers/gh-spawn-mock.ts` Bun.spawn mock harness;
  review-dispatch tool lists pinned in `scratch-smoke.test.ts`
  and `plugin-integration.test.ts`; 1263 tests passing under
  the unchanged 94 percent statements / 85 percent branches
  coverage floors (no threshold edits per the gatekeeping
  rule). Originating issue #861.
  (Spec: openspec/changes/extend-fetch-pr-review-context/)
  (Docs: #873, unbound-force/website#342, unbound-force/website#310)
- uf-fix-ci: Added a new post-PR command `/uf.fix-ci
  [PR_NUMBER]` (canonical `.opencode/commands/uf.fix-ci.md`
  plus a byte-identical scaffold mirror under
  `internal/scaffold/assets/opencode/commands/`): diagnoses
  each failing CI check on an open PR from fetched log
  evidence (`gh run view --log-failed`), classifies the root
  cause as exactly one of code failure, stale branch,
  conflict, or flaky/infra, then carries the repair through.
  When the branch is conflicted or behind, the head branch
  is rebased onto the base branch -- rebase only, never
  merge; no option ever merges base into head, and pushes
  after a rebase use force-with-lease. Code fixes run
  through a gated loop: a MANDATORY GATE question-tool
  approval of the diagnosis and fix plan before any fix
  executes, conventional `fix:` commits scoped to the
  failing checks, and a `/uf.review-council` gate before any
  push, with a single durable iteration counter capped at 3
  (exhaustion never pushes). Flaky/infra failures offer
  `gh run rerun --failed` instead of a code change. Final
  verification polls CI until green (timeout-bounded) and
  reports the outcome. Guardrails: never merge, approve, or
  close the PR; never touch protected base branches; no fix
  without the human gate; no push without the council gate;
  untrusted log text never interpolated into shell
  arguments. PR-state/checks/log fetches remain command-prose
  deviations bound to companion issue #861
  (fetch_pr_review_context extension plus intent-level
  checks/log fetches -- not implemented here), with the
  supporting mechanical choreography enumerated in the
  change's design D1 deviation inventory. Complements
  `/uf.rebase-pr` (#851) in the post-PR
  command set; originating issue #858. The structural pin
  test `TestCommandContracts_FixCiCommandStructure` asserts
  the never-merge wording, gate markers, resume guard, gate
  wiring, terminal PR state, tool references, and the D1
  choreography pins on both command copies.
  (Spec: openspec/changes/uf-fix-ci/)
  (Docs: #862, unbound-force/website#342)
- port-uf-rebase-pr: Ported the verified personal global command
  `resolve-branch-conflict.md` into the UF command set as
  `.opencode/commands/uf.rebase-pr.md`, invoked as
  `/uf.rebase-pr <PR_NUMBER>`: rebases a conflicted PR's head
  branch onto its base branch, resolves conflicts, and
  force-pushes with lease so the PR can be merged with rebase
  once approved. The port follows `uf.review-pr.md` conventions
  (YAML frontmatter, `<protect>` block, session-resume guard,
  `^[1-9][0-9]*$` argument pre-validation) across six phases —
  prerequisites, PR fetch/gating, checkout and rebase, conflict
  resolution, push, verify — under nine guardrails: no
  merge/approve/close, never push the base branch, remotes
  resolved never assumed, `--force-with-lease` never `--force`,
  conflicts-only edits, no guessing, clean backout, gh
  verification before API calls, and determinism. Prefers the
  existing git plugin tools (`git_fetch`, `git_rebase`,
  `git_rebase_abort`, `git_push_force_with_lease`) per
  determinism pack DE-001, with fork-aware remote resolution
  that verifies the branch upstream is
  `<HEAD_REMOTE>/<HEAD_BRANCH>` before using the plugin push,
  falling back to `git push --force-with-lease <remote>
  <branch>`. The scaffold mirror
  `internal/scaffold/assets/opencode/commands/uf.rebase-pr.md`
  is byte-identical and registered in `expectedAssetPaths`
  (commands count 11 → 12) so scaffold drift detection stays
  green; `TestRunInit_FreshDir`'s count range shifts
  [70-74] → [71-75] for the 76th embedded asset.
  (Spec: openspec/changes/port-uf-rebase-pr/, Fixes: #851,
  Docs: #856, unbound-force/website#340)
- replicator-version-floor: Added a version-floor sub-check to
  `uf doctor`'s Replicator check group (Spec 024, FR-011): when
  the `replicator` binary is present, the sub-check runs
  `replicator --version` under a 5s timeout, parses the
  semantic-version major.minor.patch core (development and
  build suffixes ignored, so a dev build past the floor
  satisfies it), and emits a `replicator version` CheckResult.
  At or above v0.7.0 it Passes with the detected version;
  below the floor it Warns -- never Fails, since an outdated
  install is not treated more harshly than a missing one and
  Replicator is optional under Constitution Principle II --
  with a message like `v0.6.0 (requires >= v0.7.0)` and the
  upgrade hint `brew upgrade unbound-force/tap/replicator`
  (install URL: https://github.com/unbound-force/replicator).
  Version-command errors, timeouts, and unparseable output
  all fail open to a Warn "version could not be verified"
  while the remaining group sub-checks still run; an absent
  binary preserves the existing early return with no version
  result. The floor lives in a single named constant in the
  doctor package, and the new row serializes additively into
  `uf doctor --format json` (no breaking changes). Doctor
  tests cover the above-floor, below-floor, dev-suffix,
  unparseable, and version-command-failure paths.
  (Spec: openspec/changes/replicator-version-floor/)
  (Docs: #844, unbound-force/website#337)
- derive-per-run-timeout: Reviewer runs now derive their timeout
  windows from the work being dispatched instead of a fixed 600s
  default (issue #849). A pure derivation module
  (`.opencode/lib/run-timeout.ts`) computes each run's window from
  the review-matrix limit -- `limits.per_run_timeout_seconds` with
  optional `per_run_timeout_tiers` overrides (RD-FR-009; the three
  schema copies are synced) -- plus the dispatched work: prompt
  bytes above a 64 KiB free window, the plan entry's tier, and a
  contention multiplier when `max_parallel_runs` is greater than
  one, clamped to the configurable ceiling and the schema's
  1_800_000 ms bound. `dispatch_review_runs` and
  `redispatch_failed_runs` apply the derivation when
  `per_run_timeout_ms`/`timeout` is omitted (RD-FR-029); explicit
  values still win, and every run file records its timeout
  provenance (effective window, derivation inputs, source).
  Dispatch and redispatch invocations now persist operation-window
  records (`operation-*.json`: started/completed timestamps,
  status, run-count buckets) that `dispatch_status` reports as an
  `operations` array with `total_operations` and
  interrupted-operation detection (RD-FR-031). The DR-FR-002
  recovery path now re-runs cancelled reviewers -- a cancelled run
  is an infrastructure failure (parent tool-call window death), not
  a judgment call to discard -- and restores stranded superseded
  archives (a killed recovery invocation that archived a prior run
  but never persisted its replacement) to their active run-file
  paths before target resolution, most recent per plan entry,
  reporting `restored_runs`/`restored_agents` in the redispatch
  summary; prior cancelled records stay archived with their
  harvested submissions for audit, and consolidation excludes
  submissions from runs that fail or are cancelled again under
  RD-FR-025. Command prose in `/uf.review-council`,
  `/uf.review-pr`, `/uf.triage-issue`, `/uf.address-feedback`, and
  `/uf.unleash` now recommends omitting `per_run_timeout_ms` so
  derivation applies. DE-007 seam verification: the
  `per_run_timeout_ms` schema maximum aligns with the executor's
  TimeoutSchema 1_800_000 bound, the tool descriptions document
  the derived windows, operation reporting, and recovery
  semantics, and the new behaviors round-trip through real
  invocations (operation records written by dispatch/redispatch
  are read back through `dispatch_status`; run-file provenance is
  read through consolidation). Constitution alignment verified:
  the derivation is a pure standalone function (II, IV);
  provenance and window records are machine-parseable artifacts
  with no runtime coupling (I, III); input validation is tightened,
  not weakened (V); timeout selection is deterministic with no LLM
  judgment (VI); every new behavior is testable in isolation
  without external services (IV). Classified user-facing through
  the scaffold-shipped surfaces (the `per_run_timeout_ms` input
  contract, the `dispatch_status` and run-record output fields,
  the RD-FR-009 matrix config, the command-prose guidance, and
  the DR-FR-002 recovery semantics); documentation gates filed as
  in-repo issue #859 and website issue
  unbound-force/website#341. Vitest suites pin the derivation
  matrix, provenance persistence, operation windows, and recovery
  semantics (258 dispatch-suite tests; plugin coverage 95.41%
  statements / 88.41% branches against the 94/85 floors); scaffold
  mirrors under `internal/scaffold/assets/opencode/` updated
  byte-identically (drift-detection tests).
  (Spec: openspec/changes/derive-per-run-timeout/)
- unleash-interactive-clarify: Added an interactive clarify ask
  gate to `/uf.unleash` Step 1 (Clarify, Speckit tier): each
  question Dewey cannot answer -- and every question when
  Dewey is unavailable -- is now put to the human through a
  single question tool call before any exit, instead of
  ejecting the developer into manual spec editing and a full
  pipeline re-run. One question per call, no batching (the
  next ask only after the previous response), each carrying
  the question text and its surrounding spec context; the
  ask SHOULD present 2-5 mutually exclusive options drawn
  from the question's own enumerated alternatives, MUST
  list the spec's stated working default first labeled
  "(Recommended)" when the spec states one, MUST keep a
  custom free-form answer enabled, and MUST include a defer
  option in every ask. Questions Dewey answered never reach
  the gate (silent auto-resolution preserved); there is no
  question cap -- every unanswerable question is resolved or
  deferred before Step 1 completes. A chosen or custom
  answer resolves through the existing `resolve_clarification`
  tool with a new optional `resolved_by` enum ("dewey" |
  "human", default "dewey"): `source` is now conditionally
  required -- required for "dewey", rejected for "human"
  (the ask gate itself is the answer's provenance), enforced
  fail-closed at the schema level -- and the `## Clarifications`
  log entry now records who resolved each question:
  "(Dewey-resolved from ...)" as today versus
  "(human-answered during /uf.unleash clarify)" for ask-gate
  answers. Deferred questions -- and all unanswerable
  questions when the question tool is unavailable in the
  session -- flow to the existing clarify exit unchanged,
  the fallback for unattended runs. After processing, Step 1
  announces the counts of questions auto-resolved by Dewey,
  human-answered at the ask gate, and deferred, proceeding
  to Step 2 (Plan) when the unanswerable list is empty; the
  Guardrails section names the ask gate as a second
  legitimate mid-pipeline human interaction alongside the
  Step 4 per-finding spec review decision gate. Vitest cases
  in `.opencode/test/unleash-clarify.test.ts` pin the
  `resolved_by` schema, the conditional `source` requirement,
  and both provenance renderings (existing Dewey-path
  assertions unchanged under the default); scaffold mirrors
  under `internal/scaffold/assets/opencode/` updated
  byte-identically (drift-detection tests).
  (Spec: openspec/changes/unleash-interactive-clarify/)
- rebase-upstream-before-push: Added a step 4 "Sync with Base
  Branch" to `/uf.finale`, between commit and push, so PRs
  are no longer born stale when the base branch has advanced.
  Backed by a new `uf_finale_base_state` uf-finale plugin
  tool (8 → 9 registered tools): it resolves the target
  remote (`upstream` when present, else `origin`), resolves
  the base branch (the remote's HEAD via `git symbolic-ref`,
  `main` fallback), fetches it, and returns `{remote,
  baseBranch, baseRef, aheadBy, behindBy}` in the standard
  success/failure envelope. When the branch is behind, a
  session-resume-guarded MANDATORY GATE offers rebase onto
  `<remote>/<baseBranch>` (recommended), continue without
  rebase (PR will be behind), or abort; a confirmed rebase
  that conflicts aborts via `git_rebase_abort`, reports the
  conflicted files, and stops with the pre-rebase state
  preserved. The push step selects its mode from the sync
  outcome: `git_push_force_with_lease` after a rebase on a
  branch with an upstream, plain `git_push` (with
  `setUpstream`) for a first push. A base-state check failure
  (fetch or count, e.g. offline) is a soft gate -- a retryable
  `BASE_SYNC_FETCH` or `BASE_SYNC_COUNT` warning, then the
  finale proceeds exactly as before. Steps
  after the insertion renumber (Push 4→5, PR 5→6, CI 6→7
  with 6b→7b, Return to Main 7→8, Summary 8→9); the step-2
  no-changes path now lands on the sync step; all in-file
  cross-references update. Scaffold mirrors under
  `internal/scaffold/assets/opencode/` updated byte-identically
  (enforced by drift-detection tests). Unit tests in
  `.opencode/test/uf-finale.test.ts` cover upstream remote
  preference, origin-only fallback, symbolic-ref failure
  falling back to `main`, the offline `BASE_SYNC_FETCH`
  envelope, the `BASE_SYNC_COUNT` envelope when commit
  counting fails, and ahead/behind counts.
  (Spec: openspec/changes/rebase-upstream-before-push/)
- delegation-skill-enforcement: Closed the enforcement gap
  behind audit finding F6 / issue #787 -- the "mechanical
  tasks use repo tooling" prose rule in `uf.unleash.md`
  graduated to a fail-closed tool contract (Principle VI
  applied to its own enforcement gap).
  `uf_delegate_command`
  (`.opencode/plugins/uf-workflow/index.ts`) gained three
  parameters: `task_classification` (`deterministic` |
  `judgment`; required for prompt-only delegation, optional
  for command-path), `skill` (a repo skill name resolved
  against `.opencode/skills/` then `.opencode/skill/` with
  first-root precedence on collision, or a
  traversal-guarded repo-relative script path --
  existence-verified before the command-file read,
  host-model resolution, and child spawn), and
  `skill_absent_reason` (the recorded justification).
  Deterministic-classified calls supplying neither field
  fail non-retryably with `SKILL_REFERENCE_REQUIRED`;
  unresolvable references fail with `SKILL_NOT_FOUND`
  naming the searched locations; both envelopes carry
  remediation text. A resolved reference is injected into
  the composed prompt as a tool-rendered directive at a
  fixed position (after command text and feature-dir
  context, before the extra prompt), and the
  classification, resolved reference, and any recorded
  reason are written to the checkpoint data and success
  envelope -- the audit record, not prompt archaeology.
  The command template's "Mechanical tasks use repo
  tooling" prose block (with its other-project
  renumber-adr example) became a self-describing
  classification step plus the parameter contract on
  every prompt-only call shape -- quick-tier
  whole-change, per-task implement, and code-fix -- and
  Swarm parallel spawn prompts now carry the
  classification and verified reference in the subtask
  description. Governance layers: the determinism pack
  (repo-local) bumped 1.2.0 → 1.3.0 adding DE-008
  [MUST] (deterministic delegation carries a verified
  skill/script reference or a substantive recorded
  absence reason; delegation instruction uses the
  parameter contract, not prose-only skill
  instructions) and joined the MUST/CRITICAL severity
  mapping; the determinism-first skill gained matching
  delegation guidance and Related Enforcement entries;
  and divisor-guard's Code Review Mode checklist gained
  a Deterministic Delegation Discipline section
  (self-describing; cites no repo-local artifacts). All
  three scaffold-embedded sources (uf-workflow plugin,
  uf.unleash.md template, divisor-guard.md) updated
  with byte-identical mirrors, drift-detection tests
  green. Tests: `.opencode/test/unleash-delegate.test.ts`
  grew to 77 tests covering every spec scenario (skill
  resolution across both roots, first-root collision,
  script paths, traversal rejection, SKILL_* envelope
  remediation, golden-shape directive assertions,
  checkpoint round-trips, judgment pass-through,
  command-path exemption), with argument-contract pins
  added to `uf-workflow.test.ts` and
  `scratch-smoke.test.ts` (DE-007 layer 1). (Spec:
  openspec/changes/delegation-skill-enforcement/,
  Refs: #787)
- unleash-tier-routing: Added a tier-routing step to `/uf.unleash`
  for ON_MAIN/UNRECOGNIZED_BRANCH states: autonomous tier selection
  via the new tier-routing skill, a `uf_unleash_route_decision` tool
  that records the decision, mechanically enforces risk floors
  (rejecting `quick` when a floor signal is present), and creates
  `quick/<name>` branches; a `quick` workflow tier that runs
  implement, pre-flight checkpoint, review council, retrospective,
  and demo without spec artifacts; issue-reference resolution
  (`#NNN`/URLs via `gh issue view`, untrusted-content handling)
  threaded through delegation prompts and re-keyed checkpoint state
  so `/uf.finale` emits `Closes #<N>` for every tier; and
  `WORKFLOW_TIERS` gaining `quick` across gate, resume, templates,
  and finale.
- seam-verification: Added three test layers verifying the
  seams prose→tool migrations create, plus pack rule DE-007
  codifying the obligation. Layer 1 — registration
  surface: new suite `.opencode/test/registration-surface.test.ts`
  (20 tests) dynamically enumerates all 47 tools across
  the four plugins, asserting presence with description
  and execute, the pre-#732 `parameters:` mis-key guard
  (the load-bearing content of the biconditional
  schema-non-empty-iff-declared property — the SDK
  `tool()` helper is an identity pass-through over `args`
  alone), and execution of the `z.object(args)` conversion
  the host validates input through (it throws on non-zod
  shapes and rejects non-object payloads); the zod-shape
  fallback is covered structurally with a documented
  transition condition. Layer 2 — documented
  output shape: new suite `.opencode/test/golden-output-shape.test.ts`
  (41 tests over 43 documented-output-shape sites across
  uf-workflow/uf-finale/review-dispatch) executes each
  documented tool call in the suite's registry scope and
  asserts the documented fields are present in the
  returned envelope, with all 8 `format_unleash_exit`
  variants additionally pinned by tool==lib equivalence
  to the renderer goldens; the fixture inventory is
  grep-derived to prevent drift, and the four
  `dispatch_review_runs` sites are pinned in
  `dispatch-review-runs.test.ts` as golden tests across
  every documented input variant (triage — the #735
  contract — code-mode, and feedback-mode). Layer 3 — round-trips
  and command-contract pins: write→read round-trip
  assertions in `uf-finale.test.ts` pin all 8 writer
  parameters (including `review_verdicts` and
  `learnings_stored`); the dispatch-artifact resume-parser
  round-trip in `finalize-review-dispatch.test.ts`
  confirms every consumed field pinned; and Go-side
  command-contract pins expand `TestUnleashCommand_ToolRoutingContracts`
  (`internal/scaffold/scaffold_test.go`) from 10 → 32
  subtests — census-complete: every tool the `uf.unleash`
  command text invokes is pinned or excluded as a
  non-migration seam. Determinism pack
  `.opencode/uf/packs/determinism.md` bumped 1.1.0 →
  1.2.0, adding DE-007 [MUST] (seam verification: a
  prose→tool migration increment must verify the seams
  it creates in the same increment — registration
  surface, documented output shape, round-trips;
  tasks.md must include a seam-verification task naming
  the three layers; violations are CRITICAL), with the
  matching authoring-time hook in
  `.opencode/skills/determinism-first/SKILL.md`. Pack and
  skill remain repo-local; no scaffold-mirror changes.
  (Spec: openspec/changes/seam-verification/, Refs: #765)
- router-pattern-codification: Extended the determinism
  policy one level, from step-level extraction to
  sequence-level collapse (the Router Pattern: the agent
  outputs a single intent; deterministic tool code maps it
  to the required sequence of calls). Determinism pack
  `.opencode/uf/packs/determinism.md` bumped 1.0.0 → 1.1.0,
  adding a terminal "Intent-Level Orchestration" section
  with DE-005 [MUST] (a fixed sequence of tool calls with
  no judgment between steps — each next call and its
  arguments fully determined by the previous call's
  structured result — MUST collapse into a single
  intent-level tool; ordering, batching, retry, failure-stop,
  and branching on structured results belong to the tool)
  and DE-006 [SHOULD] (intent-level tools keep judgment at
  their boundary: judgment products arrive as validated
  parameters, routing decisions execute inside the tool as
  constants or pure functions). Severity mapping updated to
  (DE-001, DE-003, DE-005) / (DE-002, DE-004, DE-006);
  DE-001..DE-004 rule text byte-unchanged (append-only IDs).
  Determinism-first skill
  `.opencode/skills/determinism-first/SKILL.md` gained
  chain-collapse guidance: new "Chains Without Judgment
  Collapse Too" section plus five prose updates (frontmatter
  description second sentence, twice-test extension, two
  mechanical-procedure bullets, extraction step 4, "DE-001
  through DE-006"). Rules enter every Divisor review prompt
  via existing pack auto-discovery. Both files remain
  repo-local (no constitution, AGENTS.md, scaffold, or
  plugin changes).
  (Spec: openspec/changes/router-pattern-codification/, Refs: #758)
- gaze-ai-report: Added a push-to-main Gaze AI report job to Local CI
  (`ci_local.yml`) consuming the SHA-pinned unbound-force/gaze reusable
  workflow with thresholds 30/12/55; `build-and-test` now uploads
  `coverage.out` as the `coverage-out` artifact for the same-run handoff.
- unleash-deterministic-tools: Migrated `/uf.unleash`'s
  remaining deterministic procedures into 11 tested
  uf-workflow plugin tools (uf_unleash_branch_gate,
  uf_delegate_command, uf_unleash_detect_resume,
  scan_clarifications, resolve_clarification,
  format_unleash_exit, format_unleash_demo,
  write_unleash_summary, parse_tasks, mark_task_complete,
  scan_conflict_markers) plus shared
  lib/branch-classification.ts (consumed by the refactored
  uf_finale_startup; legacy output shape preserved) and
  lib/unleash-templates.ts. Step tools auto-record
  checkpoints (13 ceremony blockquotes removed; the two
  review-loop iteration checkpoints remain explicit);
  every exit and the demo render through formatter tools
  with verbatim emission (OUTPUT FIDELITY GUARD and
  re-read instructions deleted; guardrail amended).
  Resumability is artifact-authoritative:
  uf_unleash_detect_resume derives review completion from
  .uf/artifacts/dispatch/ verdicts (branch-scoped,
  per-mode freshness) over checkpoint state.
  write_unleash_summary persists the unleash summary
  (.uf/artifacts/unleash-summary/<branch>.json) in the
  exact shape /uf.finale reads, closing the
  finale-token-reduction contract
  (whatWasBuilt/howToVerify/testResults; implements #707).
  The legacy persist_unleash_summary name remains
  registered as a compatibility alias of
  write_unleash_summary, but it now carries the
  write_unleash_summary contract (change_name, tier,
  feature_dir, review_verdicts, learnings_stored, plus the
  three demo bodies) — callers of the sibling change's
  three-parameter contract (whatWasBuilt/howToVerify/
  testResults only) receive INVALID_INPUT; alias removal
  will be revisited once
  openspec/changes/persist-unleash-summary/ archives.
  Command markdown + Go/TS test updates only; scaffold
  mirrors byte-identical. The tool layer is fail-closed
  end to end: every tool validates input through shared
  glue (parseToolInput/resolveToolParam/failureFromError),
  rejects git pathspec magic in feature_dir, and converts
  unexpected runtime failures into *_ERROR envelopes,
  with each catch-all path carrying a regression test
  (scan/resolve/demo); delegation pins for Steps 2/3/5
  and staged lib-module lockstep (content, tool-ownership,
  Divisor staging) guard the scaffold side. Includes the
  maintainer-approved review-matrix per-run timeout bump
  600s → 1200s (mirrors synced). Increment 2 of the
  deterministic-instruction migration. (Spec:
  openspec/changes/unleash-deterministic-tools/, Refs:
  #716, #707)
- persist-unleash-summary: `/uf.unleash` Step 10 (Demo) now
  persists the presented demo output via a new
  `persist_unleash_summary` uf-workflow tool: the What Was
  Built, How to Verify, and Test Results section bodies
  (excluding `##` headers) are written as a flat JSON
  artifact at `.uf/artifacts/unleash-summary/<branch>.json`
  (validation-first, hardened write: 0o700 dir, 0o600 file,
  symlink refusal, detached-HEAD and git-error failures).
  `/uf.finale` reads the artifact in the PR body through
  `uf_finale_pr_context`, replacing grep-based working-tree
  recovery. Soft-fail in the pipeline: persist errors warn
  and continue. Plugin tool + tests (14), command update,
  scaffold mirrors.
  (Spec: openspec/changes/persist-unleash-summary/,
  Refs: #707)
- unleash-direct-review-dispatch: Rewrote `/uf.unleash`
  Steps 4 (Spec Review) and 6 (Code Review) to call
  review-dispatch tools directly (resolve_base_ref,
  plan_review_dispatch, dispatch_review_runs,
  consolidate_dispatch) instead of delegating to a
  cobalt-crush-dev driver session — one child session
  removed per review step and the 1,145-line council
  command payload no longer enters the dispatch path;
  council protocol sections are cited as normative
  reference instead. Added a parent-side spec auto-fix
  consent gate: on REQUEST CHANGES with only LOW/MEDIUM
  findings, the question tool MUST be presented (Apply
  auto-fixes / Review findings first / Skip auto-fixes)
  before any spec artifact is modified; consented fixes
  follow severity.md (HIGH/CRITICAL report-only), then
  every included reviewer re-runs (max 3 iterations).
  Step 6 now runs the pre-flight soft gate before
  dispatch (branch-caused failures stop with fail-closed
  finalization and zero child reviewers; pre-existing
  failures recorded non-blocking) and maps verdicts:
  approving verdicts proceed, REQUEST CHANGES enters the
  ≤3-iteration fix loop with cobalt-crush-dev applying
  code fixes, INCONCLUSIVE/UNAVAILABLE exits with
  reasons. The "NEVER pause between steps" guardrail now
  names the consent gate as the single legitimate
  mid-pipeline human interaction. Command markdown only;
  scaffold mirror byte-identical (791 → 1006 lines).
  Increment 1 of the deterministic-instruction migration.
  (Spec: openspec/changes/unleash-direct-review-dispatch/,
  Refs: #716)
- review-voice-pack: Added `review-voice` + `review-voice-custom`
  convention pack pair for per-project review-output prose style
  (sentence style, framing, tone). Always-deployed via scaffold
  registration (7→9); applied to GitHub-posted prose in
  `/uf.review-pr`, `/uf.review-council`, `/uf.triage-issue`,
  `/uf.address-feedback`. Style-only precedence: MUST NOT alter
  severities, verdicts, finding counts, footers, or structural format.
  Custom packs are exempt from embedded-asset drift comparison (FR-023).
  (Spec: openspec/changes/review-voice-pack/, Fixes: #657)
- determinism-codification: Repo-local codification of the
  Deterministic Execution principle across the governance stack.
  Constitution amended 1.3.1 → 1.4.0, adding Principle VI
  "Deterministic Execution" to `.specify/memory/constitution.md`.
  New governance convention pack `.opencode/uf/packs/determinism.md`
  with rules DE-001 [MUST], DE-002 [SHOULD], DE-003 [MUST],
  DE-004 [SHOULD], auto-discovered by review dispatch with no
  plugin changes. New skill `.opencode/skills/determinism-first/SKILL.md`
  with classify-before-writing guidance for coding agents. AGENTS.md
  gains a "Determinism gate" Behavioral Rule (CRITICAL severity)
  plus an exhaustive Convention Packs listing. Test-file
  registration: `knownNonEmbeddedFiles` entries in
  `internal/scaffold/scaffold_test.go` and a governance-pack
  skip-list entry in `internal/schemas/packvalidator_test.go`.
  Pack and skill are deliberately repo-local (no `uf init`
  propagation). Motivated by the token-reduction epic (#659) and
  the uf.unleash audit (#716).
  (Spec: openspec/changes/determinism-codification/)
- finale-token-reduction: Reduced `/uf.finale` command template from
  ~1,089 to ~300 lines (~70% token reduction). Extracted deterministic
  procedures into plugin tools: 10 shared git tools and
  `normalize_model_name` in uf-workflow plugin, 8 composite tools in
  new uf-finale plugin. Implements unleash-summary artifact consumption
  (#707). Uniform plugin probe validation across all 4 plugins.
  (Spec: openspec/changes/finale-token-reduction/, Closes: #659)
- uf-workflow-checkpoint: Extended uf-workflow plugin with three new tools
  (`checkpoint`, `get_progress`, `clear_progress`) for unified workflow
  progress tracking. Six command files updated to replace execution
  checklist checkbox editing with structured JSON state files in
  `.uf/progress/`. Includes symlink defense, path sanitization, and
  restrictive file permissions (0o600/0o700). (Spec:
  openspec/changes/uf-workflow-checkpoint/)
- build-from-source: Added "Build from Source" install
  instructions to README.md and QUICKSTART.md. Documents
  prerequisites (Go 1.25+, make), clone + install commands
  (`make install`), and verification (`uf version`).
  (Spec: openspec/changes/build-from-source/,
  Fixes: #536)

### Removed

- remove-mutimind-mxf-orchestration: Removed Muti-Mind product owner
  hero (cmd/mutimind, internal/backlog, internal/sync, schemas/hero-manifest,
  schemas/backlog-item, .opencode/agents/muti-mind-po.md, 12 muti-mind
  commands). Removed Mx F manager hero (cmd/mxf, internal/coaching,
  internal/dashboard, internal/impediment, internal/metrics, internal/sprint,
  schemas/acceptance-decision, schemas/metrics-snapshot, schemas/coaching-record,
  schemas/workflow-record, .opencode/agents/mx-f-coach.md, 5 workflow commands).
  Removed internal/orchestration swarm orchestration engine. Updated
  documentation, configuration, and constitution references. Constitution
  amended to v1.3.1 (PATCH: remove Muti-Mind and Mx F references).
  (Spec: openspec/changes/remove-mutimind-mxf-orchestration/, Fixes: #565, Fixes: #566)

### Changed
- reviewer-capabilities-init-survival: Reclassified
  `.uf/reviewer-capabilities.yaml` from tool-owned to
  project-owned in the scaffold's `isToolOwned` table,
  mirroring `.uf/review-matrix.yaml`: seeded once on first
  deploy, skipped on plain `uf init` re-runs, and reset only
  by `--force`, which remains a destructive reset for both
  review-config files. Custom Divisor agent registrations in
  a generated repository's `reviewers:` list now survive
  re-init byte-for-byte instead of being silently reverted
  to the embedded canonical list; canonical capability-list
  updates no longer auto-flow to generated repositories, so
  the seeded file stays as deployed until a project edits
  it, and a manifest missing a newly deployed agent fails
  closed at dispatch with the remediation message naming
  the missing agent and file. The ownership split is pinned
  by the flipped test-table entry and a re-run regression
  test; `docs/architecture.md` and `docs/configuration.md`
  state the symmetric seeding and survival semantics for
  both review-config files.
  (Spec: openspec/changes/reviewer-capabilities-init-survival/,
  Fixes: #884)
- question-tool-presentation-contract: Applied the question-tool
  presentation contract to all 22 human-facing decision gates in
  `/uf.triage-issue` (8), `/uf.address-feedback` (6), and
  `/uf.review-pr` (8): each gate prints its decision content as
  plain session text above the question tool (visibility
  directive — the same details and options), adds per-option
  reasoning (what each choice does and when it fits) and an
  advisory recommendation grounded in the command's own
  assessment data, and presents the recommended option first,
  labeled "(Recommended)". Option sets, MANDATORY GATE markers,
  session-resume guards, and CRITICAL RULEs are unchanged; the
  scaffold mirrors are synced byte-identical, and a new structural
  pin test (TestQuestionToolPresentationContractGates) enforces
  the shared contract fragments, the per-gate option inventories,
  and the per-command gate counts (8/6/8).
  (Spec: openspec/changes/question-tool-presentation-contract/)
  (Docs: #853, unbound-force/website#339)
- update-run-pipeline-plan: `/uf.run-pipeline-plan` now executes
  each work item through two `uf_delegate_command` delegations
  (`/uf.unleash`, then `/uf.finale`) instead of three `general`
  subagents running `/opsx-propose` -> `/uf.unleash` ->
  `/uf.finale` — the child `/uf.unleash` session resolves the
  issue, routes the tier, and runs the entry command itself.
  The orphaned Gate Relay Protocol (`GATE:<name>` structured
  values, `task_id` resumption, the "subagents MUST NOT invoke
  the question tool" rule — implemented by no command) is
  removed: human gates inside delegated commands surface
  as tool-rendered paused exits relayed verbatim
  (question-tool prompts inside a parented child do
  not surface; the delegation wedges at the gate until
  timeout, recovered by clearing the gate on the
  branch or an orchestrator re-anchor), and the
  orchestrator's `question` use is limited to Phase 0
  triage and pause-relay decisions. Paused exits are now
  modeled: any `/uf.unleash paused at:` heading (eleven known
  variants) is relayed verbatim, the todowrite entry is
  annotated `[BLOCKED: <variant>]`, and a pause-relay question
  offers re-run / skip / abort; hard delegation failures route
  through the same protocol. The disk handoff table now checks
  the feature branch before `/uf.finale` and the PR plus clean
  return to `main` after. The `/uf.finale` description
  refreshes to nine steps (including sync-with-base), the
  explicit-staging guardrail re-anchors to verification-only
  (staging happens inside the finale child via
  `uf_finale_stage`), the Closes linkage becomes a
  verification backstop (`gh pr view` / `gh pr edit`) threaded
  natively via `originating_issue`, and tracking requires
  `gh pr checks` green before `[x]` with a non-green-CI relay
  protocol. Scaffold mirror under
  `internal/scaffold/assets/opencode/commands/` synced
  byte-identically (enforced by drift-detection tests).
  (Spec: openspec/changes/update-run-pipeline-plan/)
  (Docs: #838, unbound-force/website#332)
- spawn-subtask-classification-params: The `/uf.unleash` Swarm
  parallel path's spawn call now carries the delegation
  classification contract as tool parameters. The implement
  step's parallel-execution spawn call block passes
  `task_classification` and, for deterministic tasks, the
  verified `skill` reference or `skill_absent_reason` to
  `replicator_forge_spawn_subtask` in a fenced call block
  mirroring the `uf_delegate_command` call blocks, forwards
  any returned `skill_directive` verbatim, and names the
  fail-closed behavior the tool enforces before the subtask
  spawns (input validation on a missing classification or on
  skill fields in a judgment-classified call,
  `SKILL_REFERENCE_REQUIRED`, `SKILL_NOT_FOUND`). The
  description-carrier prose and the "no classification
  parameters" sentence are deleted: both were false since
  unbound-force/replicator#143 shipped the parameter
  contract on the spawn tool, and the prose-only spawn calls
  failed the deployed tool's required `task_classification`
  input validation, breaking the Swarm parallel path; the
  parameterized template stays forward-compatible with older
  replicator binaries (Go's `json.Unmarshal` ignores unknown
  fields). Scaffold mirror
  `internal/scaffold/assets/opencode/commands/uf.unleash.md`
  synced byte-identically. Governance wording corrected to
  the same carrier: DE-008's Swarm-path clause in
  `.opencode/uf/packs/determinism.md` now binds Swarm-path
  `[P]` delegations to pass the classification and reference
  as spawn-tool parameters (pack bumped 1.3.0 -> 1.4.0; rule
  strengthened, MUST/CRITICAL mapping unchanged), the Guard
  reviewer's Deterministic Delegation Discipline checklist
  (`.opencode/agents/divisor-guard.md` + scaffold mirror)
  audits spawn call parameters instead of subtask prompt
  prose, and the determinism-first skill's Swarm-path
  guidance and Related Enforcement entries name both
  delegation tools' contracts. Pin tests in
  `internal/scaffold/scaffold_test.go` assert the parameters
  are present and the stale sentence absent in both the
  canonical command and its mirror. (Spec:
  openspec/changes/spawn-subtask-classification-params/,
  Fixes: #830, Refs: unbound-force/replicator#143)
- unleash-build-test-fix-loop: /uf.unleash no longer halts for a
  human when a blocking gate reports a failing check. The quick-tier
  pre-flight gate, the per-phase checkpoint, and Step 6's pre-flight
  soft gate (branch-caused or unknown failures) now classify every
  failure by causality at gate entry, then run a build/test
  fix loop for change-caused or unknown-origin failures: up to
  3 iterations (a durable budget recorded in the gate's checkpoint
  data, spanning sessions across resumes) of delegating the
  diagnosis and repair to cobalt-crush-dev via prompt-only
  uf_delegate_command (task_classification declared per call), then
  deterministically re-running the exact pre-flight commands in the
  calling gate's mode and re-classifying remaining failures.
  A phase checkpoint whose loop succeeds re-records its verdict as
  green (superseding the failed record); the quick gate records
  quick-pre-flight; Step 6 proceeds to its dispatch. Pre-existing
  and environmental failures never enter the loop; the gate treats
  them as green and records them in the checkpoint inventory. Each
  gate still exits through its unchanged format_unleash_exit
  variant once change-caused checks persist past the budget's
  3 total iterations, and the guardrails exit list now reads
  "build/test failures that persist after the fix loop's
  3 iterations". Scaffold pins cover the loop in both the command
  and its mirrored scaffold asset. (Spec:
  openspec/changes/unleash-build-test-fix-loop/, Docs: #819,
  unbound-force/website#322)
- spec-review-per-finding-gate: Replaced the bulk spec auto-fix
  consent gate and the HIGH/CRITICAL dead-stop in spec review
  with a per-finding decision gate. Each REQUEST CHANGES finding
  is presented one at a time in severity order (CRITICAL, HIGH,
  MEDIUM, LOW) as session text with per-option reasoning and a
  recommendation, followed by a question tool offering "Apply
  suggested fix" (any severity -- a human-directed consented
  edit), "Override & continue" (justification recorded in
  checkpoint data and the final report), "Show more context",
  "Pause & exit" (routes through the tool-rendered spec-review
  exit), and "Apply all remaining LOW/MEDIUM fixes" (bulk
  shortcut, LOW/MEDIUM only). The pipeline cannot proceed past a
  HIGH/CRITICAL finding without an explicit per-finding human
  choice; the autonomous auto-fix boundary stays LOW/MEDIUM-only
  (severity.md gains a clarifying note); the 3-iteration
  re-review cap, session-resume guard, and override recording
  are retained. The spec-review exit reason string now reads
  "Unresolved spec review findings (pause or iteration cap)".
  Applies to `/uf.unleash` Step 4 and `/uf.review-council` Spec
  Review Mode; scaffold mirrors and structural tests updated in
  lockstep.
  (Spec: openspec/changes/spec-review-per-finding-gate/, Fixes: #777)
- ts-plugin-coverage-gate: Raised the statements coverage
  floor for the `.opencode/` TypeScript plugin suite from 90
  to 94 percent in `.opencode/vitest.config.ts`; enforced by
  the existing `plugin-test` / `make check` / Local CI
  `plugin-tests` chain; resolves the coverage-gate deferral
  tracked by #780. (Spec:
  openspec/changes/ts-plugin-coverage-gate/, Closes: #780)
- unleash-pipeline-hardening: Hardened the `/uf.unleash`
  pipeline against the failure modes surfaced by the
  fullsend field audit. Delegation: `resolveHostModel` is
  variant-optional and throws the new
  `StructuralHostModelError`; `uf_delegate_command`
  classifies host-resolution failures as structural
  (non-retryable, with remediation) or transient
  (retryable) instead of failing opaquely. Review diffs:
  new shared core `.opencode/lib/review-diff.ts` plus a
  `capture_review_diff` tool (review-dispatch plugin)
  replaces prose diff-capture in the template's spec and
  code review steps — bounded, fail-closed, with
  uncommitted-local and divergence fallbacks — verified
  per DE-007 (registration surface, golden output shape,
  real-repo round-trip). Tier routing: route decisions
  validate tier entry-command existence
  (`ROUTE_TIER_TOOLING`, `TIER_ENTRY_COMMANDS`) and stop
  through the new `tier-routing-dirty-tree` and
  `tier-routing-missing-tooling` exit variants; review
  planning stops through the new `planner-failure` exit
  variant fed by enriched plan errors (missing
  reviewer-manifest entries now carry remediation, and
  stale manifest entries emit LOW drift advisories).
  Template: session-resume re-anchoring through
  `get_progress`/`parse_tasks`, phase-scoped
  `implement-phase-<N>` checkpoint records with failure
  inventory classification, mechanical-task tool-reference
  guidance, registered `replicator_forge_*` tool names,
  and a single `write_unleash_summary` registration (the
  redundant `persist_unleash_summary` alias registration
  removed; factory retained). Scaffold mirrors synced for
  every changed source including the new
  `opencode/lib/review-diff.ts` asset, with the asset
  inventories and the `TestRunInit_FreshDir` derived
  range ([70-74]) updated to match.
  (Spec: openspec/changes/unleash-pipeline-hardening/)
- clarification-marker-superset: Single-sourced the
  superset clarification-marker grammar: one shared
  `CLARIFICATION_MARKER_PATTERN` is consumed by
  `scan_clarifications`, `resolve_clarification`, and
  the `uf_unleash_detect_resume` clarify probe — the
  divergent dual-constant grammar (colon-only extractor
  pattern plus bare-substring `CLARIFICATION_MARKER_PREFIX`
  probe) is gone. The grammar accepts three forms —
  colon (`[NEEDS CLARIFICATION: q]`, scoped), bare with
  text (`[NEEDS CLARIFICATION text]`, unscoped), and
  completely bare (`[NEEDS CLARIFICATION]`) — bare and
  empty forms extract as unscoped questions whose text
  falls back to the raw marker text, so a hand-written
  bare marker resolves instead of wedging clarify
  (detector firing while the extractor returns nothing).
  The completely-bare and empty-colon fallback forms
  leave a self-matching Clarifications log mention;
  terminate by removing the literal marker syntax from
  that log entry. Pluralized (`[NEEDS CLARIFICATIONS]`) and unterminated
  marker-like text no longer holds clarify open; the
  probe now derives from the grammar per line,
  replacing the substring match. Command prose, lib doc
  comments, and scaffold mirrors updated. Includes the
  declared dependency update (design D7):
  `@opencode-ai/plugin` 1.4.10 → 1.19.0 in
  `.opencode/package.json` and `package-lock.json`,
  maintainer-directed — the gatekeeping exception's
  justification condition is unsatisfied, so the update
  is declared for the review council to judge.
  Pin-enforcement gates moved to the new pin: scaffold
  manifest mirrors, exact-version test constants,
  rollback baseline lock SHA256, the SC-FR-003
  "Reproducible Plugin Dependencies" spec record
  (specs/scaffold/spec.md delta), and the AGENTS.md
  Technology Stack pin. Nine transitive packages
  refreshed and two new transitives (`@ai-sdk/provider`
  3.0.8, `json-schema` 0.4.0); the `toml`/`uuid`
  security overrides hold; npm audit reports zero
  known vulnerabilities. Node engine floor recorded:
  `ini@7.0.0` requires ^22.22.2 || ^24.15.0 ||
  >=26.0.0, and CI's node-version: 22 resolves to
  v22.23.3, which satisfies it. The manifests are
  scaffold assets deployed via `uf init`, so the bump
  is changelog-relevant for scaffold consumers. (Spec:
  openspec/changes/clarification-marker-superset/,
  Refs: #743)
- address-feedback-model-normalization: `/uf.address-feedback`
  Phase 4.2 (Commit Changes) now instructs agents to call the
  existing tested `normalize_model_name` MCP tool with their
  raw model identifier and use the returned `normalized` value,
  replacing the inline 6-step model-name normalization prose;
  the warn-user fallback on `unknown-model` is preserved
  verbatim and the scaffold mirror was synced byte-identical.
  Command markdown only; no tool behavior change. Increment 2
  of the deterministic-instruction migration (Principle VI). (Spec:
  openspec/changes/address-feedback-model-normalization/,
  Fixes: #709)
- dispatch-finalize-determinism: Review dispatch finalization is
  now deterministic and single-call. `consolidate_dispatch`
  internally finalizes when given provenance `{branch, commit,
  workflow_id}` and returns the authoritative review-verdict v2
  projection; run-level `workflow_verdict` values are projected
  kind-native (triage: VALID/INVALID/NEEDS-CLARIFICATION;
  feedback: ACCEPT/AUTHOR-DECIDES) instead of generic council
  verdicts. `/uf.triage-issue` §4.6, `/uf.review-council` §5,
  and `/uf.address-feedback` §2.3.8 now make exactly one
  `consolidate_dispatch` call per finalization attempt —
  manual payload assembly and direct `finalize_review_dispatch`
  calls are prohibited. Failures classify deterministically
  (default-deny; only validation-stage rejections retry, at
  most twice) and fail closed to INCONCLUSIVE. Also fixes the
  `gh api --slurp` + `--jq` incompatibility in `/uf.triage-issue`
  comment fetching and adds a deterministic duplicate-search
  fallback. (Spec:
  openspec/changes/dispatch-finalize-determinism/, Docs: #717,
  Fixes: #712)
- move-schemas-to-uf: Scaffold schema assets relocated from
  `schemas/` to `uf/schemas/`. The `schemas/` prefix no longer
  has a dedicated mapping in `mapAssetPath`; schemas/ paths now
  use the default pass-through branch. (Spec:
  openspec/changes/move-schemas-to-uf/, Fixes: #649)
- command-prose-pilot: Rewrote `/uf.run-pipeline-plan`
  (both the `.opencode/commands/` copy and its
  byte-identical scaffold mirror) to a new house prose
  standard for agent instruction files: ten canonical
  control verbs (STOP, EXIT, PROCEED, RELAY, RESUME,
  SKIP, DELEGATE, ANNOUNCE, ASK, VERIFY) with a ban
  list, reference by name, one numbering axis, a
  20-word procedural sentence limit, decision tables
  for branching, and the Action/On-failure/Next step
  template. Behavioral contract preserved (three-
  subagent pipeline, gate relay, disk handoff
  verification, guardrails, tracking); three documented
  prose defects corrected: the "Step 8 (code review)"
  cross-command reference, the Gate Relay STOP collision
  (now RELAY), and the session-resume guard recast as a
  precondition. Before/after static metrics and
  comprehension A/B evidence recorded in the change's
  metrics.md. No CI gate added (conformance validated
  with change-time rg checks). (Spec:
  openspec/changes/command-prose-pilot/)

### Fixed
- fix-review-submission-validation: the review-dispatch submit tools
  now enforce their declared arg schemas at execution time.
  OpenCode 1.19.0 does not validate plugin tool arg schemas
  host-side, so `submit_review_findings` accepted findings missing
  required keys (e.g. `severity`) or carrying unrecognized ones and
  persisted them to run-*.json files; `finalize_review_dispatch`'s
  strict payload validation then rejected the consolidated payload
  ("invalid severity fields") only after the child session had
  ended, stranding /uf.unleash code-review dispatches at
  code-review-no-success with no recovery path. Both submit tools
  (`submit_review_findings`, `submit_lesson_proposal`) now
  `safeParse` their declared input schemas before storing and
  return a descriptive per-issue error (`<tool> validation failed:
  <path>: <message>`) the child reviewer can fix and resubmit
  in-session; nothing is stored on rejection. As defense-in-depth,
  `consolidate_dispatch` validates every harvested finding and
  proposal from successful runs against the same schemas, excludes
  invalid entries from the payload (per-run findings, top-level
  findings, verdict votes, and collected proposals), and reports
  them as `invalid_submissions_excluded` warnings (agent, run_id,
  invalid counts, per-entry validation errors bounded to the first
  20 lines with the total issue count) so the emitted payload is
  always finalize-shaped — by entry and by count: valid findings
  are also capped at the finalizer's declared cardinality bounds
  (512 per run, 512 consolidated after deduplication), with the
  overflow surfaced as `run_findings_overflow` and
  `consolidated_findings_overflow` warnings (agent, run_id, kept
  and overflow counts; kept and dropped counts). Schemas,
  thresholds, and limits are unchanged — the existing bounds are
  enforced earlier in the consolidate path, never modified. Plugin
  suite: rejection and defense-in-depth coverage added to
  submit-tools and consolidate-dispatch tests.
- forward-aggregate-correlation-id: all five review
  orchestration commands (`/uf.review-pr`,
  `/uf.review-council`, `/uf.triage-issue`,
  `/uf.address-feedback`, `/uf.unleash` at both dispatch
  sites) now capture the `correlation_id` reported by
  `dispatch_review_runs` and pass it as the explicit
  `correlation_id` on every subsequent
  `consolidate_dispatch` call (no-provenance retrieval,
  provenance finalization, retries), so cross-session
  consolidation (process restart mid-workflow, session
  handoff) succeeds instead of failing closed with
  `not_found`; the in-memory session map remains only
  as a same-session fallback. With the ID forwarded, a
  `not_found` for a dispatch that executed runs is
  reclassified as a permanent failure (design D5). The
  `dispatch_review_runs` tool description gained the
  forwarding guidance (RD-CF-002), and the behavior is
  locked by a cross-session consolidation test suite,
  a registration-surface schema lock, and Go command
  phrase contracts; scaffold mirrors synced
  byte-identical. (Spec:
  openspec/changes/forward-aggregate-correlation-id/,
  Fixes: #872, Docs: #879, unbound-force/website#348)
- fix-init-speckit-sentinel: `uf init` no longer silently skips
  the Speckit delegation on a fresh run. Since the
  starter-constitution asset landed (an embedded
  `.specify/memory/constitution.md`), the file-walk phase
  created `.specify/` before `initSubTools()` ran, which
  satisfied the bare `.specify` sentinel and suppressed
  `specify init` entirely -- every fresh non-`--force` `uf init`
  produced a constitution-only `.specify/` with none of the
  Speckit framework (scripts, templates, workflows, upstream
  speckit commands). The sentinel now detects a real install
  (`.specify/scripts`), the invocation runs
  `specify init --here --integration opencode --offline --force`
  (exactly one `--force`; specify 1.1.0 merges into non-empty
  targets and preserves user-owned files such as the starter
  constitution), and already-installed tooling renders a
  visible "already configured" line in the sub-tool summary
  instead of a silent skip. Users who want a constitution-only
  `.specify/` can set `setup.tools.specify.method: skip`.
  Regression coverage: `TestRun_SpecifyDelegationOnFreshInit`
  (full `scaffold.Run()`, the level where the original test gap
  hid the bug) and `TestSimpleToolsSentinelsNotScaffolded`
  (mechanical guard that no embedded asset ever maps under a
  sub-tool sentinel).
  (Spec: openspec/changes/fix-init-speckit-sentinel/,
  Docs: #827, unbound-force/website#325)
- fix-push-state-catch-to-zero: uf_finale_push_state no longer
  swallows rev-list count failures into zeroed ahead/behind
  counts (issue #821); counting failures and unparseable output
  now return a retryable PUSH_SYNC_COUNT envelope mirroring
  BASE_SYNC_COUNT, /uf.finale step 5 treats it as a soft gate
  with a safe push mode, and shared parsing moved to a tested
  parseRevCount helper. Scaffold mirrors updated byte-identically.
- fix-review-dispatch-correlation-release: `dispatch_review_runs`
  now mints a fresh correlation ID per invocation instead of
  binding one correlation ID per opencode conversation for the
  process lifetime, so consecutive review-dispatch workflows
  from the same conversation (spec review then code review in
  `/uf.unleash`, per-item dispatches in `/uf.address-feedback`,
  consecutive `/uf.triage-issue` or `/uf.review-pr` runs) no
  longer clobber each other's dispatch-session state: each
  invocation gets its own session directory, and prior
  sessions survive as audit trail instead of being
  overwritten. The tool's output now reports the invocation's
  `correlation_id` on every post-validation return path
  (success, all-skip, prompt-failure, budget-exhausted),
  making each dispatch session addressable by explicit ID.
  `consolidate_dispatch` and `dispatch_status` without an
  explicit `correlation_id` resolve the latest dispatch
  session; an explicit `correlation_id` retrieves an earlier
  session's payload. Direct `dispatch_agent_run` fan-out
  behavior is unchanged (one session per fan-out; a direct
  call after an aggregate dispatch joins the latest aggregate
  session). Fixes the root cause of #760 and #746. Regression
  tests cover the lifecycle in `dispatch-review-runs.test.ts`,
  `consolidate-dispatch.test.ts`, `dispatch-status.test.ts`,
  and `dispatch-agent-run.test.ts`, with shared fixtures in
  `.opencode/test/helpers/aggregate-dispatch-fixtures.ts`;
  scaffold mirror synced
  byte-identical. (Spec:
  openspec/changes/fix-review-dispatch-correlation-release/,
  Fixes: #794, Docs: #808, unbound-force/website#319)
- address-feedback-diff-path-input: `/uf.address-feedback` §2.3.4
  (`dispatch_review_runs` inputs) now documents the required
  `diff_path` input — path to the saved immutable diff file —
  and describes `changed_files` as a pre-formatted list of
  changed file paths, matching `/uf.review-council` §2 and the
  review-dispatch plugin schema (both are required strings for
  every non-triage mode, feedback included). The template
  previously omitted `diff_path` entirely and described
  `changed_files` as the signal's sorted affected-file objects,
  so an agent following it failed its first feedback-mode
  `dispatch_review_runs` call with an input-validation error
  and had to self-diagnose from plugin source; the golden
  feedback-variant comment in
  `.opencode/test/dispatch-review-runs.test.ts` now enumerates
  the corrected documented-input list. Command markdown only;
  no tool behavior change. The adjacent counts-only
  output-shape claim is corrected separately as PR #761
  feedback item 1 (issue #735) and is untouched here. Scaffold
  mirror
  `internal/scaffold/assets/opencode/commands/uf.address-feedback.md`
  synced byte-identical. (Fixes: #763)
- address-feedback-diff-path-producer: `/uf.address-feedback` §1.4
  "Fetch and Save the PR Diff" (new; former §1.4-§1.6 renumbered
  §1.5-§1.7) gives the §2.3.4 `diff_path` input (documented by
  address-feedback-diff-path-input) its producing step: one
  intent-level call to the existing registered
  `fetch_pr_review_context` tool (review-dispatch plugin) with the PR
  number; the returned `diff_path` is the immutable diff reused for
  every Tier 2 dispatch. Fetch failures report the tool's structured
  error envelope verbatim and STOP (`rate_limited` / `timeout` /
  `api_error`); oversized diffs surface at dispatch time as the
  structured `prompt_size_exceeded` error rather than a command-text
  size pre-check (determinism pack DE-001/DE-005). Hard-coded
  uf.address-feedback.md line-number site labels in
  `.opencode/test/golden-output-shape.test.ts` and
  `.opencode/test/dispatch-review-runs.test.ts` recomputed from the
  final template. Command markdown only, referencing an existing
  registered tool; no tool behavior change. Scaffold mirror
  `internal/scaffold/assets/opencode/commands/uf.address-feedback.md`
  synced byte-identical. (Refs: #811)
- fix-dispatch-verdict-derivation: `consolidate_dispatch` now
  derives `workflow_result.value` and `verdict` from
  `calculatedWorkflowValue` — the same vote-based derivation
  function the finalizer validates with — instead of the
  severity ladder over consolidated findings, so emission and
  validation share one derivation and cannot disagree. Fixes
  deterministic finalization rejection of feedback split
  panels (any successful zero-findings ACCEPT run plus
  findings-bearing runs), triage split-majority panels, and
  zero-successful-runs panels: those panel shapes previously
  could never finalize — rejected with identical value-mismatch
  errors that burned the bounded mechanical-retry budget and
  fail-closed INCONCLUSIVE with no persisted artifact (observed
  on PR #761). `verdict_reason` and advisories remain
  severity-derived; council and test-review outputs are
  unchanged (both derivations agree there). Four regression
  tests added in `.opencode/test/consolidate-dispatch.test.ts`;
  scaffold mirror
  `internal/scaffold/assets/opencode/plugins/review-dispatch/index.ts`
  synced byte-identical. (Spec:
  openspec/changes/fix-dispatch-verdict-derivation/,
  Fixes: #764)
- fix-review-dispatch-recovery: Review-dispatch sessions can no
  longer be bricked by a failed child run's accepted submissions:
  `consolidate_dispatch` now scopes its finalization payload to
  successful runs, excluding orphaned findings and proposals with
  machine-readable `orphaned_submissions_excluded` warnings
  (agent, run id, counts), and a reviewer run that aborts
  mid-final-response after submitting (child message error) is
  reclassified as a success run with a `post_submission_abort`
  audit record so its submissions count. `consolidate_dispatch`
  and `dispatch_status` accept an optional explicit
  `correlation_id` (recovery from a fresh session after a
  restart) and `dispatch_status` reports per-run token usage; run
  prompts are persisted and session metadata refreshed on
  re-dispatch, enabling reliable fix-loop re-runs. Two new
  recovery tools in the review-dispatch plugin:
  `discard_orphaned_submissions(correlation_id, agents?)` empties
  orphaned submissions from failed runs with an audit block
  (refuses successful runs), and
  `redispatch_failed_runs(correlation_id, agents?, timeout?)`
  re-runs failed reviewers from their original prompts and plan
  entries, archiving prior run files under `superseded/` (a re-run
  that exits before execution restores the prior run file to its
  active path, keeping the session consolidatable).
  Fail-closed semantics preserved: zero-success dispatches still
  finalize INCONCLUSIVE; finalizer validation unchanged.
  consolidate_dispatch results now always carry top-level
  warnings/proposals observability fields on the no-provenance
  path; callers forwarding to the external
  finalize_review_dispatch tool must strip them (documented in
  the tool description). (Spec:
  openspec/changes/fix-review-dispatch-recovery/, Fixes: #751)
- fix-tool-args-schema-mismatch: uf-workflow (8) and uf-finale
  (3) tool factories now declare parameters via the
  SDK-recognized zod `args` contract instead of plain
  `parameters:` objects that @opencode-ai/plugin silently
  ignored, restoring model-visible schemas and runtime
  argument passing for 11 tools (checkpoint,
  get_progress, clear_progress, normalize_model_name,
  git_fetch, git_push, git_merge, git_rebase,
  uf_finale_commit_context, uf_finale_create_pr,
  uf_finale_ci_poll). Names, types, optionality, and
  descriptions preserved exactly; execute bodies, result
  envelopes, and runtime validation guards untouched
  (defense-in-depth). Adds 12 schema-exposure regression
  tests (sorted-args key equality, including
  persist_unleash_summary as a future-regression guard)
  and byte-identical scaffold mirrors. (Spec:
  openspec/changes/fix-tool-args-schema-mismatch/,
  Refs: #732, Docs: #740, unbound-force/website#310)
- fix-tui-git-output-leak: Removed all 50 console.log/console.error
  calls per tree (100 total across `.opencode/plugins/` and the
  embedded `internal/scaffold/assets/opencode/` mirrors) from
  uf-workflow and uf-finale plugin tool handlers. Tool output now
  flows exclusively through JSON result envelopes, fixing TUI
  alternate-screen corruption caused by plugin handlers writing
  to the shared terminal. `uf_finale_stage` success envelope
  enriched with `files: string[]` (staged file list, previously
  TUI-only); ci_poll 30-second progress messages removed. Adds
  spy-based zero-terminal-writes regression tests for every
  handler plus files-field unit tests. (Spec:
  openspec/changes/fix-tui-git-output-leak/, Docs: #710,
  unbound-force/website#306)
- fix-stale-base-ref: New `resolve_base_ref` tool in uf-workflow
  plugin with three-level fallback (`upstream/main` → `origin/main`
  → `main`) replaces hardcoded `main` ref in review-council local
  path. Prevents phantom files in diffs when local `main` is stale.
  Also removes token-wasting Key Files Changed block from
  uf.unleash demo output. (Spec:
  openspec/changes/fix-stale-base-ref/, Closes: #687)

### Added
- fetch-pr-review-context: New `fetch_pr_review_context` tool in
  the review-dispatch plugin. Fetches all PR metadata needed for
  review dispatch in one deterministic call: PR metadata, CI
  checks, diff (persisted to disk with 0o600 permissions), existing
  reviews, inline comments, and pre-computed derived fields. Uses
  Bun.spawn with argument arrays for shell injection resistance.
  (Spec: openspec/changes/fetch-pr-review-context/, Docs: #696)
- uf-workflow-scaffold: New `uf-workflow` plugin scaffold with
  `ToolSuccess<T>`, `ToolFailure`, `ToolResult<T>` result envelope
  types and `success()`/`failure()` helper functions. Includes plugin
  entry point, canonical scaffold copies, opencode.json registration,
  activation gating, and vitest test harness. (Spec:
  openspec/changes/uf-workflow-scaffold/, Closes: #682)
- tier-gated-agent-count: Review dispatch now enforces tier-based
  agent count caps. Lightweight diffs cap at 2 agents by default;
  standard and heavy remain uncapped. Always-required agents
  survive caps via floor enforcement. A HIGH severity advisory
  warns when a custom heavy cap is applied to security-sensitive
  diffs. The `--full` flag bypasses all caps. Caps are
  configurable via `limits.tier_caps` in the review matrix.
  (Spec: openspec/changes/tier-gated-agent-count/, Closes: #654)
- optional-profile-models: Review matrix profile `model` fields
  are now optional — profiles without a model fall back to host
  model resolution at invocation time. Added
  `.uf/review-matrix.override.yaml` for per-developer model and
  variant customization with a whitelist constraint (only
  `profiles.<tier>.[model, variant]` paths permitted). Bumped
  review-matrix schema to version 3. (Spec:
  openspec/changes/optional-profile-models/, Fixes: #636)
- issue-pr-traceability: OpenSpec changes now support an
  optional `originating_issue` field in `.openspec.yaml`.
  `/uf.finale` reads the field and emits `Closes #<N>`
  after `## Summary` in the PR body when present. Absence
  of the field is backward compatible (no `Closes` line
  emitted). (Spec: openspec/changes/issue-pr-traceability/,
  Closes: #554)
- uf-init-protect-block: `/uf.init` slash command and scaffold asset
  now wrap core instructions in `<protect>` blocks to prevent
  context-compaction-induced instruction loss. Updated insertion
  logic to respect existing `<protect>` blocks in target files.
  (Spec: openspec/changes/uf-init-protect-block/, Closes: #540)
- run-pipeline-plan: New `/uf.run-pipeline-plan` slash command
  that orchestrates a prioritized work plan through the full
  Unbound Force pipeline (`/opsx-propose` → `/uf.unleash` →
  `/uf.finale`) with fresh subagent isolation per phase,
  human gate relay, and per-item progress tracking. Supports
  plan file and GitHub issue-ref input modes. (Spec:
  openspec/changes/run-plan-command/)

### Fixed
- content-portability-guardrails: Add Content Portability
  rules (CP-001, CP-002, CP-003) to the default convention
  pack and targeted guardrails in four issue-generating
  components (triage child issues, tasks-to-issues command,
  uf.init taskstoissues template, divisor-curator agent).
  Prevents LLM-generated issue content from using
  org-config-specific terminology for non-org-config repos.
  (Spec: openspec/changes/content-portability-guardrails/,
  Fixes: #593)
- devcontainer-forward-ports: Podman sandbox commands
  (`uf sandbox create`, `uf sandbox start`) now automatically
  read `forwardPorts` from `.devcontainer/devcontainer.json`
  and publish them via `-p` flags. Supports numeric ports,
  `"host:container"` string mappings, and JSONC comments/
  trailing commas. `uf sandbox create` (persistent path)
  deduplicates against both `DefaultServerPort` and
  `--demo-ports`; `uf sandbox start` (ephemeral path)
  deduplicates against `DefaultServerPort`. (Fixes: #595)
- fix-sandbox-persistent-workdir: `buildPersistentRunArgs()`
  now sets `--workdir` and `WORKSPACE` env var for persistent
  Podman workspaces. OpenCode starts in the project directory
  instead of `/`. Mirrors the ephemeral-path fix from PR #123.
  Removed private function `buildPersistentRunArgs` from CRAP
  baseline (Gaze only analyzes exported functions).
  (Spec: openspec/changes/fix-sandbox-persistent-workdir/)
- fix-init-force-passthrough: `uf init --force` now
  correctly re-initializes Group B sub-tools (specify,
  replicator, openspec, gaze) and passes `--force`
  through to sub-tool CLIs. Previously, `initSimpleTool`
  silently skipped re-initialization regardless of the
  force flag.
  (Spec: openspec/changes/fix-init-force-passthrough/,
  Fixes: #479)

### Changed
- constitution-v1.3.0-gatekeeping: Aligned AGENTS.md
  gatekeeping rule with constitution v1.3.0 language
  (amended in prior PR #607). Gatekeeping Integrity now
  permits narrowly scoped dependency pin updates with
  immutable SHA verification, gate preservation, rollback
  documentation, and reproducible validation. Explicitly
  prohibits mutable tags and other name-addressed refs.
  (Fixes: #610)
- add-todowrite-to-pipelines: Pipeline commands
  (/uf.unleash, /uf.finale, /uf.review-council,
  /uf.address-feedback) now use TodoWrite for live
  session progress visibility alongside the existing
  Edit tool execution checklist.
  (Spec: openspec/changes/add-todowrite-to-pipelines/,
  Fixes: #504)
- reconcile-speckit-uf-refs: `uf init` Step 6 now injects
  Dewey (`dewey_semantic_search`) and
  `.specify/memory/constitution.md` references into the
  spec-phase, implement, and constitution speckit guardrail
  blocks, plus `/uf.review-council` into the implement block
  only. Step 6 idempotency checks were tightened so
  already-initialized repos re-running an upgraded binary
  re-inject the new references (self-healing on upgrade)
  instead of being skipped. Gated by
  `TestSpeckitTemplates_RequiredReferences`.
  (Spec: openspec/changes/reconcile-speckit-uf-refs/,
  Fixes: #549)
- scaffold-agents-md-stale-refs: `uf init` stale command
  reference detection now also scans the root `AGENTS.md`
  file (in addition to `.opencode/agents/*.md`), so
  references to pre-namespace command names in the
  project's own AGENTS.md are surfaced. Matching is now
  word-boundary anchored via `regexp`, so an
  already-migrated reference (e.g. `/uf.review-council`)
  is no longer falsely flagged as containing the stale
  form (`/review-council`). Detection remains warn-only;
  the user-owned `AGENTS.md` is never modified.
  (Spec: openspec/changes/scaffold-agents-md-stale-refs/,
  Fixes: #567)
- refine-bash-permissions: Refine `permission.bash`
  `gh api` rules in `opencode.json`. Read-only API
  calls (implicit or explicit GET) execute without
  prompting; mutating methods in all flag forms (`-X`,
  `-XPOST` glued, `--method`, `--method=`, each in
  uppercase and lowercase) and data-sending flags
  (`-f`, `-F`, `--field`, `--raw-field`, `--input`,
  including `=` variants) require approval. All
  original mutation guards for `gh issue`, `gh pr`,
  `git push`, `git commit`, and `rm` preserved.
  (Spec: openspec/changes/refine-bash-permissions/,
  Fixes: #531)
- pre-flight-file-scope-filter: Add Phase 2a file-scope
  filtering to skip tools that have zero in-scope files
  in the branch diff. Applies to all three execution
  modes (hard-gate, ci-aware, soft-gate). Conservative
  fallbacks ensure all tools run when diff computation
  fails, the repository is a shallow clone, or the diff
  is empty on a feature branch. Aggregate tools
  (make check, pre-commit) are always in scope. Tool
  config files are included in their respective tool
  scopes.
  (Spec: openspec/changes/pre-flight-file-scope-filter/,
  Fixes: #434)
- triage-mandatory-gate: Add `>>> MANDATORY GATE <<<` to
  `uf.triage-issue` Phase 4 entry. Agents now stop and
  require explicit human confirmation before executing
  mutations (labels, comments, child issues). Includes
  compressed-context resume guard and mutation safety
  reiteration.
  (Spec: openspec/changes/triage-mandatory-gate/,
  Fixes: #473)
- adopt-org-infra-release-workflows: Replace inline
  release preflight and GoReleaser jobs with org-infra
  reusable workflow callers (`reusable_release_preflight`
  + `reusable_release_goreleaser` @ v0.7.1). Adds smart
  re-run detection, spec-compliant semver comparator
  (replaces `sort -V`), configurable `ci_checks` input
  (build, lint, and security scan gates), skip inputs
  for debugging, and concurrency guard. Signing secrets
  check expanded to verify all 6 required secrets.
  GoReleaser config gains `release.extra_files` for
  Homebrew artifact upload. `sign-macos` stays inline.
  Updated `docs/RELEASE_PROCESS.md` to reflect the new
  pipeline, including org-infra upgrade path.
  (Spec: openspec/changes/adopt-org-infra-release-workflows/,
  Fixes: #428)

### Added
- review-council-multi-model-fanout: Replace single-host
  Divisor delegation with explicit-first multi-model fan-out
  via two OpenCode plugins registered in `opencode.json`
  (`.opencode/plugins/invoke-agent` and `review-dispatch`)
  driven by a closed reviewer manifest
  (`.uf/reviewer-capabilities.yaml`) and an explicit-first
  hybrid matrix (`.uf/review-matrix.yaml`). Adds bounded
  sibling-repository evidence (`.uf/sibling-repos.yaml`),
  source-grounded lesson-proposal provenance, staged
  reproducible plugin activation in `uf init`, `uf doctor`
  review-plugin checks, a new scope-based Go coverage gate
  (`internal/coveragegate`, `cmd/coverage-gate`,
  `coverage-gate.json`), and `make plugin-test` and
  `make coverage-gate` targets (both wired into `make check`
  and `ci_local.yml`). Emits a Hero Interface Contract
  review-dispatch envelope and versions the canonical
  review-verdict artifact to 2.0.0 with native
  `INCONCLUSIVE` and `UNAVAILABLE` decisions.
  (Spec: openspec/changes/review-council-multi-model-fanout/,
  Fixes: #635)
- uf-init-stealth-mode: `uf init` gains a `--stealth` flag for
  local-only scaffolding that leaves the working tree git-clean
  and tracked files (`.gitignore`, `AGENTS.md`) unmodified. Uses
  `.git/info/exclude` for exclusion, snapshots and restores all
  tracked files (bytes and mode/permission bits), and provides a
  `--check` flag for verifying git-cleanliness without writing.
  (Spec: openspec/changes/uf-init-stealth-mode/, Fixes: #638,
  Website issue: unbound-force/website#292)
- add-speckit-ref-assertion-test: Add red-first
  content-assertion test `TestSpeckitTemplates_RequiredReferences`
  (internal/scaffold) that verifies Step 6 of the embedded
  `uf.init.md` injects the required Unbound Force references —
  Dewey (`dewey_semantic_search`) and
  `.specify/memory/constitution.md` for all five in-scope
  speckit files, plus `/uf.review-council` for the implement
  file only (implement-only boundary enforced via
  `mustNotContain`). Regression gate anchored to ADR-002.
  (Spec: openspec/changes/add-speckit-ref-assertion-test/,
  Fixes: #548)
- adversary-concurrency-checklist: Add "Concurrency Correctness"
  section (5 items: goroutine leaks, deadlock/lock-ordering, unsafe
  concurrent access, `context.Context` propagation, `sync.WaitGroup`
  misuse) to `divisor-adversary.md`, scoped to production code. Adds
  an Adversary/SRE Out of Scope boundary row (concurrency
  correctness → The Adversary; concurrency efficiency → The SRE) and
  scopes test-code concurrency to The Tester. Synced to the embedded
  scaffold mirror.
  (Spec: openspec/changes/adversary-concurrency-checklist/,
  Fixes: #469)
- scaffold-dcp-config: `uf init` now creates `.opencode/dcp.jsonc`
  with `protectTags: true`, enabling `<protect>` tag preservation
  during DCP context compression. Without this config, `<protect>`
  tags scaffolded in slash commands were inert. Follows the
  idempotent `configureOpencodeJSON()` pattern: skips if already
  configured, adds `protectTags` if missing, respects `--dry-run`
  and `--force` flags.
  (Spec: openspec/changes/scaffold-dcp-config/, Fixes: #502)
- mandatory-gate-hardening: Add mandatory confirmation
  gates with session-resume guards to 4 command files
  (uf.address-feedback Phase 4 entry, uf.review-pr
  fix-branch commit, uf.review-council spec auto-fix,
  uf.finale commit and push). Prevents agents from
  executing mutations without human consent under
  compressed context.
  (Spec: openspec/changes/mandatory-gate-hardening/,
  Fixes: #474)
- granular-bash-permissions: Runtime-enforced bash permission
  rules in `opencode.json` gating 14 GitHub-mutating, git-write,
  and destructive commands (`gh issue create/edit/close/comment`,
  `gh pr create/merge/close/comment/edit/review`, `gh api`, `git push`,
  `git commit`, `rm`) behind `"ask"` approval prompts. Read-only
  commands remain auto-allowed via `"*": "allow"` catch-all.
  Hardened `divisor-curator.md` with granular frontmatter
  permissions (default deny) replacing prose-only restrictions.
  Structural fix for #465 triage runaway where prose-based
  confirmation gates were lost under context compression.
  (Spec: openspec/changes/granular-bash-permissions/,
  Fixes: #465)
- council-review-action: Reusable composite GitHub Action
  for multi-persona AI code review on PRs; discovers
  Divisor reviewer personas from the repo's
  `.opencode/agents/` directory, pre-fetches PR context,
  and outputs structured JSON for inline review comments;
  three-tier persona fallback (repo, bundled,
  single-agent); diff noise filtering and line annotation
  for accurate inline comment placement
  (Spec: openspec/changes/council-review-action/,
  Closes: #253)

### Fixed
- fix-ask-user-question-tool-name: Rename all `AskUserQuestion`
  tool references to `question` across 24 command, agent, skill,
  and scaffold files. The actual OpenCode tool is named `question`;
  the incorrect name could cause agents to silently skip
  interactive confirmation gates.
  (Spec: openspec/changes/fix-ask-user-question-tool-name/,
  Fixes: #475)
- fix-constitution-scaffold: `uf init` now scaffolds a working
  starter constitution at `.specify/memory/constitution.md`
  instead of delegating to `specify init` (which fails
  silently in non-interactive mode). The starter includes
  all five org principles with generic project language and
  no `[PLACEHOLDER]` tokens. Post-scaffold hints updated
  from "create" to "customize".
  (Spec: openspec/changes/fix-constitution-scaffold/,
  Fixes: #213)
- shield-constitution-from-force: `uf init --force` no
  longer overwrites `.specify/memory/constitution.md`,
  preserving user governance customizations. Adds an
  `isNeverOverwrite` guard that protects the constitution
  from force-overwrite while still allowing first-time
  creation from the embedded starter template.
  (Spec: openspec/changes/shield-constitution-from-force/,
  Fixes: #495)
- fix-doctor-hints: Replace five hardcoded OpenCode-specific
  `InstallHint` strings in `uf doctor` agent context checks
  with tool-agnostic plain-language remediation instructions.
  Adds regression guards for `/uf.` and `OpenCode` references
  in hints, and a drift-detection test for Tier 1 sections.
  (Spec: openspec/changes/fix-doctor-hints/, Fixes: #460)
- fix-unleash-autonomy: `/uf.unleash` no longer halts
  between pipeline steps to ask for human confirmation.
  Added an inter-step autonomy guardrail and explicit
  "proceed immediately" transition instructions at all
  10 step checkpoints (Steps 0-9). All existing
  STOP/EXIT conditions (HIGH/CRITICAL spec findings,
  build failures, merge conflicts, review iteration
  exhaustion) remain unchanged.
  (Spec: openspec/changes/fix-unleash-autonomy/,
  Fixes: #457)
- fix-specify-init-invocation: Fixed `uf init` failing
  to create `.specify/` directory due to upstream
  specify-cli interface changes. Updated specify
  invocation from bare `specify init` to
  `specify init --here --integration opencode --offline`
  to match the new CLI contract.
  (Spec: openspec/changes/fix-specify-init-invocation/,
  Fixes: #216)
- unleash-demo-output-fidelity: `/uf.unleash` Step 10
  Demo now includes a JIT re-read guard ensuring
  prescribed Next Steps output is reproduced verbatim
  from the template, not improvised after compression.
  Step 6 spec review exits now provide tier-aware
  recovery guidance, directing OpenSpec users to update
  change artifacts instead of running Speckit commands.
  Reconciles divergent Next Steps blocks into one
  canonical format. Adds anti-improvisation guardrail
  and Step 8 equivalence note (review council already
  runs in Step 8). Scaffolds `always-on-guidance` skill
  for org-wide propagation via `uf init`. Clarifies
  AGENTS.md review-council rule. Extends reverse-drift
  test to cover `.opencode/skills/`.
  (Spec: openspec/changes/unleash-demo-output-fidelity/)
- fix-rpm-version-skew: `uf setup` RPM install path now
  resolves each companion tool's latest release
  independently via `gh release view` instead of using
  the uf binary's own version. Fixes 404 errors on
  Fedora/RHEL when gaze or replicator versions diverge
  from uf. Adds `ResolveRelease` injectable dependency
  to `Options` for testability. Reorders `buildSteps()`
  so GitHub CLI is installed before Gaze.
  (Spec: openspec/changes/fix-rpm-version-skew/,
  Fixes: #455)
- fix-incorrect-guardrails: Step 6 of `/uf.init` now
  injects command-specific guardrails for
  `speckit.implement.md`, `speckit.constitution.md`,
  and `speckit.taskstoissues.md` instead of sharing
  the spec-phase guardrails that contradicted each
  command's operational model. Existing repos with
  incorrect guardrails are automatically corrected
  on next `/uf.init` run via replace-if-incorrect
  idempotency. Added guardrail content regression
  tests.
  (Spec: openspec/changes/fix-incorrect-guardrails/,
  Fixes: #256)
- uf-init-rename-migration: Hardened `uf init` command
  rename migration with warnings for stale agent
  references and removal failures. `cleanupRenamedCommands`
  now reports non-removable files; `warnStaleCommandRefs`
  scans agent files for old command paths.
  (Spec: openspec/changes/fix-uf-init-command-rename-migration/,
  Fixes: #419)
- fix-review-pr-step7-gate: Step 7 verdict-posting offer
  in `/uf.review-pr` was gated behind a HIGH+ findings
  condition, causing the `AskUserQuestion` to be skipped
  when a review had only MEDIUM/LOW findings or zero
  findings. Restructured Step 7 so the framing text is
  conditional on HIGH+ findings while the posting offer
  always executes.
  (Spec: openspec/changes/fix-review-pr-step7-gate/,
  Fixes: #441)
- fix-devpod-provider: Removed incorrect `LookPath("podman")`
  pre-flight check from `DevPodBackend.Create()`. The docker
  provider aliased as "podman" (configured by `uf setup`) does
  not require the podman binary in PATH. Added diagnostic hint
  ("run 'uf doctor' to diagnose or 'uf setup' to configure")
  to `devpod up` error paths in both `Create()` and `Start()`.
  (Spec: openspec/changes/fix-devpod-provider/,
  Fixes: #431)
- mutimind-acceptance-gate: Added AskUserQuestion
  confirmation gate before acceptance decision CLI
  (`go run cmd/mutimind/main.go decide`) in
  muti-mind-po.md agent. Prevents irreversible
  governance action from firing without user
  confirmation.
  (Spec: openspec/changes/mutimind-acceptance-gate/,
  Fixes: #351)

### Removed
- remove-bridge-file-scaffolding: Removed CLAUDE.md and
  .cursorrules bridge file scaffolding from `uf init` and
  bridge file checks from `uf doctor`. OpenCode is the only
  officially supported platform; bridge files for Claude Code
  and Cursor were experimental and are no longer maintained.
  Existing bridge files in user repos are not deleted.
  Updated `/uf.agent-brief` command to remove bridge file
  verification step and references. Cleaned up QUICKSTART.md
  and architecture docs.
  (Spec: openspec/changes/remove-bridge-file-scaffolding/,
  Fixes: #459)

### Changed
- council-review-sandbox: Runtime permission sandbox for
  council review action; enforces tool restrictions via
  `OPENCODE_CONFIG_CONTENT` permission config (denies
  bash, edit, webfetch, websearch, skill); `--pure` flag
  isolates external MCP plugins; explicit deny of
  ask-default permissions for headless CI; migrated 9
  Divisor agents from deprecated `tools:` to `permission:`
  frontmatter syntax
  (Spec: openspec/changes/council-review-sandbox/)
- All 10 uf-owned slash commands renamed to `uf.`
  dot-notation namespace prefix: `/address-feedback`
  → `/uf.address-feedback`, `/agent-brief` →
  `/uf.agent-brief`, `/cobalt-crush` →
  `/uf.cobalt-crush`, `/constitution-check` →
  `/uf.constitution-check`, `/finale` → `/uf.finale`,
  `/review-council` → `/uf.review-council`,
  `/review-pr` → `/uf.review-pr`, `/triage-issue` →
  `/uf.triage-issue`, `/uf-init` → `/uf.init`,
  `/unleash` → `/uf.unleash`. Scaffold engine gains
  migration map for automatic orphan cleanup on
  `uf init` re-run. See issue #302 for migration
  guide.
  (Spec: openspec/changes/uf-command-namespace-prefix/,
  Closes: #302)
- `/triage-issue` label gates: all label mutations (create
  and apply) now require explicit user confirmation via
  AskUserQuestion, not just the `duplicate` label. Duplicate
  label retains supplementary close-semantics confirmation.
  (Spec: openspec/changes/review-council-label-gates/,
  Fixes: #352)
- `/speckit.clarify`: added mandatory AskUserQuestion tool
  call gate to sequential questioning loop (step 4);
  prevents question batching and answer assumption under
  context compression
  (Spec: openspec/changes/speckit-clarify-askuser-gate/,
  Fixes: #364)
- `/review-council` Code Review Mode: added optional
  Step 7 (GitHub Review Posting) for posting consolidated
  multi-persona council findings as a GitHub PR review;
  PR detection via `gh pr view` or explicit PR number
  argument with input validation; review state awareness
  (duplicate detection, stale review warnings, CODEOWNER
  warnings); multi-persona finding aggregation with
  per-persona sections and severity-first round-robin
  inline comment allocation (capped at 15); verdict
  mapping (APPROVE, REQUEST_CHANGES, COMMENT); human
  confirmation required before posting; graceful
  degradation when `gh` unavailable or API errors;
  Protocol 2 (Issue Linking) conditionally enabled when
  explicit PR number provided; review body size cap at
  60,000 characters with truncation
  (Spec: openspec/changes/review-council-github-posting/,
  Fixes: #314)
- `/review-council` Code Review Mode: added Phase 1c
  (review-context skill) for spec artifact discovery,
  path classification, and walkthrough generation;
  restructured Step 2 context enrichment to pass review
  context and quality context to Divisor agents; added
  review context summary to Step 6 final report
  (Fixes: #296)
- `/finale` PR body generation: structured sections (Summary,
  How to Test, How to Demo, Key Files Changed) replacing
  minimal bullet-point summary; PR template detection and
  respect for `.github/PULL_REQUEST_TEMPLATE.md`; unresolved
  `/review-council` findings surfaced in Known Issues section;
  AI attribution footer on PR descriptions; AI attribution on
  commit messages (git trailer + human-readable footer); shell
  injection fix switching from inline `--body` to `--body-file`
  (Spec: openspec/changes/improve-finale-pr-description/,
  Closes: #142)
- Cross-command AI attribution standardization: unified
  `Assisted-by: <model>` trailer format across `/finale`,
  `/address-feedback`, and `/review-pr`; model name included
  in attribution (derived from agent system configuration);
  character-set validation prevents trailer injection;
  `unknown-model` fallback with user warning
  (Spec: openspec/changes/improve-finale-pr-description/)

### Fixed
- `gaze-test-generator` agent: replaced advisory compile
  verification prose ("ALWAYS verify generated code compiles")
  with a concrete pre-write compile gate protocol. The agent
  now MUST run `go build` before any Write or Edit tool call
  and halt if compilation fails. Resolves T3 weakness where
  the advisory check could be skipped under context compression.
  Also resolves existing drift between the active copy (242
  lines) and the canonical scaffold copy (278 lines).
  (Spec: openspec/changes/fix-gaze-test-generator-compile-gate/,
  Fixes: unbound-force/gaze#204)
- `speckit.specify`: Dirty-tree check before branch creation now
  uses an explicit `AskUserQuestion` tool call with structured
  options ("Stash changes and continue" / "Abort -- keep changes
  as-is") instead of prose-only instructions. Includes `git stash`
  failure handling and stash recovery notification. Hardens against
  context compression skipping the guard (T1+T3 weakness).
  (Spec: openspec/changes/fix-speckit-dirty-tree-check/,
  Fixes: #358)
- `uf init --force` no longer hangs on Dewey re-indexing.
  Dewey indexing now runs with `--no-embeddings` for a fast
  metadata-only pass. Run `dewey index` separately to
  generate embeddings for semantic search. (Fixes: #163)

### Added
- `ci-convention-pack`: Added CI convention pack (`ci.md`,
  `ci-custom.md`) with 12 rules across 5 sections (Action
  Pinning & Supply Chain, Workflow Structure, Permissions &
  Secrets, Reusable Workflow Design, Custom Rules). Fills
  the Divisor SRE `[PACK]` CI/CD guidance gap. CI-002 rule
  prevents SHA hallucination by requiring tag-first lookup
  at authoring time. Always-deployed (language-agnostic),
  following the `content.md`/`severity.md` pattern.
  Centralizes CI rules from `coding-standards.md` into the
  convention pack system. Related: complytime/org-infra#337
  (CI guardrail).
  (Spec: openspec/changes/ci-convention-pack/)
- `/triage-issue` slash command for multi-agent GitHub issue
  triage using the Divisor review panel. Classifies issues
  (bug, feature, enhancement, question, opinion, duplicate,
  needs-clarification), determines validity via majority
  consensus, checks for duplicates, recommends splitting
  compound issues, and posts tone-appropriate comments.
  (Spec: openspec/changes/triage-issue/)
- `issue-triage` JSON schema (v1.0.0) for capturing triage
  decisions as artifacts with per-agent assessment provenance
  (Spec: openspec/changes/triage-issue/)
- `python-convention-pack`: Added Python convention pack
  (`python.md`, `python-custom.md`) with 46 rules across 7
  sections (Coding Style, Architectural Patterns, Security
  Checks, Testing Conventions, Type Annotations, Documentation
  Requirements, Custom Rules). Expanded `detectLang()` with 5
  additional Python markers (`setup.py`, `setup.cfg`,
  `requirements.txt`, `tox.ini`, `Pipfile`). Added conditional
  "Python Tools" doctor check group with 9 tool categories
  and tool-agnostic alternative detection (ruff satisfies
  formatter, linter, import sorter, security scanner).
  (Spec: openspec/changes/python-convention-pack/)
- `/address-feedback` slash command for structured PR review
  feedback triage (Spec: openspec/changes/address-feedback/)
- `feedback-triage` JSON schema (v1.0.0) for capturing triage
  decisions as artifacts (Spec: openspec/changes/address-feedback/)
- Constitution Principle V (Security by Default) covering
  supply chain integrity, input validation, least privilege,
  and dependency necessity (version 1.1.0 → 1.2.0)
  (Spec: openspec/changes/review-finding-consolidation/)
- Compound Severity Escalation section in `severity.md` —
  instructs personas to assess combined severity when
  multiple findings share a root cause
  (Spec: openspec/changes/review-finding-consolidation/)
- Severity Calibration step (8g) in `/review-pr` — forces
  re-reading `severity.md` definitions to counter anchoring
  bias (Fixes #233)
- Adversarial Input Enumeration in `divisor-adversary.md`
  (new checklist item 5) and `/review-pr` Step 8b — per-input
  threat analysis for new parameters/secrets
  (Fixes #233)
- Issue Suggestion Gap Detection in `/review-pr` Step 8a —
  checks linked issues for unimplemented code suggestions
  (Fixes #233)
- CI Bot Annotation Cross-referencing (Step 8e) and Finding
  Consolidation (Step 8f) in `/review-pr` — ingests bot
  findings as corroborating evidence and merges related
  findings by root cause (Fixes #228)
- Cross-persona finding consolidation in `/review-council`
  Code Review Step 3 and Spec Review Step 2 (Fixes #228)
- Dependency necessity check and content-integrity
  verification guidance in `divisor-adversary.md` Audit
  Checklist item 2 (Fixes #228)

### Changed
- `uf init` sub-tool initializations (dewey, replicator,
  specify, openspec, gaze) now run concurrently, reducing
  wall-clock time from sequential sum to parallel max.
  Dewey indexing (30-60s) no longer blocks the other four
  tools. Individual tool failures remain non-fatal.
  (Spec: openspec/changes/parallel-subtool-init/)
- `uf init` now respects `setup.tools.<name>.method: skip`
  in `.uf/config.yaml` — skipped tools are excluded from
  initialization entirely, even when the binary is on PATH.
  (Spec: openspec/changes/parallel-subtool-init/)
- fix-triage-issue-dispatch-docs: Corrected the dispatch
  data-flow documentation across the four dispatch-driving
  commands. `/uf.triage-issue` §2.3/§2.4 (and the mirrored
  §4/§5 sections in `/uf.review-council`, §2.3.7/§2.3.8 in
  `/uf.address-feedback`, §5a in `/uf.review-pr`) now
  document the counts-only `dispatch_review_runs` output
  contract, the single sanctioned no-provenance
  `consolidate_dispatch` retrieval (side-effect-free; the
  exact same payload feeds proposal storage, verdict
  resolution, and finalization), the validated
  `{agent, information, tag, category}` proposal shape
  (there is no `reference` field), the dual-mode
  exactly-once scoping, and FR-001 transient/permanent/all-
  skip failure behavior. `/uf.triage-issue` §3.1 verdict
  resolution now names run-recorded `workflow_verdict` as
  the governing vote source and sources assessment fields
  (category, objectivity, reasoning, split recommendations)
  exclusively from the new additive `read_assessment_fields`
  review-dispatch plugin tool (+257 lines, +36 tests, bounds:
  128/64/4096 chars, 16x512 split entries; DE-001
  extraction), which reads persisted child text from the
  dispatch store with tool-owned untrusted-input validation.
  Adopted the maintainer-approved `@opencode-ai/plugin`
  1.4.10 -> 1.19.0 pin with its full lockstep set (manifests
  + scaffold mirrors, test expectations, provenance/
  rollback testdata, Node engine range (^22.22.2 ||
  ^24.15.0 || >=26.0.0) declared in `engines` fields,
  AGENTS.md/README documentation) as
  dedicated commit 02b593e2, separate from the command-doc
  and plugin-tool commits. `make check` green (924 plugin
  tests, coverage 94.62%/87.2%). Supersedes the repo-wide
  1.4.10 pin from review-council-multi-model-fanout.
  (Spec: openspec/changes/fix-triage-issue-dispatch-docs/,
  Fixes: #735)

### Fixed
- `uf setup` DevPod now falls back to binary download from GitHub
  Releases when Homebrew is absent, gated by `--yes` flag or
  interactive terminal (#268)
- `uf setup` auto mode now falls back to `dnf install`
  for Podman and curl installer for Ollama when Homebrew
  is absent on Fedora/RHEL (#268)
- Doctor hints now show `dnf install` commands on Fedora
  systems instead of `brew install`
- Fixed `genericInstallCmd("replicator")` returning
  `brew install` in the no-package-manager codepath
- `uf setup` auto mode now falls back to native package
  managers and `go install` when Homebrew is absent.
  Previously, 4 companion tools (Gaze, Replicator, Dewey,
  GitHub CLI) were skipped with download-link hints on
  Fedora/RHEL. Fallback chain: Homebrew -> dnf -> go install
  -> skip. Added `resolveMethod()` and `installViaGo()`
  helpers. `UF_PACKAGE_MANAGER=dnf` now forces the dnf
  install path. Fixes #214.
  (Spec: openspec/changes/setup-auto-native-fallback/)
- accept-variantless-parent-model: review-dispatch finalization
  (used by `/uf.review-pr`, `/uf.review-council`, and
  `/uf.unleash` code review) now accepts successful host and
  host-resolved advisor runs whose parent session model
  legitimately carries no variant (e.g.
  `enmaas/rits/zai-org/glm-5-3`); `validateRunModel` rejects
  such runs only when `resolved_parent_model` is null, matching
  the persistence layer's null-sanctioned
  `resolved_parent_variant`. Previously both fields were
  required, fail-closing real dispatches with "requires
  resolved parent model and variant" and an INCONCLUSIVE
  verdict (observed on unbound-force/dewey PR #135). Scaffold
  mirror synced byte-identical; finalize error-boundary tests
  cover null-model rejection and null-variant acceptance for
  both run sources. Documentation gate filed as in-repo
  issue #891 and unbound-force/website#360. (Spec:
  openspec/changes/accept-variantless-parent-model/)

## Recent Changes

- 036-subtool-error-reporting: Show actual error output from sub-tool failures in `uf init` and `uf setup`. Captures `CombinedOutput()` at all `ExecCmd` call sites, adds `TruncateOutput` helper (20-line threshold) in `internal/textutil`, decomposes `printSummary`/`printStepResult` to fix CRAP regressions. 22 new tests. Fixes #215.
- opsx/setup-sandbox-tools: Added Podman and DevPod to `uf setup` (13→16 steps) and `uf doctor`. Setup installs Podman via Homebrew with macOS Podman machine initialization (best-effort, with timeout via gtimeout/timeout fallback), installs DevPod via Homebrew, and configures the DevPod Podman provider alias (`devpod provider add docker --name podman -o DOCKER_PATH=podman`). Corrected DevPod provider option from `DOCKER_COMMAND` to `DOCKER_PATH` across setup, doctor, and spec artifacts. Doctor adds conditional DevPod check group (only when DevPod is detected): binary presence, version >= 0.5.0, podman provider registration, and `.devcontainer/devcontainer.json` existence. Podman doctor checks: version >= 4.3, runtime health via `podman info` (platform-aware), and Docker-to-Podman shim detection. Refactored setup step dispatch from 16 repetitive if/else blocks to data-driven `stepDef` slice with `executeSteps()` loop (CRAP 32 → 2). Added `hasProvider()` helper with exact first-column name matching for provider detection. Added `podmanMachineInit()` with `initMachineWithTimeout()` helper for macOS post-install (tries gtimeout, timeout, then no-timeout fallback). 6 new doctor tests, 12 new setup tests. Modified files: `internal/setup/setup.go`, `internal/setup/setup_test.go`, `internal/doctor/checks.go`, `internal/doctor/doctor.go`, `internal/doctor/doctor_test.go`, `internal/doctor/environ.go`. Website issue: unbound-force/website#121.
- opsx/sandbox-ide-flag: Added `--ide` flag to `uf sandbox create` and `uf sandbox start` for DevPod IDE selection (none, vscode, openvscode, fleet, jupyternotebook, cursor). Default `"none"` preserves backward compatibility. IDE resolution chain: CLI flag > `UF_SANDBOX_IDE` env var > `.uf/config.yaml` `sandbox.ide` > default. Added `validateIDE()` with strict allowlist validation. Added `waitForHealth()` with exponential backoff (500ms-5s, 60s timeout) to DevPod Start for health check polling after resume. Added `postStartCommand` to devcontainer.json template for OpenCode server auto-start. Fixed `Attach()` to detect persistent workspaces via `isPersistentWorkspace()` before ephemeral fallback. Fixed `Destroy()` to handle ephemeral cleanup directly when no persistent workspace exists. Added IDE field to unified config (`merge()`, `Defaults()`, `IsEmpty()`, `applyEnvOverrides()`). Fixed DevPod tunnel error suppression via status-based detection (D12). Added SSH fallback to start OpenCode serve when postStartCommand didn't run (D13). Replaced fmt.Fscanln with bufio.Scanner for destroy confirmation (D14). Added HealthCheckTimeout to Options for DI-based test injection. OS-aware devcontainer generation (D15): macOS uses `--userns=keep-id:uid=1000,gid=1000` for Podman VM UID mapping, Linux uses plain `--userns=keep-id` to avoid subuid range conflicts on container restart. `.devcontainer/` gitignored — generated per-user by `uf sandbox init`, no longer deployed by `uf init`. `json.NewEncoder` with `SetEscapeHTML(false)` preserves shell characters (`>`, `&`) in `postStartCommand`. Deleted orphaned root `.devcontainer.json`. 25+ new tests (including OS-aware devcontainerRunArgs, shell char preservation). Website issue: unbound-force/website#120. Modified files: `internal/sandbox/sandbox.go`, `internal/sandbox/devpod.go`, `internal/sandbox/config.go`, `internal/sandbox/workspace.go`, `internal/sandbox/sandbox_test.go`, `internal/config/config.go`, `internal/config/template.go`, `internal/scaffold/assets/devcontainer/devcontainer.json`, `internal/scaffold/scaffold.go`, `internal/scaffold/scaffold_test.go`, `.gitignore`, `cmd/unbound-force/sandbox.go`, `cmd/unbound-force/sandbox_test.go`, `cmd/unbound-force/main_test.go`.
- opsx/documentation-overhaul: Comprehensive documentation reorganization and expansion. Moved `USAGE.md` to `docs/usage.md` (expanded from ~60% to ~95% command coverage, added 18 missing slash commands to quick reference). Moved `unbound-force.md` to `docs/heroes.md` (renamed, fixed typos, added per-hero implemented-vs-planned capability markers). Created `docs/architecture.md` (ecosystem overview with ASCII diagrams, component relationships, specification pipelines, knowledge layer, convention packs, artifact flow, configuration). Created `docs/cli-reference.md` (full `uf` and `mutimind` command trees with flags and examples). Created `docs/configuration.md` (`.uf/config.yaml` guide with all 7 sections, hierarchy, and common scenarios). Fixed README.md factual errors (spec count 16 → 35, colon syntax `/opsx:propose` → `/opsx-propose`, expanded Repository Contents). Fixed QUICKSTART.md incomplete `git add`, added scaffolding summary, gateway/config/sandbox mentions. Fixed AGENTS.md project structure (added 5 missing internal packages, fixed spec count 001-018 → 001-035, added docs/ entry). Updated 3 agent files and 1 command file (plus 4 scaffold asset copies) to reference `docs/heroes.md` instead of `unbound-force.md`. No Go source changes. 6 task groups and 16 tasks completed.
- opsx/finale-pr-approval: Added user confirmation step to `/finale` Step 5 for PR title and body before creation. The confirmation mirrors Step 3's existing commit message approval pattern: show proposed content, ask "Approve, edit, or provide your own?" Added guardrail "NEVER create a PR without user approval of the title and body." Synced scaffold asset. No Go code changes. Closes #156.
- fix/finale-fork-detection: Enhanced `/finale` with fork detection. When a fork is detected via `gh repo view --json isFork,parent`, the user is asked whether to target the upstream parent or their fork's main before creating a PR. Website issue: unbound-force/website#117.
- fix/remove-mxf-setup-doctor: Removed dead `mxf` binary references from `uf setup` and `uf doctor`. Removed `installMxF()` function and setup step (14→13 steps, renumbered). Removed `mxf` from `coreTools` in doctor checks. Removed dead `mxf` switch cases from `homebrewInstallCmd` and `genericInstallCmd`. Cleaned 2 test functions and `mxf` from LookPath stubs in 5 other tests. Mx F hero availability check (agent file detection) is unaffected. Closes #160. PR #170.
- opsx/release-automation: Redesigned release pipeline from 10 manual steps to one-click workflow_dispatch trigger. Added preflight validation (semver format, tag uniqueness, semver ordering, CI status verification, unreleased commit detection) and automated tag creation. Added SBOM generation (Syft) and Cosign keyless signing to `.goreleaser.yaml`. Added Fedora packaging via Packit (`.packit.yaml`) targeting rawhide, f44, f43 with COPR builds on PRs and automated dist-git proposals on release. Created `unbound-force.spec` following Fedora Packaging Guidelines with `rpmlint`-clean validation. Created `docs/RELEASE_PROCESS.md` with trigger instructions, automation description, cadence, supply chain verification, rollback procedures, and Fedora packaging overview. Added `.fmf/version` for Testing Farm metadata. Updated AGENTS.md project structure and CI workflow table. Cleaned up orphaned draft releases (v0.13.0, v0.4.1). No Go code changes. Modified files: `.github/workflows/release.yml`, `.goreleaser.yaml`, `AGENTS.md`. New files: `.packit.yaml`, `unbound-force.spec`, `.fmf/version`, `docs/RELEASE_PROCESS.md`. 6 task groups and 22 tasks completed.
- fix/scaffold-mcp-skip: Fixed `uf init` ignoring `setup.tools.*.method: skip` in `.uf/config.yaml`. Tools configured as "skip" no longer get MCP entries added to `opencode.json` even when the binary is in PATH. Added 3 new tests for skip behavior.
- docs/commit-scope-rule: Added commit scope behavioral rule to AGENTS.md requiring explicit file staging on feature branches. Tooling scaffolds and unrelated fixes MUST go on separate branches.
- fix/scaffold-bridge-file-drift: Fixed bridge file staleness in scaffold engine. When a managed block is present but its content is stale, `updateBridgeFile` now replaces the block instead of skipping it. Updated CRAP baseline for scaffold refactoring. PR #168.
- fix/skill-to-skills-path: Normalized `skill/` to `skills/` in scaffold assets to match the canonical OpenCode directory convention. PR #167.
- opsx/commands-dir-migration: Migrated `.opencode/command/` to `.opencode/commands/` to align with OpenCode's canonical directory naming. Fixed printf `%s` pattern in hero contract script (SC2059). Suppressed errcheck lint for intentionally-ignored `os.Remove` calls. PR #164.
- opsx/agent-brief-consolidation: Consolidated AGENTS.md ownership into `/agent-brief` command and compressed AGENTS.md from ~983 to ~227 lines (~77% reduction). Externalized Recent Changes to CHANGELOG.md. Removed Active Technologies and Constitution summary sections (redundant with go.mod and constitution.md). Compressed 6 multi-paragraph governance subsections into 8 one-line RFC-style rules in Behavioral Rules. Removed `/uf-init` Step 9 (AGENTS.md behavioral guidance injection) -- governance blocks now embedded as verbatim templates in `/agent-brief`. Restructured `/agent-brief` from 11-section Tier taxonomy to flat 10-section model with conditional governance templates. Redirected 6 Divisor agents (Guard, Curator, Scribe, Herald, Envoy, Architect) from AGENTS.md Recent Changes to CHANGELOG.md. Updated `update-agent-context.sh` to target CHANGELOG.md and removed Active Technologies logic. Updated agent-file-template.md. Synced 8 scaffold assets. Updated `TestUFInitAsset_GuidanceBlockPresence` to `TestAgentBriefAsset_GovernanceBlockPresence`. No new Go dependencies. 7 task groups and 22 tasks completed.
- 035-openspec-parallel-markers: Added `[P]` parallel marker support to OpenSpec task template and schema instructions. Updated `openspec/schemas/unbound-force/templates/tasks.md` with `[P]` marker examples (sequential + parallel tasks in same group) and an HTML comment explaining the convention. Updated `openspec/schemas/unbound-force/schema.yaml` tasks instruction to guide the LLM on when to add `[P]` (different files, no dependencies) and when not to (same file). Added "Parallel Task Markers" subsection to AGENTS.md documenting the alignment between Speckit and OpenSpec task formats. No Go code changes. Closes #153.
- opsx/sandbox-uid-mapping: Fixed rootless Podman UID mismatch that broke `uf sandbox` direct mode and git operations on both macOS and Linux. Added `--userns=keep-id:uid=1000,gid=1000` to all `podman run` invocations via shared `uidMappingArgs()` helper (both `buildRunArgs()` and `buildPersistentRunArgs()`). Added macOS Podman machine UID probe using `busybox:latest` with `--entrypoint stat` override — fails with actionable error when virtiofs doesn't map UIDs correctly. Added `--uidmap` CLI flag as escape hatch for macOS users whose Podman machine can't be reconfigured (explicit `--uidmap`/`--gidmap` mapping). Added `Platform *PlatformConfig` injection field for cross-platform testability. Added `parsePodmanVersion()` enforcing Podman >= 4.3 at runtime. Added `isRootlessPodman()` guard rejecting `--uidmap` under rootful Podman (privilege escalation prevention). Added post-copy `chown -R dev:dev /workspace` in `PodmanBackend.Create()`. New `UIDMap` field in `Options`, `SandboxConfig`, and CLI flags. Added 21 new test functions, fixed 8 existing tests. Modified files: `internal/sandbox/config.go`, `internal/sandbox/podman.go`, `internal/sandbox/detect.go`, `internal/sandbox/sandbox.go`, `internal/sandbox/sandbox_test.go`, `internal/config/config.go`, `internal/config/config_test.go`, `cmd/unbound-force/sandbox.go`, `cmd/unbound-force/sandbox_test.go`. No new dependencies. 9 task groups and 50 tasks completed.
- opsx/gateway-token-refresh-fix: Fixed silent token expiry failures in the gateway's Vertex AI and Bedrock providers. Added token/credential expiry tracking (`tokenExpiry time.Time` on `VertexProvider`, `credExpiry time.Time` on `BedrockProvider`) with 55-minute/50-minute safety margins. When a background refresh fails, the stored token/credentials are now cleared atomically under the write lock, causing `PrepareRequest` to return a clear "Re-authenticate" error instead of silently forwarding stale credentials that Vertex rejects with `ACCESS_TOKEN_TYPE_UNSUPPORTED`. Added proactive refresh in `PrepareRequest` — when a token is within 5 minutes of expiry, a synchronous refresh is attempted with `sync.Mutex.TryLock()` deduplication to prevent thundering herd. Detached gateway child process now redirects stdout/stderr to `.uf/gateway.log` (0600 permissions) instead of discarding logs. `uf gateway status` displays the log file path. Named constants extracted (`vertexTokenLifetime`, `proactiveRefreshWindow`, `bedrockCredLifetime`). Added 14 new test functions including regression test (`TestVertexPrepareRequest_StaleTokenRegression` per TC-006), concurrent deduplication test with `sync.WaitGroup` barrier and `atomic.Int32` counter, and foreground negative case. Updated 8 existing test call sites. Synced 18 pre-existing scaffold asset drifts. Modified files: `internal/gateway/provider.go`, `internal/gateway/gateway.go`, `internal/gateway/gateway_test.go`. No new dependencies. 7 task groups and 42 tasks completed.
- opsx/review-pr-reliability: Fixed 6 reliability issues in `/review-pr` command that caused 12 wasted tool calls and an incomplete review during PR #139. Added argument-first parsing gate (prevents auto-detection when PR number provided), execution mode check in Step 0 (stops early when local tools can't run in plan/read-only mode), mandatory CI coverage matrix in Step 4 (makes CI→local tool dedup decision visible and auditable), save-and-navigate diff handling in Step 5 (replaces nonexistent `gh pr diff -- <path>` file-filter syntax with concrete temp-file technique), GitHub API guidance for PR branch file access (replaces failing `git show`/`git fetch` with `gh api`), and PR-introduced spec detection in Step 6 (finds specs in the PR's changed file list when not on base branch). Fixed pre-existing step numbering issue in Step 10. Modified files: `.opencode/command/review-pr.md`, `internal/scaffold/assets/opencode/command/review-pr.md`, `AGENTS.md`. No Go code changes. 8 task groups and 28 tasks completed.
- opsx/review-pr-enhancements: Enhanced `/review-pr` command with 5 improvements. Added PR walkthrough section (file-by-file change summary with path-based focus classification in Walkthrough table). Added GitHub suggestion blocks for one-click fixable code suggestions (triple-backtick `suggestion` syntax, restricted to literal replacements, no security control removal). Added built-in path-based review focus heuristics (7 categories: test-quality, cli-ux, security, documentation, ci-cd, dependencies, standard) applied as additive context during AI review. Added issue linking that parses `Fixes/Closes/Resolves #N` from PR body, fetches issue details (max 5, same-repo only, 2000-char truncation), extracts acceptance criteria, and validates coverage during alignment check. Added verdict-aligned review posting as single GitHub API call with correct APPROVE/REQUEST_CHANGES/COMMENT event type, graceful fallback to COMMENT on permission failures, explicit "approve" text required for APPROVE verdicts, AI-generated label in review body, and `change-verdict` option in confirmation prompt. Modified files: `.opencode/command/review-pr.md`, `internal/scaffold/assets/opencode/command/review-pr.md` (scaffold sync), `AGENTS.md` (PR Review section + Recent Changes). No Go code changes. All tasks completed.
- opsx/review-pr-command: Added `/review-pr` scaffold command for post-PR GitHub review. Single-agent, token-lean command adapted from org-infra's `review_pr.md` with standardization deltas: kebab-case naming, optional PR number with auto-detection, generic constitution reference, convention pack awareness with graceful degradation, severity pack integration, shell injection mitigations (`--body-file`, JSON input files, branch-name sanitization). Fetches CI check results with causality analysis (PR-caused vs pre-existing), runs local deterministic tools, performs AI review (alignment, security, constitution compliance), offers fix branches for pre-existing CI failures (with dirty-tree guard, user confirmation, collision handling), and offers in-line PR comments (15-comment cap, human confirmation gate). Requires `gh` CLI. New files: `.opencode/command/review-pr.md`, `internal/scaffold/assets/opencode/command/review-pr.md`. Modified files: `internal/scaffold/scaffold_test.go` (expectedAssetPaths 7→8 commands), `cmd/unbound-force/main_test.go` (file count 34→35), `AGENTS.md` (project structure + PR Review section). No Go logic changes. All tasks completed.
- opsx/agent-brief-command: Added `/agent-brief` slash command for AGENTS.md lifecycle management (create, audit, improve). The command auto-detects mode: creates AGENTS.md from project analysis when none exists (hybrid template + LLM approach filling Tier 1 sections from go.mod/package.json, Makefile, CI config, README), or audits existing files against a context-sensitive section taxonomy with scoring (Excellent/Strong/Adequate/Weak/Missing). Context-sensitive Tier 1C sections (Constitution, Spec Framework) are generated or checked only when `.specify/memory/constitution.md`, `specs/`, or `openspec/` are detected. Cross-framework governance bridge check verifies the constitution is stated as governing both Speckit and OpenSpec when both frameworks exist. Bridge file handling ensures CLAUDE.md (`@AGENTS.md` import) and .cursorrules exist. New "Agent Context" doctor check group replaces the single AGENTS.md existence check with 12 deterministic structural checks: file existence, 5 Tier 1 section headers (Project Overview, Build Commands, Project Structure, Code Conventions, Technology Stack), build section code blocks, line count (warn >300), constitution reference (context-sensitive), spec framework description (context-sensitive), CLAUDE.md bridge, .cursorrules bridge. InstallHint changed from "Run: uf init" to "Run: /agent-brief in OpenCode". Scaffold asset distributed via `uf init`. New file: `.opencode/command/agent-brief.md`. Modified files: `internal/doctor/checks.go`, `internal/doctor/doctor.go`, `internal/doctor/doctor_test.go`, `internal/scaffold/scaffold_test.go`. Added 12 new doctor test functions and updated 3 existing tests. All 7 task groups and 51 tasks completed.
- opsx/unified-config: Created `internal/config/` package with unified `Config` struct covering 7 sections (setup, scaffold, embedding, sandbox, gateway, doctor, workflow) and layered `Load()` function (user `~/.config/uf/config.yaml` > repo `.uf/config.yaml` > env vars). Added `uf config` command group with 3 subcommands: `init` (creates/updates `.uf/config.yaml` with commented template), `show` (displays merged config as YAML), `validate` (schema validation with actionable error messages). Added `github.com/goccy/go-yaml` dependency (replacing archived `gopkg.in/yaml.v3` for new code). Setup integration: config-driven `package_manager`, `skip` list, per-tool install methods, embedding model from config. Scaffold integration: removed `workflowConfigContent` constant, config-based language selection via `scaffold.language` field. Gateway integration: config-driven `port` and `provider` defaults. Agent model cleanup: removed hardcoded `model` frontmatter from 24 agent files (12 live + 12 scaffold copies). Absorbs `.uf/sandbox.yaml` into unified config with backward-compatible fallback. Go 1.24+ (CLI), `github.com/goccy/go-yaml` (YAML parsing). All tasks completed.
- opsx/gateway-global-region-error: Replaced silent `"global"` → `"us-east5"` region fallback in `newVertexProvider()` with a clear, actionable error. When `VERTEX_LOCATION` or `CLOUD_ML_REGION` resolves to `"global"`, the gateway now returns an error at startup explaining that Vertex AI `rawPredict`/`streamRawPredict` requires a specific regional endpoint and recommending `ANTHROPIC_VERTEX_REGION` as the override. Changed `newVertexProvider()` signature from `*VertexProvider` to `(*VertexProvider, error)`, propagated through `DetectProvider()` and `NewProviderByName()`. Preserved `"us-east5"` default for the empty-region case. Added 6 new test functions. Updated 3 existing test call sites. Modified files: `internal/gateway/provider.go`, `internal/gateway/gateway_test.go`. No new dependencies. 16 tasks completed.
- 034-gateway-vertex-translation: Extended `uf gateway` Vertex provider with full request/response translation. Request body transformation: removes `model` field (Vertex uses URL path), injects `anthropic_version: "vertex-2023-10-16"` if absent, preserves existing `anthropic_version`, updates `Content-Length` after body modification. Streaming endpoint detection: selects `streamRawPredict` when `"stream": true`, `rawPredict` otherwise (`count_tokens` always uses `rawPredict`). Header stripping: removes `anthropic-beta` and `anthropic-version` HTTP headers before forwarding to Vertex. SSE response filtering: created `internal/gateway/sse.go` with `sseFilterReader` that drops `vertex_event` and `ping` SSE events from Vertex streaming responses. Synthetic model catalog at `/v1/models`. New file: `internal/gateway/sse.go`. Modified files: `internal/gateway/provider.go`, `internal/gateway/gateway.go`, `internal/gateway/gateway_test.go`. All stdlib. All 5 user stories and 65 tasks completed.
- opsx/quickstart-usage-docs: Created QUICKSTART.md and USAGE.md onboarding documentation. No Go code changes -- documentation only. 29 tasks completed.
- 033-gateway-command: Added `uf gateway` command with 3 subcommands (`start`, `stop`, `status`) for LLM reverse proxy gateway. Created new `internal/gateway/` package. Three provider implementations: AnthropicProvider, VertexProvider, BedrockProvider. All 5 user stories and 87 tasks completed.
- 031-unleash-openspec: Extended `/unleash` command to support OpenSpec (`opsx/*`) branches alongside existing Speckit (`NNN-*`) branches. No Go code changes. All 4 user stories and 17 tasks completed.
- 029-sandbox-cde-lifecycle: Extended `internal/sandbox/` with persistent workspace lifecycle and CDE backend. All 4 user stories and 111 tasks completed.
- opsx/workflow-phase-boundaries: Added workflow phase boundary enforcement. 27 tasks completed.
- 028-sandbox-command: Added `uf sandbox` command with 5 subcommands. All 4 user stories and 87 tasks completed.
- 027-externalize-tool-init: Externalized Speckit, OpenSpec, and Gaze initialization from embedded scaffold assets to CLI delegation. All 4 user stories and 41 tasks completed.
- 026-documentation-curation: Added The Curator Divisor agent. All 4 user stories and 38 tasks completed.
- opsx/gitignore-init: Added `ensureGitignore()` to scaffold engine. 9 tasks completed.
- 025-uf-directory-convention: Unified all tool workspace directories under `.uf/`. All 5 user stories and 79 tasks completed.
- opsx/gatekeeping-protection: Added gatekeeping value protection rules. 13 tasks completed.
- opsx/divisor-content-agents: Added 3 content-creation Divisor agents (Scribe, Herald, Envoy). 18 tasks completed.
- 024-replicator-migration: Replaced Node.js Swarm plugin with Replicator Go binary. All 4 user stories and 63 tasks completed.
- 023-doctor-setup-dewey: Updated `uf doctor` and `uf setup` for Dewey unified memory. All 3 user stories and 38 tasks completed.
- 022-hivemind-dewey-migration: Migrated all Hivemind tool references to Dewey equivalents. All 5 user stories and 29 tasks completed.
- 020-dewey-knowledge-retrieval: Added Dewey knowledge retrieval behavioral instructions across the agent ecosystem. All 5 user stories and 22 tasks completed.
- 019-divisor-council-refinement: Refined the Divisor Council review system. All 6 user stories and 56 tasks completed.
- 018-unleash-command: Created `/unleash` autonomous Speckit pipeline command. All 6 user stories and 31 tasks completed.
- 017-init-opencode-config: Moved `opencode.json` management from `uf setup` to `uf init`. All 4 user stories and 37 tasks completed.
- 016-autonomous-define: Enabled autonomous define stage. All 5 user stories and 28 tasks completed.
- setup-init-full-stack: Extended `uf setup` to install ALL ecosystem tools. 43 tasks completed.
- 015-dewey-integration: Integrated Dewey as the semantic knowledge layer. All 5 user stories and 40 tasks completed.
- 013-binary-rename: Renamed CLI binary from `unbound` to `unbound-force` with `uf` alias. All 5 user stories and 35 tasks completed.
- 012-swarm-delegation: Added swarm delegation workflow. All 4 user stories and 42 tasks completed.
- 011-doctor-setup: Implemented Doctor and Setup commands. All 5 user stories and 79 tasks completed.
- 009-shared-data-model: Implemented shared data model. All 5 user stories and 20 tasks completed.
- 008-swarm-orchestration: Implemented swarm orchestration engine. All 5 user stories and 31 tasks completed.
- 007-mx-f-architecture: Implemented Mx F Manager hero. All 6 user stories and 68 tasks completed.
- 006-cobalt-crush-architecture: Implemented Cobalt-Crush Developer persona.
- 005-the-divisor-architecture: Implemented The Divisor PR Reviewer Council. All 5 user stories and 44 tasks completed.
- graphthulhu-homebrew-distribution: Added graphthulhu as Homebrew cask dependency.
- 004-muti-mind-architecture: Implemented Muti-Mind backend CLI and OpenCode agent.
- 003-specification-framework: Implemented unified two-tier specification framework. All 8 user stories and 67 tasks completed.
- 010-knowledge-graph-integration: Implemented knowledge graph integration. All 5 user stories verified.
- 002-hero-interface-contract: Completed spec implementation. All FRs and SCs validated.
- 001-org-constitution: Completed spec implementation -- constitution ratified v1.0.0.
- 001-org-constitution: Ratified org constitution v1.0.0 with three principles.
- Specs 001-009: Added architectural design specs for all heroes and cross-cutting concerns.
