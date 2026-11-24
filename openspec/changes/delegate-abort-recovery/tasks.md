<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file --
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Executor and recovery library

- [ ] 1.1 [P] Add `message_error_name` to
  `InvokeAgentResult` and set it on the
  `child_message_error` branch of `executeAgentSession`
  in `.opencode/lib/agent-executor.ts`, keeping the
  existing `retryable` value (DAR-FR-002, D2).
- [ ] 1.2 [P] Create `.opencode/lib/report-recovery.ts`
  with `MAX_DELEGATE_FINAL_TEXT_LENGTH`,
  `MAX_RECOVERY_MESSAGES`, `boundFinalText`, the pure
  `summarizeTranscript`, and the fail-open
  `readChildTranscript` (DAR-FR-001, DAR-FR-003, D1,
  D3-D5).

## 2. Wrapper wiring

- [ ] 2.1 In the `uf_delegate_command` failure branch of
  `.opencode/plugins/uf-workflow/index.ts`, for
  `child_message_error` and `model_mismatch`: read the
  transcript once, add `final_text`,
  `final_text_source`, and `report_diagnostics`; for a
  command-path `MessageAbortedError`, add `resume` and
  set `retryable` from it (DAR-FR-001..003).
- [ ] 2.2 Add `final_text`, `final_text_source`,
  `report_diagnostics`, and the abort `resume` to the
  `DelegateCommandData` documentation.

## 3. Tests

- [ ] 3.1 [P] Unit tests for `report-recovery.ts`:
  bound and truncation marker, sanitization, latest
  non-empty assistant text, injected final turn,
  missing fields, read failure (`.opencode/test/report-recovery.test.ts`).
- [ ] 3.2 [P] Delegate tests in
  `.opencode/test/unleash-delegate.test.ts` for every
  DAR scenario: response text, transcript fallback,
  unreadable transcript, command-path abort with and
  without resume state, prompt-path abort, other
  message errors, and model mismatch with an injected
  turn.
- [ ] 3.3 Executor test pinning `message_error_name`
  and the unchanged `retryable` classification.

## 4. Verification

- [ ] 4.1 Run `make plugin-test` (unit, integration,
  smoke, 94 percent coverage floor) and `make check`.
- [ ] 4.2 Add a `CHANGELOG.md` entry.
- [ ] 4.3 Verify constitution alignment: Observable
  Quality (envelope fields), Autonomous Collaboration
  (durable-state resume), Security by Default
  (sanitized text).
