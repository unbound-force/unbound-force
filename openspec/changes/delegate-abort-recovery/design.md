## Context

Issue 1199 reports `uf_delegate_command` command-path
runs failing at report time with `child_message_error:
MessageAbortedError` (classified `retryable: false`,
no `resume`) or `model_mismatch`, while the child had
already left durable work. The wrapper drops the
executor's final text from failure envelopes, so the
`/uf.unleash` demo block never reaches the caller.

Forensics on the host database traced both shapes to
the opencode-agent-supervisor plugin: on a severe audit
finding it calls `session.abort` on the child (the
executor's blocking `session.prompt` then resolves with
the aborted message), and every finding injects a
synthetic `[supervisor]` user turn with no agent or
model, which runs as `build` on the default model and
becomes the child's final message. The supervisor fix
(jflowers/opencode-supervisor#86) and the gh-hud
`repair-only` stopgap (jflowers/gh-hud#182) remove the
trigger. This change makes the wrapper robust to any
outside abort or injected turn, from that plugin or
another.

Constitution alignment (proposal): Observable Quality
and Autonomous Collaboration are the principles this
design serves; see D3 and D2.

## Goals / Non-Goals

### Goals
- Keep the child's final report on report-time failures
- Let orchestrators resume an aborted command-path
  delegation from durable state
- Make the turn that ended a failed run visible

### Non-Goals
- Changing model selection or the prompt body (the
  model is already passed; see proposal)
- Reclassifying `model_mismatch` (a governance check:
  it stays `retryable: false`)
- Changing `invoke_agent` or review-dispatch behavior
- Waiting for or re-prompting the child after an abort

## Decisions

**D1: Recovery lives in the wrapper, reads once.**
`uf_delegate_command` owns the failure envelope, so it
assembles `final_text`, `resume`, and
`report_diagnostics`. A new `.opencode/lib/report-recovery.ts`
holds the pure transcript summary and a fail-open
reader that calls `client.session.messages` once with
`limit: MAX_RECOVERY_MESSAGES` (50; the endpoint returns
the most recent messages, oldest first). The read runs
only for `child_message_error` and `model_mismatch`, so
successful and timeout paths pay nothing.

**D2: The executor names the message error; the
wrapper decides resumability.** The executor adds
`message_error_name` to its result on the
`child_message_error` branch and keeps its own
`retryable` value. Only the command-path wrapper, which
can read durable progress state, upgrades a
`MessageAbortedError` to retryable, and only when
`readDelegateResumeState` returns non-null. This keeps
`invoke_agent` and review dispatch (RD-FR-026 keys on
the error code) unchanged.

**D3: Diagnostics name the final turn.** The injected
turn is a user message whose text parts are all
`synthetic: true`. `final_turn` reports that message's
agent and model next to the final assistant model, so a
`model_mismatch` caused by an injected turn reads as
such instead of as a model-matrix fault. `last_tool`
and `last_input_tokens` cover the issue's "last tool
call, context usage" request.

**D4: Tail-preserving bound.** Reports end with their
demo block, so `final_text` keeps the last
`MAX_DELEGATE_FINAL_TEXT_LENGTH` (16 384) characters
behind a leading `…[truncated N chars]` marker, after
`sanitizeInvocationError` redaction. The existing
2 048-character head bound for compaction excerpts is
unchanged.

**D5: Fail open everywhere.** A transcript read error,
a missing endpoint, or an unexpected message shape
yields null fields. The error code, message, and
`DELEGATE_FAILED` code never change because of
recovery.

## Risks / Trade-offs

- A recovered transcript text may predate the abort by
  several turns; `final_text_source: "transcript"`
  tells the caller it is not the final response.
- `retryable: true` invites a re-delegation that may
  race a child the injected turn keeps running. The
  resume payload points the caller at durable state,
  which the pipelines already treat as idempotent.
