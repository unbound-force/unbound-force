<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the same group, (b) has no dependency
  on prior tasks in the group, and (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Error preview integrity (design D1, D2)

- [x] 1.1 In `.opencode/plugins/review-dispatch/index.ts`, drop
      `errorChars` from `SUMMARY_LIMITS` and make
      `buildCompactFinalization` copy each error string verbatim;
      `errors_truncated` is true only when whole entries were
      dropped (count cap or byte budget), never for shortening
      (RD-FR-066).
- [x] 1.2 Extend the compact builder's post-assembly byte-budget
      loop: after proposals are exhausted, drop trailing errors
      whole and set `errors_truncated`, keeping the result
      strictly under 40,960 bytes (RD-FR-066, D2). Do not alter
      the budget constant.
- [x] 1.3 Add `failure_record` (`null` or
      `{ correlation_id, path }`) to
      `FinalizeReviewDispatchResult` and the compact result,
      inserted immediately after `artifact_path` and omitted when
      null (RD-FR-066, D3).

## 2. Failure record persistence (design D3, D4)

- [x] 2.1 In `.opencode/plugins/review-dispatch/index.ts`, add
      the strict failure-record payload schema
      (`dispatch_correlation_id` nullable, `command`/`mode`
      optional, `status: "failed"`, `operation_verdict`,
      `errors` with sanity bounds of 131,072 characters per
      entry and 65,536 entries — fail closed, never truncate)
      and the envelope builder emitting
      `artifact_type: "review-dispatch-failure"` with
      best-effort provenance context.
- [x] 2.2 Add `persistFailureRecord` over
      `FinalizationDependencies` using the `persistEnvelope`
      atomic discipline: 0o750 directory, exclusive-create temp
      file, rename-by-link, cleanup on failure, and `EEXIST`
      collision retry that mints a fresh record correlation id.
- [x] 2.3 Wire best-effort record writes into all three
      `finalizeReviewDispatch` failure paths: structural
      rejection (use the input's correlation id only when
      UUID-usable, else mint), semantic rejection, and
      persistence failure. A record-write failure MUST NOT mask
      the finalization failure: return `failure_record: null`
      and append one `persist failure record: <message>` entry
      to the result's `errors` (RD-FR-066, D4).

## 3. Reader section and tool descriptions (design D5, D7)

- [x] 3.1 In `.opencode/plugins/review-dispatch/index.ts`,
      extend `read_dispatch_artifact` with section `errors`:
      derive `.uf/artifacts/dispatch/<correlation_id>.failure.json`,
      apply the existing regular-file / 64 MiB / parse / type /
      payload validation chain with the RD-FR-057 error codes,
      and return `{ status: "ok", correlation_id, section,
      total, errors }` complete and untruncated; a dispatch with
      only a success artifact returns `not_found` (RD-FR-066).
- [x] 3.2 Correct the two tool descriptions: the
      `consolidate_dispatch` errors-cap sentence now describes
      complete-or-dropped entries, the `failure_record` key, and
      its position after `artifact_path`; the
      `read_dispatch_artifact` description gains the `errors`
      section. No command workflow prose changes (DE-001).

## 4. Tests (regression per TC-001, error paths per TC-003)

- [x] 4.1 In `.opencode/test/summary-retrieval.test.ts`: a
      composed validation error longer than 256 characters
      survives whole with no `truncated` key; a 7-error overflow
      keeps five whole entries with `errors_truncated`; the byte
      budget drops trailing errors whole; the parsed key order
      places `failure_record` after `artifact_path` and `errors`
      last; update the `SUMMARY_LIMITS` literals.
- [x] 4.2 [P] In `.opencode/test/finalize-review-dispatch.test.ts`
      and `.opencode/test/finalize-error-boundaries.test.ts`:
      a `review-dispatch-failure` record is written on each of
      the three failure paths with the complete errors verbatim;
      a garbage input with no usable correlation id yields a
      minted record id and `dispatch_correlation_id: null`;
      `EEXIST` appends a new record without overwriting; a
      record-write failure returns `failure_record: null` with
      the write error among the result's `errors`.
- [x] 4.3 [P] In `.opencode/test/read-dispatch-artifact.test.ts`:
      section `errors` returns the complete array for a written
      record; `not_found` for a success-only correlation;
      `invalid_artifact` for a missing, oversized, symlinked,
      unparseable, wrong-type, or schema-failing record.
- [x] 4.4 [P] In `.opencode/test/review-dispatch-artifact.test.ts`:
      `parseReviewDispatchArtifact` returns null for a
      `review-dispatch-failure` envelope (resume detection keeps
      ignoring failure records).
- [x] 4.5 [P] In `.opencode/test/consolidate-dispatch.test.ts`:
      end-to-end failed finalization surfaces `failure_record`
      in the compact result and the record is readable through
      `read_dispatch_artifact` section `errors`.

## 5. Schema registry and scaffold mirrors (design D6)

- [ ] 5.1 [P] Author `schemas/review-dispatch-failure/v1.0.0.schema.json`
      declaring the record payload (strict, sanity bounds
      matching task 2.1), following the hand-authored schema
      conventions of `schemas/review-dispatch/v1.0.0.schema.json`.
- [ ] 5.2 Register the schema in `internal/schemas` fixture
      validation (`handAuthoredSchemas` in `ci_test.go`) with a
      valid fixture.
- [ ] 5.3 Mirror the schema to
      `.uf/schemas/review-dispatch-failure/` and
      `internal/scaffold/assets/uf/schemas/review-dispatch-failure/`,
      and add the path to the scaffold asset list in
      `internal/scaffold/scaffold_test.go`.
- [ ] 5.4 Copy every edited `.opencode/` file byte-identical to
      `internal/scaffold/assets/opencode/` (plugin, and lib only
      if touched), keeping the drift tests green.

## 6. Documentation, CI parity, and constitution verification

- [ ] 6.1 Add a CHANGELOG.md entry describing the
      whole-or-absent error preview and the failure record.
- [ ] 6.2 File the in-repo documentation issue for the
      user-facing tool-description changes (documentation gate;
      `uf init` deploys the mirrored plugin).
- [ ] 6.3 Run `make check` (lint, test, build, coverage gate,
      plugin tests) and confirm the 94 percent plugin statements
      floor and the review-dispatch 85 percent branch gate hold.
- [ ] 6.4 Verify constitution alignment (v1.4.0): the failure
      record is a self-describing envelope artifact with
      provenance (I, III); no new runtime coupling and scaffold
      mirrors byte-identical (II); writer, reader, and projection
      are unit-tested over injected dependencies with no external
      services (IV); the record is written atomically with
      bounded, fail-closed payload schemas (V); behavior lives in
      tool code with prose corrections only (VI).
