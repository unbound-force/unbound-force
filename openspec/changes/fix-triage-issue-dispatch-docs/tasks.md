# Tasks

## 1. Command edits (canonical)

- [x] 1.1 Edit `.opencode/commands/uf.triage-issue.md` Section 2.3:
  replace the "returns status, counters, consolidated findings,
  proposals, and warnings" claim with the counts-only summary
  contract; add the sanctioned retrieval step (one
  `consolidate_dispatch` call without provenance immediately after
  `dispatch_review_runs`, returning runs with `workflow_verdict`,
  findings, advisories, run counts, and top-level `proposals`);
  document the failure behavior — transient failure of the
  side-effect-free retrieval call retries exactly once then fails
  closed; permanent failure fails closed without retry; an all-skip
  dispatch (zero completed, zero failed runs, no dispatch session)
  is recorded as a valid outcome per the no-vote semantics, not
  treated as a retrieval failure
- [x] 1.2 Edit Section 2.4: source the proposals array from the
  no-provenance `consolidate_dispatch` payload; keep per-proposal
  `dewey_store_learning` handling using the proposal's
  `information`, `tag`, and `category` (the validated proposal
  shape is `{agent, information, tag, category}` as defined by the
  review-dispatch dispatch artifact schema — there is no
  `reference` field) and the Dewey-unavailable informational skip
  unchanged
- [x] 1.3 Edit Section 3.1 (Phase 3): name authoritative sources —
  run-recorded `workflow_verdict` from the payload's `runs` governs
  verdict resolution (child prose does not; on disagreement
  run-recorded wins); category/objectivity/reasoning/split come
  from a single `read_assessment_fields` call with the dispatch's
  run identifiers (the tool owns the persisted-text reads and the
  untrusted-input validation per FR-003 — the command prescribes no
  extraction or validation procedure as prose and performs no
  session-database reads); preserve the existing no-vote semantics
  (failed and non-success terminal runs do not vote; a persona with
  no successful run produces no persona vote; the panel majority
  counts only successful persona votes); document the
  missing-child-text behavior (the tool reports that persona's
  assessment fields as unavailable; verdict resolution proceeds
  from run records) and the all-child-text-unavailable aggregate
  case (assessment fields unavailable panel-wide; verdict
  resolution proceeds from run-recorded values alone)
- [x] 1.4 Edit Section 4.6: document `consolidate_dispatch` dual mode
  (no provenance = retrieval, no side effects; provenance =
  finalization); scope the exactly-once bound to the provenance-bearing
  call only; keep all existing finalization guarantees (provenance
  resolution, never call `finalize_review_dispatch`, authoritative
  response, bounded retry) unchanged
- [x] 1.5 Edit `.opencode/commands/uf.review-council.md` Section 4
  (Store Lesson Proposals): source proposals from the no-provenance
  `consolidate_dispatch` retrieval payload instead of claiming
  `dispatch_review_runs` returns them; store each via one
  `dewey_store_learning` call using the proposal's `information`,
  `tag`, and `category` (not `reference`); add the dual-mode note to
  Section 5 (Finalize Every Iteration) scoping the exactly-once
  bound to the provenance-bearing finalization call so the Section 4
  retrieval call cannot be read as violating it
- [x] 1.6 Edit `.opencode/commands/uf.address-feedback.md` Section
  2.3.7 (Store Lesson Proposals): apply the same corrections as
  task 1.5, with the dual-mode note added to Section 2.3.8
  (Finalize Every Tier 2 Dispatch)
- [x] 1.7 Edit `.opencode/commands/uf.review-pr.md` Section 5a
  (Store Lesson Proposals): add the sanctioned no-provenance
  `consolidate_dispatch` retrieval step after the command's
  `dispatch_review_runs` call and source proposals from that payload
  (store each via one `dewey_store_learning` call using the
  proposal's `information`, `tag`, and `category`); leave Section
  5b's direct `finalize_review_dispatch` finalization unchanged

## 2. Assessment-field extraction tool (review-dispatch plugin)

- [x] 2.1 Implement `read_assessment_fields` in
  `.opencode/plugins/review-dispatch/index.ts`: input is the
  dispatch's run identifiers (the `run_id` values from the
  consolidate payload's `runs`); for each run, read the persisted
  child final text (`PersistedRunData.text`) from the dispatch
  store, parse the structured response contract fields (category,
  objectivity, reasoning, split recommendations), and enforce the
  untrusted-input validation (Constitution V): `category` a
  non-empty string of at most 128 characters; `objectivity` a
  string of at most 64 characters; `reasoning` a string of at most
  4096 characters; split recommendations an array of at most 16
  strings of at most 512 characters each; type-check every value
  against the contract shape; skip values that fail validation as
  unavailable for that run; return per-run validated fields or
  unavailable. No session-database reads (`list_sessions` /
  `db_read_session` are not used). Register the tool with its zod
  input schema alongside the existing plugin tools
- [x] 2.2 Add unit tests for the tool: per-field bounds (at and
  over each limit), type mismatches skipped, contract-shape
  validation, missing or empty persisted text reported as
  unavailable, malformed structured response reported as
  unavailable, the split-recommendations array-length cap, and the
  happy path per field
- [x] 2.3 Add an integration test end-to-end through a dispatched
  run: a persisted child structured response returns validated
  fields; a disagreement case (child text verdict vs run-recorded
  `workflow_verdict`) returns fields while verdict resolution stays
  on run-recorded values
- [x] 2.4 Run `make plugin-test` (unit, integration, smoke,
  coverage) green with the new tool

## 3. Mirror sync and pin lockstep assets

- [x] 3.1 Copy the canonical file to
  `internal/scaffold/assets/opencode/commands/uf.triage-issue.md` and
  verify byte-identical parity with `cmp`
- [x] 3.2 Copy the three edited sibling command files to their
  mirrors under `internal/scaffold/assets/opencode/commands/` and
  verify byte-identical parity with `cmp` for each
- [x] 3.3 Verify the approved pin against upstream (AGENTS.md
  Gatekeeping): confirm npm registry integrity and each package's
  actual upstream license — MIT for `@opencode-ai/plugin` 1.19.0
  and `@opencode-ai/sdk` 1.19.0; Apache-2.0 for
  `@ai-sdk/provider` 3.x; AFL-2.1 OR BSD-3-Clause for
  `json-schema` 0.4.0 (do not assume MIT across the set) — and
  verify the msgpackr 1 → 2 major transitive bump (resolved
  2.1.1): review the 2.x changelog for breaking changes affecting
  the plugin runtime and confirm `make plugin-test` green;
  following the org precedent
  (`review-council-multi-model-fanout` task 1.5); record the
  verification evidence
> Reconciled by `stop-shipping-plugin-lock`: tasks 3.4, 3.5, 3.7,
> 4.3, and 4.4 and the locks in 4.5 record work done while the lock
> was tracked. Both package-lock.json files and
> `plugin-package-rollback.json` are deleted,
> `TestPluginPackageInputs_ExactVersionsAndLockIntegrity` is now
> `TestPluginPackageInputs_ExactVersions`, and no lockfile
> regeneration applies. The tasks stay as a historical record.

- [x] 3.4 Update `internal/scaffold/assets/opencode/package.json`
  and `internal/scaffold/assets/opencode/package-lock.json` to
  mirror the approved `@opencode-ai/plugin` 1.19.0 pin (matching
  `.opencode/package.json` / `.opencode/package-lock.json`)
  (Reconciliation: stop-shipping-plugin-lock later removed the
  lockfile mirror; the manifest mirror above remains the shipped
  artifact.)
- [x] 3.5 Update `TestPluginPackageInputs_ExactVersionsAndLockIntegrity`
  in `internal/scaffold/scaffold_test.go`: expectedRuntime
  `@opencode-ai/plugin` 1.4.10 → 1.19.0
  (Reconciliation: stop-shipping-plugin-lock later renamed the
  test to `TestPluginPackageInputs_ExactVersions`, dropping the
  lock-integrity half with the lockfile.)
- [x] 3.6 Update `internal/scaffold/testdata/plugin-package-provenance.json`:
  plugin 1.19.0, fresh `verified_at`, registry-maintainer and HTTPS
  provenance, and the task 3.3 verification evidence (consumed by
  `TestPluginPackageInputs_ApprovedMaintenanceProvenance`)
- [x] 3.7 Update `internal/scaffold/testdata/plugin-package-rollback.json`:
  re-verify the baseline-lock SHA256 against the 1.4.10 rollback
  baseline lockfile so verified rollback remains executable
  (consumed by `TestPluginPackageInputs_RollbackPrerequisites`)
  (Reconciliation: stop-shipping-plugin-lock later removed the
  rollback testdata and its test with the lockfile.)
- [x] 3.8 Run `make check` (lint, test, build, coverage gate, plugin
  tests) and confirm the scaffold drift and pin-governance tests
  pass

## 4. Pin verification and documentation

- [x] 4.1 Verify the Node engine floor locally (AGENTS.md
  Gatekeeping): assert the manifest `engines.node` equals ini 7's
  declared range (^22.22.2 || ^24.15.0 || >=26.0.0) and record the
  check and its result as reproducible evidence (not deferred to
  CI's floating `node-version: 22`; local Node v24.11.1 predates
  the >=24.15.0 24.x bound, so npm reports an advisory EBADENGINE
  warning, not a failure)
- [x] 4.2 Update `AGENTS.md` Technology Stack: `@opencode-ai/plugin`
  1.4.10 → 1.19.0 and the Node engine range (^22.22.2 ||
  ^24.15.0 || >=26.0.0); update `README.md` with the engine range
  and a note that `uf init` consumers inherit it via the
  scaffolded package manifests
- [x] 4.3 Validate the pin update: `make plugin-test` green against
  1.19.0 and `go test ./internal/scaffold -run
  TestPluginPackageInputs_ExactVersionsAndLockIntegrity` green;
  record the output as the pin-update validation evidence
- [x] 4.4 Declare the Node engine range in the `engines` field of
  `.opencode/package.json` and its scaffold asset mirror
  `internal/scaffold/assets/opencode/package.json` (ini 7 range
  ^22.22.2 || ^24.15.0 || >=26.0.0) so npm tooling enforces it
  rather than it being only documented; confirm the lockfile is
  unaffected or regenerate it in lockstep
- [x] 4.5 Commit sequencing (AGENTS.md commit-scope rule): stage and
  commit the pin lockstep file group (manifests, locks, scaffold
  asset mirrors, test expectation, testdata, AGENTS.md and README
  lines) as its own dedicated commit on the feature branch, staged
  by explicit path only (never `git add -A` or `git add .`),
  separate from the command-doc and plugin-tool commits; document
  the commit split in the PR description

## 5. Change finalization

- [x] 5.1 Verify `openspec status` shows all artifacts ready for
  `fix-triage-issue-dispatch-docs` and run `openspec validate
  fix-triage-issue-dispatch-docs`; a strict-mode failure limited to
  the known validator quirk with ADDED requirements (see design
  Risks) is not blocking — the gate is a clean non-strict
  validation plus the delta following the FR-NNN/MUST/Scenario
  convention; fix any other validation error before finalization
- [x] 5.2 Add the `CHANGELOG.md` entry for this change
- [x] 5.3 Before filing, search for duplicates: existing
  documentation issues in this repository for the triage-issue
  command docs, and `unbound-force/website` issue #303 (the
  cross-repo documentation gate tracker); reference or extend an
  existing issue instead of filing a duplicate where one exists.
  Then file the documentation issues required by the documentation
  gate for user-facing command-doc changes — both gates: the repo
  documentation issue (AGENTS.md gate, this repository) and the org
  constitution Cross-Repo Documentation issue in
  `unbound-force/website` (specs/026 and 030) — covering the
  command-doc corrections and the Node engine-floor documentation
  surface (README/AGENTS.md updates from task 4.2) — and record the
  issue numbers in the proposal
- [x] 5.4 File the follow-up tracking issues for the payload schema
  extension (assessment fields in the consolidate payload; when it
  lands, `read_assessment_fields` can read the payload instead of
  persisted text without command changes) and the
  `speckit-testreview` dispatch-data-flow audit, and record their
  numbers in the design's Open Questions
- [x] 5.5 Add the superseded-by annotation to SC-FR-003 in the
  unarchived `review-council-multi-model-fanout` delta, pointing at
  this change as the repo-wide pin supersession
