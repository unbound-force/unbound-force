## ADDED Requirements

### Requirement: DAR-FR-001 Final text on report-time failures

When a `uf_delegate_command` run fails with error code
`child_message_error` or `model_mismatch`, the failure
envelope MUST carry a `final_text` field holding the
child's final assistant text, and a `final_text_source`
field naming where it came from: `response` (the prompt
response) or `transcript` (the child session's persisted
messages).

The wrapper MUST use the prompt response text when it
is non-empty after trimming. Otherwise it MUST read the
child session's most recent messages and use the text
of the latest assistant message whose text is non-empty.

`final_text` MUST pass through the invocation-error
sanitizer and MUST be bounded to
`MAX_DELEGATE_FINAL_TEXT_LENGTH` characters, keeping the
tail of the text (where reports end) behind a truncation
marker when longer.

When no non-empty text is available, or the transcript
read fails, the envelope MUST carry `final_text: null`
and `final_text_source: null`. A transcript read failure
MUST NOT change the error code, message, or exit path.

#### Scenario: Aborted child with response text

- **GIVEN** a delegation whose prompt response carries a
  `MessageAbortedError` and non-empty text
- **WHEN** `uf_delegate_command` returns
- **THEN** the envelope MUST carry that text as
  `final_text` with `final_text_source: "response"`

#### Scenario: Aborted child with empty response

- **GIVEN** a delegation whose aborted prompt response
  carries no text
- **AND** the child transcript holds an earlier
  assistant message with non-empty text
- **WHEN** `uf_delegate_command` returns
- **THEN** the envelope MUST carry that earlier text as
  `final_text` with `final_text_source: "transcript"`

#### Scenario: Model mismatch keeps the report

- **GIVEN** a delegation that fails with `model_mismatch`
  and non-empty response text
- **WHEN** `uf_delegate_command` returns
- **THEN** the envelope MUST carry that text as
  `final_text`
- **AND** the error MUST stay `retryable: false`

#### Scenario: Transcript unreadable

- **GIVEN** an aborted delegation with empty response
  text whose transcript read fails
- **WHEN** `uf_delegate_command` returns
- **THEN** the envelope MUST carry `final_text: null`
- **AND** the error code MUST stay `DELEGATE_FAILED`

### Requirement: DAR-FR-002 Resumable message aborts

When a command-path `uf_delegate_command` run fails with
`child_message_error` whose message error name is
`MessageAbortedError`, the failure envelope MUST carry
the same `resume` payload the timeout path emits
(`command`, `key`, `last_checkpoint_step`,
`completed_steps`).

The envelope MUST be `retryable: true` when the resume
payload is non-null, and `retryable: false` when it is
null.

Prompt-path delegations (no `command_path`) and every
other `child_message_error` name MUST keep their
current classification and carry no `resume` payload.

The shared executor MUST expose the child message
error name on its result (`message_error_name`) and MUST
NOT change its own `retryable` classification, so
`invoke_agent` and review dispatch behave as before.

#### Scenario: Command-path abort with durable state

- **GIVEN** a command-path delegation aborted with
  `MessageAbortedError`
- **AND** readable progress state for that command
- **WHEN** `uf_delegate_command` returns
- **THEN** the envelope MUST carry `resume` with that
  state's `last_checkpoint_step` and `completed_steps`
- **AND** the envelope MUST be `retryable: true`

#### Scenario: Prompt-path abort

- **GIVEN** a prompt-path delegation aborted with
  `MessageAbortedError`
- **WHEN** `uf_delegate_command` returns
- **THEN** the envelope MUST NOT carry `resume`
- **AND** the envelope MUST be `retryable: false`

#### Scenario: Other message errors

- **GIVEN** a command-path delegation whose child
  message error is not `MessageAbortedError`
- **WHEN** `uf_delegate_command` returns
- **THEN** the envelope MUST keep the executor's
  `retryable` value and MUST NOT carry `resume`

### Requirement: DAR-FR-003 Report diagnostics

A `child_message_error` or `model_mismatch` failure
envelope MUST carry a `report_diagnostics` object with:

- `error_name`: the child message error name, or null
- `elapsed_ms`: milliseconds from the delegation's
  executor call to its result
- `messages_examined`: how many recent child messages
  the transcript read returned, or null when unread
- `last_tool`: the latest tool part's tool name and
  status in the examined messages, or null
- `last_input_tokens`: the latest assistant message's
  reported input tokens, or null
- `final_turn`: the latest user message's `agent` and
  `model` (`provider/model`), the latest assistant
  message's `model`, and `injected` — true when that
  user message's text parts are all marked synthetic —
  or null when no user message was examined

Every field MUST fail open to null; diagnostics MUST
NOT change the error code, message, or exit path.

#### Scenario: Injected turn ends the run

- **GIVEN** a delegation whose child transcript ends
  with a synthetic user message from agent `build` on
  model `p/other` and an assistant reply on `p/other`
- **WHEN** it fails with `model_mismatch`
- **THEN** `report_diagnostics.final_turn` MUST report
  agent `build`, model `p/other`, and `injected: true`

#### Scenario: Diagnostics without a transcript

- **GIVEN** an aborted delegation whose transcript read
  fails
- **WHEN** `uf_delegate_command` returns
- **THEN** `report_diagnostics` MUST carry `error_name`
  and `elapsed_ms`
- **AND** its transcript-derived fields MUST be null
