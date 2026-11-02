## Context

`uf_finale_ci_poll` (uf-finale plugin,
`.opencode/plugins/uf-finale/index.ts`, `createUfFinaleCiPollTool`)
cannot poll CI against the installed gh CLI 2.91.0 (issue #895).
Its polling loop requests `gh pr checks <N> --json
name,state,conclusion,detailsUrl`; gh 2.91.0's `gh pr checks
--json` field vocabulary is `bucket, completedAt, description,
event, link, name, startedAt, state, workflow`, so gh validates
the field list, rejects `conclusion`, and exits before emitting
data. The throwing `execSync` call surfaces this as `GH_CLI_ERROR`.

Two further latent mismatches would keep the tool broken even
after correcting the field list: gh exits 8 for pending and 1 for
failing checks (JSON array still on stdout), and exits 1 with
empty stdout plus a "no checks reported" stderr marker for a
checkless PR — it never emits `[]`. `execSync` throws on all of
these, so the poll loop cannot survive its first iteration on a
non-green PR and the zero-checks branches are unreachable.

The same drift degrades `fetch_pr_review_context`'s advisory
`ci_checks` sub-call (review-dispatch plugin, section 1.2c),
which requests `name,state,conclusion` and now always fails into
its non-fatal degradation path. The sub-call also runs on the
throwing `runGhCommand`, so correcting only the field list would
still degrade on pending (exit 8) and failing (exit 1) PRs —
where the authoritative array is on stdout. The fix therefore
applies the tolerant runner to the sub-call too.

The `uf_finale_ci_poll` output contract is already governed by
the unarchived `finale-token-reduction` change's `plugin-tools`
delta (`uf_finale_ci_poll tool` requirement: poll cadence,
`timeoutMinutes` default, status vocabulary, `failedChecks`
shape, and the `CI_POLL_TIMEOUT` / `GH_CLI_ERROR` envelopes).
This change refines only the gh wire schema and error
classification; the new `uf-finale-ci-poll` delta preserves
that contract (FCP-FR-004) and records the archive ordering.

In-repo prior art: the `fetch_pr_checks` tool (change
`extend-fetch-pr-review-context`, issue #861) verified the gh
2.91.0 wire schema (`name,state,bucket,link`; bucket values
`pass`, `fail`, `pending`, `skipping`, `cancel`; states uppercase
like `SUCCESS`, `FAILURE`, `PENDING`, `SKIPPED`, `CANCELLED`) and
built the exit-code-tolerant `Bun.spawn` invocation pattern this
design reuses. This change is a bugfix scoped to restore
documented behavior; it is tactical (single repo, under three
stories), which is why it runs as OpenSpec.

## Goals / Non-Goals

### Goals

- `uf_finale_ci_poll` works against gh 2.91.0 for every PR
  state: green, failing, pending, cancelled-skewed, checkless
  (with and without workflows), conflicting, and unresolvable.
- The tool's output contract stays byte-compatible in shape —
  envelope keys, `failedChecks` entry keys, status vocabulary —
  so `/uf.finale` Step 7, its mandatory gates, and
  `uf.finale.md` require no changes.
- Classification keys off gh's authoritative `bucket` field with
  fail-closed handling of unknown vocabulary.
- `fetch_pr_review_context`'s `ci_checks` sub-call stops
  degrading for every check state — green, pending (exit 8),
  and failing (exit 1) — and restores CI signal to review
  contexts, with `ci_pending` derived from the bucket
  vocabulary.
- Regression tests pin the verified schema, exit codes, and
  stderr markers, failing loudly on future drift.
- Scaffold twins under `internal/scaffold/assets/opencode/`
  stay byte-identical to the edited plugin sources.

### Non-Goals

- No gh version detection, probing, or dual-schema parser (see
  D1).
- No refactor of the plugin's other `execSync` call sites
  (`gh pr view`, git operations) — only the checks invocation
  needs exit-code tolerance.
- No extraction of the tolerant runner into `.opencode/lib/` and
  no refactor of review-dispatch's `runGhCommandTolerant` (see
  D2).
- No changes to `fetch_pr_checks` (already correct), to command
  docs, or to any quality/governance gate including the 94
  percent statements coverage floor.
- No changes to Go code beyond keeping the scaffold asset twins
  in sync.

## Decisions

### D1: Single verified schema; no dual parser

The fix requests exactly `name,state,bucket,link` — verified
against the installed gh 2.91.0 — and handles unknown bucket
vocabulary fail-closed (FCP-FR-003). The issue's alternative (accept
either field set with a deprecation path) is rejected:

- The legacy `conclusion`/`detailsUrl` expectations cannot be
  verified against ANY gh version available to us; the old test
  fixtures (`state: "COMPLETED", conclusion: "SUCCESS"`) resemble
  `gh pr view`'s CheckRun shape, suggesting the original field
  list was conflated from a different command, not drifted from
  a real `gh pr checks` schema.
- A dual-schema parser needs a fallback round trip (gh rejects
  unknown fields at argument validation, before emitting data),
  doubles the test matrix, and still cannot classify statuses it
  does not recognize.
- Fail-closed unknown-vocabulary handling is the stronger
  deprecation path: future gh drift surfaces immediately as a
  visible failure instead of a silent misparse.

Constitution: Principle VI (Deterministic Execution) — one
verified schema pinned by tests beats runtime schema negotiation.

### D2: Exit-code-tolerant `Bun.spawn` runner, module-local

The checks invocation moves to a module-local async runner in
`uf-finale/index.ts` that spawns `gh` with an argument array and
resolves `{stdout, stderr, exitCode}` without throwing —
mirroring review-dispatch's `runGhCommandTolerant` (design D4 of
`extend-fetch-pr-review-context`), including `AbortSignal`
timeouts. The per-invocation timeout is normative, not merely an
implementation detail: FCP-FR-001 pins a bound of at most 30
seconds (`AbortSignal.timeout(30_000)`, the review-dispatch
`GH_CLI_TIMEOUT_MS` value) and FCP-FR-005 carries a
hung-invocation regression test. Parsing follows the same
rule: on any exit code, stdout that parses as a JSON array is
data; anything else is a fetch failure.

Alternatives rejected:

- `execSync`/`execFileSync` with try/catch on the thrown error's
  `stdout`: keeps the shell string-interpolation layer, which is
  safe today only because `prNumber` is validated to an integer.
  Argument-array spawn removes the shell structurally
  (constitution Principle V, Security by Default).
- Extracting a shared runner into `.opencode/lib/`: this would be
  the second consumer; the DRY threshold (three duplications) is
  not met, and refactoring review-dispatch's runner is out of
  scope for a bugfix. Both runners carry pointer comments naming
  each other so a third consumer triggers extraction.

The `gh pr view` conflict pre-check stays on `execSync`: it exits
0 with JSON on success and its failure is already caught and
ignored by design. Minimal diff.

### D3: Poller-semantics classification from `bucket`

While any check buckets as `pending`, keep polling (gh's bucket
subsumes the old `PENDING`/`QUEUED`/`IN_PROGRESS` state matching
— gh owns that mapping, so new non-terminal states need no code
change). Terminal classification:

- `fail`, `cancel` → `failedChecks`. Cancel-as-failed preserves
  the pre-fix behavior where a CANCELLED conclusion counted as
  failed; a poller must treat cancelled as terminal.
- `pass`, `skipping` → passing (skipping preserves the old
  SKIPPED exclusion).
- Unknown bucket → fail-closed as failed, with the check's state
  value surfaced in the `failedChecks` entry.

The unknown-bucket choice deliberately diverges from
`fetch_pr_checks`, which classifies unknown as `pending`. The
tools have different jobs: `fetch_pr_checks` gates repair
eligibility, where "do not proceed on an unrecognized status" is
the safe answer; `uf_finale_ci_poll` is a poller, where an
unrecognized classification will not change by waiting — spinning
to the timeout wastes up to 15 minutes and reports only
`CI_POLL_TIMEOUT`. Surfacing the check as failed makes drift
visible immediately, and `/uf.finale` Step 7's fail gate already
routes the agent to investigate. Constitution: Principle III
(Observable Quality) — drift is observable, not swallowed.

`state` never controls flow anymore; it only flows into
`failedChecks[].conclusion`, preserving the old value vocabulary
(`FAILURE`, `CANCELLED`) so the printed gate output in Step 7
reads the same as before.

### D4: Output contract frozen

Envelope keys (`status`, `failedChecks`, `mergeable`,
`workflowsConfigured`, `skipReason`), status vocabulary, and
`failedChecks` entry keys (`{name, conclusion, url}`) are
unchanged; `conclusion` is sourced from `state`, `url` from
`link`. This keeps `uf.finale.md` (a scaffolded command asset
with drift tests) untouched and limits the golden-output-shape
update to the pinned gh command string. Zero-checks and
not-found detections reuse the existing branch structures.

### D5: Zero-checks and not-found via verified stderr markers

A checkless PR is detected by gh's verified shape (exit 1, empty
stdout, /no checks reported/i on stderr) and routed to the
existing `workflowsConfigured` branches; a parsed empty array is
still honored should a future gh emit `[]`. An unresolvable PR is
detected by the prior-art stderr pattern
(/not found|no pull request|could not find|could not resolve to a
pullrequest/i) and returns a structured non-retryable failure
naming the PR. Both markers carry in-code comments citing the gh
version and verification date, matching the review-dispatch
convention.

### D6: ci_checks sub-call tolerant runner and ci_pending

`fetch_pr_review_context`'s sub-call switches to
`--json name,state,bucket` and moves from the throwing
`runGhCommand` to the exit-code-tolerant `runGhCommandTolerant`
(the same prior-art path `fetch_pr_checks` uses). On any exit
code, stdout that parses as a JSON array is check-status data;
empty or unparseable stdout without a recognized marker is a
sub-call failure that still degrades non-fatally. This makes
pending (exit 8) and failing (exit 1) PRs populate `ci_checks`
instead of degrading.

The sub-call maps `status ← state`, `conclusion ← bucket`,
preserving the internal `{name, status, conclusion}` type and
every consumer (prompt building, CiChecks interface). Because
`status` now carries gh's uppercase state vocabulary, the
consumer's `ci_pending` computation MUST be reconciled: it is
derived from the bucket vocabulary (`true` when any check
buckets as `pending`, else `false`) rather than comparing
`status` to the lowercase `"completed"` value gh never emits.
This is a normative correction to the origin capability's
FR-006, so it is recorded as a MODIFIED requirement (with a
Previously paragraph) rather than left as internal detail.
Because a MODIFIED requirement is restated in full, the
restatement composes the unarchived sibling change
review-dispatch-diff-manifest's `exceeds_threshold` bullet
(RD-FR-032: shared line-counting function, iff-manifest-
delivery) instead of reverting it to origin-era wording; this
change MUST archive after that sibling. The published
consumer-facing context shape is unchanged; only the
derivation changes. Test fixtures move to the verified gh
shapes and assert `ci_pending` for green, pending, failing,
and unknown-bucket PRs.

Unknown-bucket handling now has three consumers with three
deliberately different, individually documented semantics:
`uf_finale_ci_poll` fails closed to a failed status so drift
surfaces immediately (D3); `fetch_pr_checks` classifies unknown
as `pending` (fail-closed repair eligibility) and warns; this
advisory sub-call leaves `ci_pending` false and surfaces the
unknown value verbatim in `conclusion`. The three choices share
one principle: the conservative answer for the consumer's job,
not a shared vocabulary rule. A fourth consumer MUST choose
explicitly and document its choice.

### D7: Test strategy

- `uf-finale.test.ts`'s `uf_finale_ci_poll` block migrates its
  checks-command mocking to the Bun.spawn mock harness
  (`.opencode/test/helpers/gh-spawn-mock.ts`, proven by
  `fetch-pr-checks.test.ts`), adapting the harness where needed
  for direct factory invocation. Fixtures reuse the verified
  shapes: exit 8 + pending array, exit 1 + failing array, exit 1
  + empty stdout + no-checks stderr, not-found stderr, cancel,
  skipping, and unknown-bucket (`state: "NEUTRAL"`,
  `bucket: "blocked"`). A hung-invocation fixture exercises the
  per-invocation timeout.
- `golden-output-shape.test.ts`'s command pin becomes
  `gh pr checks 43 --json name,state,bucket,link` — the primary
  regression pin (fails against the pre-fix code).
- `fetch-pr-review-context.test.ts` fixtures align to
  `name,state,bucket`, add pending (exit 8) and failing
  (exit 1) sub-call scenarios, and assert `ci_pending` is
  `false` for a green PR and `true` for a pending PR.
- Coverage: the 94 percent statements floor is enforced by
  `make plugin-test`; no threshold changes (gatekeeping rule).
  Constitution: Principle IV (Testability) — all scenarios are
  process-mocked; no gh binary, network, or live PR is required.

## Risks / Trade-offs

- **R1 — gh schema drifts again.** A future gh changing the
  field list fails loudly: unknown fields abort at argument
  validation into `GH_CLI_ERROR`, and unknown bucket values
  surface as failed checks (D3). The golden pin and vocabulary
  tests make drift a suite failure, not a silent misparse.
  Accepted.
- **R2 — Scaffold parity.** Both edited plugin sources have
  byte-identical twins under
  `internal/scaffold/assets/opencode/plugins/`; drift tests
  enforce sync. Tasks include an explicit sync-and-verify step.
- **R3 — Unknown-bucket fail-fast changes wait behavior.** A
  hypothetical benign unknown bucket now reports `fail` instead
  of waiting out the timeout. Documented divergence (D3); the
  Step 7 fail gate offers investigation options, and the
  surfaced state value makes the anomaly diagnosable. Accepted.
- **R4 — Duplicated tolerant runner.** A second ~15-line
  module-local runner exists (D2). Pointers comment both sites;
  extraction is triggered by a third consumer. Accepted for a
  bugfix-scoped change.
- **R5 — `state` flows into `failedChecks[].conclusion`
  unvalidated.** State strings come from gh, not user input, and
  are informational only (never control flow — D3), so no
  validation beyond array-membership parsing is warranted. Note:
  gh check names and links already flow through output today;
  the argument-array spawn (D2) removes the only injection
  surface on the input side.
- **R6 — cancel-as-failed differs from fetch_pr_checks'
  cancel-as-pending.** Intentional: eligibility vs. polling
  semantics (D3). Both behaviors are spec-pinned in their
  respective requirements so the difference is auditable.
- **R7 — Overlapping governance of one tool.** The unarchived
  `finale-token-reduction` change's plugin-tools delta already
  governs `uf_finale_ci_poll`. This change refines only the gh
  wire schema and classification and preserves that contract
  (FCP-FR-004); the new delta records the archive ordering
  (after `finale-token-reduction`) and cross-references the
  sibling requirement. Sibling change history is not rewritten.
