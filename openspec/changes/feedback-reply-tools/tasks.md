<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the same group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Schema 1.1.0 (#999)

- [x] 1.1 Add `schemas/feedback-triage/v1.1.0.schema.json` (design
  D9): `file` and `reply` decisions, optional `filed_issue_number`,
  `filed_issue_url`, `filed`, and `replied`.
- [x] 1.2 Add `samples/v1.1.0/` with one positive sample and three
  `invalid-*` samples; register the schema in
  `handAuthoredSchemas` (`internal/schemas/ci_test.go`) and
  generalize the directory check for version subdirectories.
- [x] 1.3 Add `TestFeedbackTriageSchema_MinorVersionCompatibility`:
  1.0.0 samples validate under 1.1.0, the 1.1.0 sample fails 1.0.0,
  the enum and the optional counts are pinned.
- [x] 1.4 Update `schemas/feedback-triage/README.md`: 1.1.0 fields,
  the FILE and REPLY record, historical 1.0.0 prefixes, version
  compatibility, and version history.

## 2. Tools (#966, #998)

- [x] 2.1 Add `.opencode/lib/feedback-replies.ts` (design D1, D3–D5,
  D7, D8): input schemas, endpoint selection, reply rendering,
  issue-URL parsing, gh error classification, the feedback cache
  read/write, the reviewThreads lookup, `postFeedbackReplies`, and
  `fileFeedbackIssue`, with an injectable gh executor.
- [x] 2.2 Register `post_feedback_replies` and
  `file_feedback_issue` in `uf-workflow` with `args` set to the
  schemas' shapes and the `reply-comments` auto-checkpoint
  (design D6).
- [x] 2.3 Add `.opencode/test/feedback-replies.test.ts`: helpers,
  input validation, happy path with ordering and injection safety,
  repo resolution, declined resolution, pagination, failure-stop,
  issue retry and failure, no-URL no-retry, crash-after-filing,
  thread lookup failures, resolve retry, cache permissions and
  symlink refusal, deduplication, envelopes, default executor.
- [x] 2.4 Add the two tools to the uf-workflow name lists in
  `uf-workflow.test.ts` and `scratch-smoke.test.ts`.
- [x] 2.5 DE-007 seam verification naming the three layers:
  registration surface (the generic `registration-surface.test.ts`
  suite enumerates both tools; `feedback-replies.test.ts` pins
  `args` key parity with `PostFeedbackRepliesInputSchema` and
  `FileFeedbackIssueInputSchema`); documented output shape
  (`golden-output-shape.test.ts` runs both tools through the
  registry and pins the Section 4.6 fields: `complete`, `counts`,
  `items[]` outcomes, `summary`, `checkpoint_recorded`,
  `issue_number`, `issue_url`, `deduplicated`); write→read
  round-trips (the tool's `reply-comments` checkpoint reads back
  through `get_progress`; its cache record reads back from
  `state.json`; a partial run is resumed by a second call; an issue
  filed by `file_feedback_issue` is reused by
  `post_feedback_replies`). `tool-inventory.ts --seams` reports no
  missing layer.

## 3. Command

- [x] 3.1 Rewrite Sections 4.5 and 4.6 of
  `.opencode/commands/uf.address-feedback.md`: compose and both
  gates, then one `post_feedback_replies` call with the parameter
  mapping, the rendered-summary examples (DE-002), and the
  error-envelope STOP. Leave Sections 1.2.1 and 4.4 untouched.
- [x] 3.2 Update Section 1.7 crash recovery, the Progress Tracking
  step list and TodoWrite list, the Phase 4 plan note, Section 4.7
  (schema 1.1.0, `file` and `reply` decisions, `clear_progress`
  after the artifact), and Guardrail 8.
- [x] 3.3 Copy the command, plugin, and new lib module to the
  scaffold assets; add `opencode/lib/feedback-replies.ts` to the
  tool-owned lists in `scaffold.go` and `plugin_activation.go` and
  to the scaffold test asset lists.
- [x] 3.4 Replace the issue-filing and artifact-prefix pins in
  `TestCommandContracts_FeedbackLoopScope`, add removed-text checks
  for the shell, and add `TestCommandContracts_FeedbackReplyTools`.

## 4. Documentation and verification

- [x] 4.1 Add the CHANGELOG entry with the DE-004 instruction delta.
- [x] 4.2 File the documentation issues for the user-visible changes
  (resolve question before posting, feedback-triage 1.1.0):
  unbound-force/unbound-force#1003 (in-repo) and
  unbound-force/website#401 (website).
- [x] 4.5 File the follow-up to extract the 4.7 feedback-triage
  artifact write into a tool (DE-001): unbound-force/unbound-force#1022.
- [x] 4.3 Run `make check` and `make lint-instructions`.
- [x] 4.4 Run `openspec validate feedback-reply-tools --strict`.
