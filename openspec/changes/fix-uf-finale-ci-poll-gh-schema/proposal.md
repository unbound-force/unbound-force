<!-- `originating_issue: 895` is set in .openspec.yaml so
     /uf.finale emits `Closes #895` in the PR body. -->

## Why

`/uf.finale` Step 7 (Watch CI Checks) cannot complete through the
plugin tool. `uf_finale_ci_poll` fails immediately with
`gh CLI error: Unknown JSON field: "conclusion"` (code
`GH_CLI_ERROR`) against the installed gh CLI 2.91.0 (2026-04-22),
forcing the finale to fall back to manual CI polling (issue
#895).

The tool lives in the `uf-finale` plugin
(`.opencode/plugins/uf-finale/index.ts`, `createUfFinaleCiPollTool`
— the issue attributes it to the uf-workflow plugin, but that
plugin registers none of the `uf_finale_*` tools). Its polling
loop invokes:

    gh pr checks <N> --json name,state,conclusion,detailsUrl

gh 2.91.0's `gh pr checks --json` supports only `bucket`,
`completedAt`, `description`, `event`, `link`, `name`, `startedAt`,
`state`, and `workflow`. The requested `conclusion` and
`detailsUrl` fields do not exist; gh validates the requested field
list and exits non-zero before emitting any data. The tool's
expected JSON schema drifted from the installed gh CLI's actual
output schema.

The drift has a second, latent layer. The tool runs `gh pr
checks` through throwing `execSync` and assumes exit 0 with `[]`
on stdout for a checkless PR. Real gh behavior (verified against
2.91.0 by the `fetch_pr_checks` work, change
`extend-fetch-pr-review-context`, issue #861):

- Pending checks: exit 8, authoritative JSON array still on
  stdout.
- Failing checks: exit 1, JSON array still on stdout.
- Checkless PR: exit 1, EMPTY stdout, "no checks reported" on
  stderr — gh never emits `[]`.

So even with corrected field names, `execSync` would throw on the
poll loop's first iteration against any pending or failing PR,
and the zero-checks branches would be unreachable. The fix must
align the field list AND tolerate gh's exit-code semantics.

The same drift class exists in the review-dispatch plugin:
`fetch_pr_review_context`'s advisory `ci_checks` sub-call requests
`--json name,state,conclusion` (review-dispatch/index.ts, section
1.2c). It degrades non-fatally — every context fetch now carries a
`ci_checks degraded: Unknown JSON field: "conclusion"` warning and
returns empty `ci_checks`, stripping CI signal from `/uf.review-pr`
and `/uf.review-council` prompt contexts.

The sub-call carries a second layer of the same bug: it runs on
the throwing `runGhCommand`, so even after the field list is
corrected, pending (exit 8) and failing (exit 1) PRs still throw
and degrade with the authoritative array on stdout. This change
moves the sub-call onto the exit-code-tolerant runner as well, so
`ci_checks` populates for every check state, not only green PRs.

The `uf_finale_ci_poll` output contract itself is already
governed by the unarchived `finale-token-reduction` change's
`plugin-tools` delta (`uf_finale_ci_poll tool`). This change
refines only the gh wire schema and error classification and
preserves that contract unchanged; the `uf-finale-ci-poll`
delta records the archive ordering (after
`finale-token-reduction`).

In-repo prior art already solved this correctly: `fetch_pr_checks`
runs `gh pr checks <N> --json name,state,bucket,link` through an
exit-code-tolerant `Bun.spawn` runner, treats stdout that parses
as a JSON array as data on ANY exit code, recognizes the
"no checks reported" stderr marker for checkless PRs, and derives
classification from gh's authoritative `bucket` field with
fail-closed handling of unknown vocabulary.

## What Changes

1. **Wire schema alignment.** `uf_finale_ci_poll` requests `gh pr
   checks <N> --json name,state,bucket,link` — the gh 2.91.0
   verified schema — replacing the nonexistent
   `conclusion,detailsUrl` fields. `failedChecks` entries keep
   their `{name, conclusion, url}` shape with `conclusion` sourced
   from the check's `state` and `url` from `link`, so the tool's
   output contract is unchanged.

2. **Exit-code-tolerant checks invocation.** The polling loop's
   `gh pr checks` call moves from throwing `execSync` to an
   argument-array `Bun.spawn` runner that resolves
   `{stdout, stderr, exitCode}` without throwing (mirroring
   review-dispatch's `runGhCommandTolerant`, design D4). Stdout
   that parses as a JSON array is check-status data on any exit
   code; stdout that does not parse as an array is a fetch
   failure.

3. **Checkless-PR detection.** The zero-checks branches
   (`no-checks-workflows-exist` / `no-checks-no-workflows`) are
   reached through the verified "no checks reported" stderr marker
   (exit 1, empty stdout) in addition to a parsed empty array,
   replacing the unreachable `[]`-on-stdout assumption.

4. **Bucket-based classification.** Poll-completion and pass/fail
   derivation keys off gh's authoritative `bucket` field:
   `pending` keeps polling; `fail` and `cancel` populate
   `failedChecks` (preserving today's CANCELLED-counts-as-failed
   behavior); `pass` and `skipping` pass; an unknown bucket value
   classifies fail-closed as failed with the value surfaced, so
   vocabulary drift is visible instead of burning the poll
   timeout.

5. **PR-not-found classification.** gh's "Could not resolve to a
   PullRequest" stderr (prior-art pattern) returns a structured
   non-retryable error instead of a confusing JSON parse failure.

6. **Secondary: ci_checks sub-call alignment.**
   `fetch_pr_review_context`'s `ci_checks` sub-call switches to
   `name,state,bucket` and moves onto the exit-code-tolerant
   runner, mapping into its existing internal
   `{name, status, conclusion}` shape (conclusion sourced from
   `bucket`). Pending (exit 8) and failing (exit 1) PRs now
   populate `ci_checks` instead of degrading, and `ci_pending`
   is derived from the bucket vocabulary rather than the
   lowercase `"completed"` comparison — eliminating the
   degradation warning and the spurious in-progress warning
   and restoring CI signal to review contexts.

7. **No dual-schema parser.** The fix aligns to the single
   verified schema with fail-closed unknown-vocabulary handling
   rather than probing gh versions or accepting legacy field sets;
   the rejected alternative and its rationale are recorded in
   design.md.

8. **Regression tests.** Tests exercise exit 8 with JSON stdout
   (poll continues), exit 1 with JSON stdout (fail reported), exit
   1 with empty stdout + "no checks reported" (zero-checks
   branches), not-found stderr, and unknown-bucket classification
   — scenarios that fail against the current implementation. The
   golden-output-shape command pin is updated to the new field
   list.

## Capabilities

### New Capabilities
- `uf-finale-ci-poll`: `uf_finale_ci_poll` gh wire schema,
  exit-code-tolerant invocation, checkless-PR and not-found
  detection, bucket-based classification, and preservation
  of the output contract governed by
  `finale-token-reduction`'s `plugin-tools` delta.

### Modified Capabilities
- `fetch-pr-review-context`: the advisory `ci_checks`
  sub-call's gh field list, exit-code-tolerant invocation,
  conclusion sourcing, and `ci_pending` derivation, removing
  the constant degradation warning.

### Removed Capabilities
- None.

## Impact

- **Code**: `.opencode/plugins/uf-finale/index.ts`
  (`createUfFinaleCiPollTool`, module-local tolerant gh runner)
  and `.opencode/plugins/review-dispatch/index.ts` (section 1.2c
  ci_checks sub-call), each synced byte-identical with its
  scaffold twin under
  `internal/scaffold/assets/opencode/plugins/` per the
  scaffold-parity obligation. No Go code changes.
- **Tests** (canonical only, not scaffolded):
  `.opencode/test/uf-finale.test.ts` (ci_poll describe block
  migrates to spawn-mock fixtures covering gh exit codes and
  stderr markers), `.opencode/test/golden-output-shape.test.ts`
  (command-string pin),
  `.opencode/test/fetch-pr-review-context.test.ts`
  (ci_checks fixtures to verified gh shapes), via `make
  plugin-test` under the 94 percent statements coverage floor.
- **Docs**: restored `/uf.finale` Step 7 CI-watch and restored
  `ci_checks` signal are user-visible, so the documentation
  gate filed both a `docs`-labeled in-repo issue
  (unbound-force/unbound-force#925) and a `docs`-labeled
  unbound-force/website issue (unbound-force/website#372);
  both numbers thread into the CHANGELOG trailing reference.
  Content pipeline skipped: a restorative change measured
  against the pre-break baseline is not a new blog/tutorial
  capability.
- **Consumers**: `/uf.finale` Step 7 flow, its mandatory gates,
  and the documented tool contract in `uf.finale.md` are unchanged
  — only the gh wire protocol and error classification improve.
  No command-doc edits required.
- **Requirement composition**: a MODIFIED FR-006 delta restates
  the requirement in full, so this change composes the
  unarchived sibling review-dispatch-diff-manifest's
  `exceeds_threshold` bullet (RD-FR-032) and MUST archive after
  that sibling. All other existing requirements are unchanged.
- **Verified not affected**:
  `council-review-action/scripts/prefetch.sh`
  (`name,state,description` — all valid gh 2.91.0 fields),
  `uf.review-pr.md`'s documented `name,state,description,link`
  invocation, the host-side `fetch_pr_checks` tool (already
  bucket-native), and the scaffold command contract test (already
  pins `name,state,bucket,link`).
- **No gate changes**: coverage floors, severity definitions, and
  CI flags are untouched.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

The tool's self-describing JSON output contract is preserved
byte-for-byte in shape, so downstream consumers (the `/uf.finale`
Step 7 flow, golden-shape tests) need no changes. The fix restores
the finale's ability to complete CI-watch autonomously — the
current breakage forces manual polling, an autonomy regression
this change eliminates.

### II. Composability First

**Assessment**: PASS

The uf-finale plugin remains independently usable; no new
cross-plugin or runtime dependency is introduced (`Bun.spawn` is
the existing plugin runtime, already used by review-dispatch). The
review-dispatch fix is internal to its own sub-call and adds no
coupling.

### III. Observable Quality

**Assessment**: PASS

Success and failure envelopes remain machine-parseable with
structured codes. The fix makes drift observable: unknown bucket
values surface verbatim in `failedChecks` instead of silently
timing out, and PR-not-found returns a distinct structured error
rather than a JSON parse message. gh's authoritative bucket
vocabulary is echoed, not paraphrased.

### IV. Testability

**Assessment**: PASS

Every new behavior is unit-testable against mocked gh process
output — exit codes, stdout payloads, and stderr markers — with no
network, gh installation, or live PR required. Regression tests
fail against the current implementation (the bug is reproduced in
fixtures), and the 94 percent statements coverage floor is
enforced by `make plugin-test`.

### V. Security by Default

**Assessment**: PASS

The checks invocation moves from shell string interpolation to
`Bun.spawn` argument arrays, removing the shell layer entirely —
previously safe only because `prNumber` is validated to an
integer, now structurally injection-safe. No new dependencies, no
new external inputs beyond the already-validated `prNumber`.

### VI. Deterministic Execution

**Assessment**: PASS

The CI-watch procedure stays a tested executable tool rather than
degrading to prose-driven manual polling (today's workaround).
Field lists, exit-code semantics, and stderr markers are pinned in
tests against the verified gh 2.91.0 schema, so future drift fails
loudly in the suite instead of silently in production.
