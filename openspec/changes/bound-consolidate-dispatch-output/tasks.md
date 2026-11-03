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

## 1. Live plugin contract (design D1-D3)

- [x] 1.1 In `.opencode/plugins/review-dispatch/index.ts`, extend
      `ConsolidateDispatchInputSchema` (L3766) with `detail: z.enum(["full", "summary"]).default("full")` and a `superRefine` that rejects `detail: "summary"` when
      `provenance` is present (message: "detail: summary is only valid without provenance"). Keep `.strict()`. Update the `createConsolidateDispatchTool` description
      (L4171-4191) to document `detail`, the summary field order and the 15/15/10/5 caps (512-character descriptions, 256-character files, 512-character proposal
      information, 128-character tags, 128-character proposal agents, 512-character `verdict_reason`, 64-character warning kinds, 256-character warning details composed
      per the design warning-derivation table) with per-entry `truncated: true` markers, the four truncation flags, the compact finalization result, the five warning kinds
      and the proposal validation schema, and the observability-first ordering (`correlation_id`, `proposals`, `warnings` before `verdict` and the large arrays). Add no
      new export.
- [x] 1.2 In the same file, extract three builders beside
      `createConsolidateDispatchTool` and rewire both `JSON.stringify` return sites (L4514-4523 and L4526-4535): (a) `summaryRetrievalResult(payload, proposals, warnings)`
      returning exactly `correlation_id`, `proposals?`, `warnings?`, `verdict`, `verdict_reason`, `run_counts`, `coverage`, `findings`, `advisories`,
      `findings_truncated?`, `advisories_truncated?`, `proposals_truncated?`, and `warnings_truncated?` (design D1), never `artifact_path`; findings/advisories projected
      to `{severity, category, description, file, line}` (no `root_cause`, no `run_ids`) and capped at 15 entries each, with `category` and `tag` truncated to 128,
      `description` to 512, `file` to 256, and `information` to 512 characters, proposal `agent` to 128, and the summary `verdict_reason` to 512 characters; proposals
      capped at 10 and warnings projected to `{kind, detail}` and capped at 5, with `kind` truncated to 64 and `detail` to 256 characters; each included entry carries
      `truncated: true` when any emitted field was shortened, and each cap sets its flag only when entries are dropped. After assembly, serialize the summary compactly (no
      indentation), measure its UTF-8 byte length, and while it is at least 40,960 bytes drop trailing entries in the fixed order findings, advisories, proposals,
      warnings, setting the matching flag, until it is strictly under budget; the loop terminates because the scalar-only summary is far below 40,960 with `verdict_reason`
      capped; (b) `fullRetrievalResult(payload, proposals, warnings)` rebuilding the full object in the D2 insertion order so `correlation_id`, `proposals`, and `warnings`
      precede `verdict`, `runs`, `plan`, `findings`, and `advisories`, with empty `proposals` / `warnings` omitted; (c) `compactFinalizationResult(result, proposals,
      warnings)` returning `status`, `operation_verdict`, `correlation_id`, `artifact_path`, scalar `review_verdict`, `run_counts`, `proposals?`, `warnings?`,
      `proposals_truncated?`, `warnings_truncated?`, `calculated_assessment` as null or `{ authoritative }`, then `errors` — proposals capped at 10 (`information` 512,
      `tag` 128) and warnings at 5 (`kind` 64, `detail` 256) with the same per-entry marker and flags. Keep the object passed to `finalizeReviewDispatch` schema-exact: no
      `proposals` or `warnings` keys. Route the no-provenance return through `detail === "summary" ? summaryRetrievalResult(...) : fullRetrievalResult(...)`.
- [x] 1.3 Checkpoint: run
      `cd .opencode && npx vitest run test/finalize-error-boundaries.test.ts` — this suite MUST stay green. Two suites are known-red until Phase 2 updates their legacy
      assertions: `test/consolidate-dispatch.test.ts` (task 2.2) and `test/golden-output-shape.test.ts` (finalization assertions L1163 and L1198-1202, task 2.1). Any
      failure outside those files MUST be fixed before Phase 2.

## 2. Test coverage (RD-FR-049..051, design D6-D7)

- [x] 2.1 [P] In `.opencode/test/golden-output-shape.test.ts`, update the
      `CONSOLIDATE_SITES` finalization assertions (L1152-1205) to the compact shape: `review_verdict` is a scalar `council_decision` string, `run_counts` equals the
      payload counts, `calculated_assessment` equals `{ authoritative: true }` (or is null on failure), and the result carries no `review_dispatch` or
      `calculated_assessment.payload`. Add a `detail: "summary"` shape test asserting the exact RD-FR-049 field set and the absence of `plan`, `runs`, `input_context`,
      `change_profile`, `plan_version`, `workflow_result`, `command`, `mode`, `full`, and `artifact_path`. Add ordering tests over `Object.keys(JSON.parse(output))`: for
      the summary result `correlation_id` precedes `proposals`, `proposals` precedes `warnings`, and `warnings` precedes `findings`; for the full result `proposals` and
      `warnings` precede `runs`, and `correlation_id` precedes `plan`. Do not use raw-string `indexOf` for ordering. Add a caps test: a session with 512 findings, 512
      advisories, 512 proposals, and 512 warnings yields 15/15/10/5, sets all four truncation flags, truncates descriptions to 512, file to 256, proposal information to
      512, tags to 128, warning kind to 64 and warning detail to 256 characters, marks shortened entries `truncated: true`, preserves the full arrays in the persisted
      artifact, and serializes strictly under 40,960 bytes with every included field pinned at its entry maximum. Add the equivalent compact-finalization caps/order test:
      12 proposals and 7 warnings must come back as 10/5 with both flags, the parsed key order must keep `proposals`, `warnings`, and the flags before
      `calculated_assessment` and `errors`, and the artifact must keep the complete findings/advisories. Assert the proposal `agent` cap of 128 characters. Add a
      multibyte, escape-heavy fixture (CJK text, astral-plane emoji, embedded quotes, backslashes, and control-character escapes) whose summary still serializes strictly
      under 40,960 bytes through the post-assembly drop loop, verifying the drop order findings, advisories, proposals, warnings and the flags set. Add a size-budget test:
      a session with 16 unique findings and 7 proposals, retrieved with `detail: "summary"`, serializes to fewer than 40,960 UTF-8 bytes and carries all 7 proposals.
- [x] 2.2 [P] In `.opencode/test/consolidate-dispatch.test.ts`, add summary-mode tests
      (field set, omitted empty keys), default-full compatibility, `detail: "summary"` plus provenance rejection, unknown `detail` rejection, and compact finalization
      success/validation-rejection/persistence-failure assertions. The persistence-failure case MAY inject a failing `link` or `open` through `FinalizationDependencies`.
      Add a full-mode omission test: a session with no proposals and no warnings returns a full result with neither key, since `fullRetrievalResult` rebuilds the object.
      Update every legacy provenance assertion site to the compact shape and parsed ordering: L987-999, L1012-1020, L1206-1259, L1338, L1377, L1482, L1530, L1576, L1624,
      L1697-1722, L1841, L1862, L1981, L2002, L2069, and L2092 — scalar `review_verdict`, `run_counts`, `calculated_assessment`, absence of `review_dispatch` /
      `calculated_assessment.payload`, and no raw-string ordering checks.
- [x] 2.3 [P] In `.opencode/test/finalize-error-boundaries.test.ts`,
      add a boundary test proving the compact finalization wrapper is not a payload: submit its top-level fields (including `calculated_assessment: { authoritative: true
      }`) to `finalizeReviewDispatch` and assert strict root rejection via `errorsFor`, without changing the existing wrapper-observability test at L473.
- [x] 2.4 Checkpoint: run `make plugin-test` — the full vitest
      suite with the 94 percent statements coverage floor MUST pass. Seam verification (DE-007) in this increment: Layer 1 registration/input-schema parity (`detail`
      accepted, invalid value rejected), Layer 2 documented output shape (summary and compact finalization golden tests), Layer 3 round-trip (compact wrapper rejected as a
      payload). If the floor fails, report and stop; never adjust `.opencode/vitest.config.ts`.

## 3. Command contracts (RD-FR-052, design D4-D5)

- [x] 3.1 In `.opencode/commands/uf.review-council.md` Section 4 (L383-431), replace the
      unconditional "Use this single payload — do not re-fetch it" instruction with the deterministic summary retrieval: call `consolidate_dispatch` without provenance
      with `correlation_id` and `detail: "summary"`; use a complete reply directly; on a host truncation notice, accept the path only from the exact single-line notice
      shape (a double-quoted absolute path may contain spaces; an unquoted path may not), ignore the notice's second line recommending explore delegation, confirm the path
      is absolute, has no `..` segment, resolves through `realpath` to a regular file strictly beneath the opencode tool-output directory (path starts with that directory
      plus `/`, never equal to it), has no symbolic link at any component, and is at most 1 MiB; read that file exactly once yourself, discard the truncated leading
      fragment, parse one JSON object, extract and schema-validate `proposals` / `warnings`, and store; fail closed on any mismatch, parse error, or validation failure.
      Retry the identical call exactly once on any tool error; a second error fails closed. Preserve the existing valid-outcome and not-found branches: zero completed and
      zero failed runs is a valid outcome (record zero counts, do not fail closed); a forwarded `correlation_id` with no persisted session is a terminal not-found; a
      missing forwarded `correlation_id` means capture is skipped. Apply the notice-shape detection to every `consolidate_dispatch` response, not only the Section 4 call.
      Keep the prohibition on `dispatch_review_runs`, child transcripts, and session-database reads. Also correct the stale sentence in Section 3 (L370-373) that says the
      response body is the final artifact, so it points at `artifact_path`. In Section 5 (L433-523), replace references to object `review_verdict` and
      `calculated_assessment.authoritative` with the scalar `review_verdict` decision and the scalar `calculated_assessment.authoritative` flag; keep the null-versus-false
      mechanical-retry classification and state that the artifact body lives at `artifact_path`.
- [x] 3.2 [P] In `.opencode/commands/uf.review-pr.md`, update only
      Section 5 finalization consumption (L527-600) to the compact result fields. It already makes exactly one provenance-bearing `consolidate_dispatch` call and MUST NOT
      add a second no-provenance retrieval: proposals/warnings come from the compact result, findings/advisories from the artifact at `artifact_path` (`payload.findings` /
      `payload.advisories`). Remove the claim that the response itself carries findings and advisories. Also correct the stale claim in Section 6 (L672-675) that describes
      the finalization response as returning the full payload; point it at `artifact_path`.
- [x] 3.3 [P] In `.opencode/commands/uf.triage-issue.md`, keep
      Section 2.3 retrieval (L285-391) on the default `full` mode — its three-rule verdict consolidation reads `runs`, which summary omits; record that rationale in the
      section, and make the section fail closed when a truncation notice is detected rather than consolidate a partial `runs` array. Retain the existing `Use this single payload — do not re-fetch it` instruction (L291-293, L364, L391) because this section stays on full mode and gains fail-closed-on-notice handling. Update only Section 4.6 finalization
      consumption (L1091-1150) to the compact result fields and the artifact read.
- [x] 3.4 [P] In `.opencode/commands/uf.address-feedback.md`,
      keep Section 2.3.7 retrieval (L490-520) on the default `full` mode, make it fail closed on a detected truncation notice rather than consolidate a partial `runs`
      array, retain the existing `Use this single payload — do not re-fetch it` instruction (L496) for the same reason, and update only Section 2.3.8 finalization consumption (L565-610) to the compact result fields and the artifact read.
- [x] 3.5 [P] In `.opencode/commands/uf.unleash.md`, update Step 4
      (L696-725) and Step 6 including the re-run branches (L1606-1690) to branch on `operation_verdict` (the workflow outcome) and treat `review_verdict` as display only,
      and to read findings and advisories once from the artifact at `artifact_path` (`payload.findings` / `payload.advisories`) instead of from the compact finalization
      result; keep the scalar verdict and run counts from the result. Add both call sites to the golden test's documented consumer list.
- [x] 3.6 Checkpoint: verify no command still reads a removed
      field or carries the old assertion. Run `rg -n "do not re-fetch" .opencode/commands/uf.review-council.md .opencode/commands/uf.review-pr.md
      .opencode/commands/uf.unleash.md` and `rg -n "calculated_assessment\.payload|review_verdict\.schema_version|\breview_dispatch\b" .opencode/commands`, plus a
      claim-sentence pattern `rg -n "response (body|payload) (is|contains|carries)|single payload — do not|as consolidate_dispatch returned|exact correspondence" .opencode/commands/uf.review-council.md
      .opencode/commands/uf.review-pr.md .opencode/commands/uf.unleash.md`, and confirm no stale matches remain. Confirm the new fail-closed obligation is present with `rg
      -n "truncation notice" .opencode/commands/uf.triage-issue.md .opencode/commands/uf.address-feedback.md`, and add a content-gate test in
      `internal/scaffold/scaffold_test.go` asserting that phrase is present in the five embedded command assets. Then run `make plugin-test` and confirm the suite stays
      green. Failing any check blocks Phase 4.

## 4. Scaffold mirror, documentation, CI parity

- [x] 4.1 Copy the edited live plugin and the five edited command
      files byte-identical into their scaffold mirrors under `internal/scaffold/assets/opencode/plugins/review-dispatch/index.ts` and
      `internal/scaffold/assets/opencode/commands/`, then run `go test -race -count=1 ./internal/scaffold/...` — `TestEmbeddedAssets_MatchSource` MUST pass.
- [x] 4.2 Add the CHANGELOG.md entry (issue #922):
      `consolidate_dispatch` gains `detail: "summary"` with its runtime caps, emits `proposals` / `warnings` before the large arrays in full mode (omitting both keys when
      empty), and returns a compact finalization result. Mark the removal of `review_dispatch`, `calculated_assessment.payload`, and object `review_verdict` from the
      finalization result as **breaking**, with the migration: read the persisted payload at `artifact_path` and the scalar `review_verdict`. State the omit-when-empty
      semantics precedented by PR #776.
- [ ] 4.3 Documentation-gate assessment across `CHANGELOG.md`,
      `AGENTS.md`, `README.md`, and `docs/`: the tool output shape is user-facing to calling agents and operators, so enumerate each documented surface in `docs/` and the
      command files, search for an existing open documentation issue before filing (avoid duplicates), then file the in-repo documentation issue and the
      unbound-force/website documentation issue with `--label docs`, cross-referencing in-repo #717 and #908 and website #303 and #367. Verify all four cross-references
      resolve, record both new issue numbers in the PR body and CHANGELOG entry, and link them.
- [ ] 4.4 Constitution-alignment verification (proposal assesses
      all six principles): confirm the retrieval result is bounded and self-describing (I, III), the change stays inside the plugin and scaffold mirror with no new
      dependency (II), the tests use injected filesystem dependencies and no external service (IV), no new credential, external service, or unvalidated input surface is
      added and the spill path is confined and validated (V), and the primary retrieval path is now tool-rendered with the DE-001 deviation recorded explicitly (VI).
      Record the verification summary in the PR body.
- [ ] 4.5 File the tracked follow-up issue for the narrow
      read tool that replaces the RD-FR-052 spill fallback; record the DE-001 deviation, the trigger (fallback firing more than once per month), and the owner
      (review-dispatch maintainer) in it, and cross-reference it from design.md and the PR body.
- [ ] 4.6 Record the cross-capability archive order and
      reconciliation in the PR body: `dispatch-finalize-determinism` MUST archive before (or in the same batch as) this change, so its dispatch-finalization FR-001 and
      review-council "Finalize Every Iteration Via Consolidation" requirements exist before RD-FR-051 is archived; restate that RD-FR-051's compact result is the
      finalizer's returned data for outcome/scalar purposes while the full projection is read once from `artifact_path`. Do not edit the sibling change's artifacts.
- [x] 4.7 Final CI parity (MP-023): run `make plugin-test` and `go test -race -count=1 ./internal/scaffold/...` — both MUST pass before the change is marked complete.
