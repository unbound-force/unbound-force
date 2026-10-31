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

## 1. Implementation (already complete, council-reviewed)

- [x] 1.1 Relax `validateRunModel` in
  `.opencode/plugins/review-dispatch/index.ts`: the host
  branch rejects a successful run only when
  `resolved_parent_model` is null, with message
  "requires resolved parent model".
- [x] 1.2 Relax the host-resolved advisor branch
  (`source: "advisor"`, `requested_model: null`) identically:
  reject only a null `resolved_parent_model`.
- [x] 1.3 Sync the scaffold mirror
  `internal/scaffold/assets/opencode/plugins/review-dispatch/index.ts`
  byte-identically.
- [x] 1.4 Update `.opencode/test/finalize-error-boundaries.test.ts`:
  variantless host pass, variantless host-resolved advisor
  pass, and null-model rejection with the updated message for
  both sources.

## 2. Spec record and changelog (this change)

- [x] 2.1 Create this OpenSpec change with a MODIFIED delta
  on 'Finalization Model Validation' stating that successful
  host and host-resolved advisor runs MUST carry a non-null
  `resolved_parent_model` while `resolved_parent_variant`
  conforms to the variant-or-null schema contract and MAY be
  null; include the variantless-success scenarios and the
  corrected failure-message expectation.
- [x] 2.2 Update the CHANGELOG.md Unreleased/Fixed entry to
  reference documentation issue #891 recording the docs-gate
  outcome.
- [x] 2.3 Run `openspec validate
  accept-variantless-parent-model --strict` and confirm the
  delta passes.

## 3. Verification

- [x] 3.1 Plugin suite green from the implementation step
  (`make plugin-test`); no code files modified by this
  spec-record change.
- [x] 3.2 Verify constitution alignment: Autonomous
  Collaboration (artifact-recorded contract, no runtime
  coupling), Composability First (contract matches the shared
  `resolveHostModel()` writer), Observable Quality
  (spec/code/test observe one contract), Testability
  (scenarios map 1:1 to green tests), Security by Default
  (identity control unchanged), Deterministic Execution
  (record of an existing deterministic rule).
