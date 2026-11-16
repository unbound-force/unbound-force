## Context

`/uf.address-feedback` Section 4.7 writes the feedback-triage
artifact by hand: a 53-line JSON envelope template plus prose for
the round scan ("highest round number and add 1 (not a file
count)"), the atomic write ("write to a temp file first, then
rename"), FILE/REPLY field routing from the
`post_feedback_replies` result, and the terminal `clear_progress`
call. `feedback-reply-tools` (#1007) extracted everything around
it — reply posting, issue filing, thread resolution — and listed
this write as a Non-Goal tracked in #1022. The council review of
#1007 flagged the section as remaining DE-001 material.

The precedent is in place: `.opencode/lib/feedback-replies.ts`
already owns the command's mechanics behind
`post_feedback_replies`, `file_feedback_issue`,
`post_issue_comment`, and `create_child_issue`, with the
uf-workflow factories supplying production dependencies, and the
Go drift and contract gates pinning the mirrors.

## Goals / Non-Goals

### Goals

- One tool call renders and writes the feedback-triage 1.1.0
  artifact (#1022): round numbering, FILE/REPLY routing from the
  `post_feedback_replies` structured result, summary counting,
  schema validation, atomic write.
- The terminal `clear_progress` moves into the tool (DE-005: the
  chain write→clear has no judgment in it).
- Section 4.7 shrinks to intent, the parameter mapping, and one
  rendered-output example (DE-002).
- Artifact bytes for the same inputs are unchanged — same
  envelope, same fields, same round derivation.

### Non-Goals

- Changing the feedback-triage schema (1.1.0 stays as written),
  its samples, or its consumers (Mx F, Muti-Mind).
- Changing Sections 4.5, 4.6, 1.2.1, or 4.4 — the gates, the
  reply call, and their contracts are settled.
- Reading filed issues back from the feedback cache. The cache
  is disposable by contract (Guardrail 6); the structured result
  of this run's `post_feedback_replies` call is the authoritative
  input.
- Emitting review-dispatch artifacts or review-verdict decisions
  — 4.7's artifact is emitted in addition to those, unchanged.

## Decisions

### D1: Extend `feedback-replies.ts`, not a new module

The operation joins the existing module's contract family: the
`feedbackSuccess`/`feedbackFailure` envelopes,
`FeedbackOperationError`, the input-schema-parity pattern, and
the item-key helper `itemKey(comment_kind, comment_id)` the
reply path already uses. A new lib module would duplicate those
helpers or export them cross-module, and would need new
asset-list entries in `scaffold.go` (two sites) and
`plugin_activation.go` plus the scaffold test lists — churn the
extension avoids because the module is already registered
everywhere. The module header is retitled to cover the artifact
write. Trade-off: the module grows by roughly 300 lines;
accepted, since it is one command's tool-owned mechanics.

### D2: Input contract — judgment products plus the replies result

`WriteFeedbackTriageArtifactInputSchema` (strict, `.superRefine`
for cross-field rules) carries:

- PR identity: `pr_number`, `pr_url`, `branch`.
- Artifact context: `commit` (HEAD SHA after fixes, or pre-fix
  HEAD — the command knows which), `backlog_item_id` (linked
  issue or `PR-N`).
- `items[]`: the 1.1.0 item fields (thread_id, reviewer,
  reviewer_role, file, line, classification, tier, evidence,
  recommendation, decision, decision_reasoning, commit_sha,
  divisor_agents_used, tier2_unavailable, conflict_flag) plus
  `comment_kind` and `comment_id` — the same identity Section
  4.6 already passes to `post_feedback_replies`. The identity is
  join input only: the 1.1.0 schema is
  `additionalProperties: false`, so the render omits it.
- `replies_result` (optional): the `post_feedback_replies`
  success data. Validated as a projection — an object whose
  `items[]` carry `key`, `comment_kind`, `comment_id`,
  `decision`, `issue_number`, `issue_url`; unknown fields are
  ignored so an evolved result still validates. Optional because
  the Section 4.5 "No -- skip posting" path skips 4.6 entirely
  and the artifact still records the triage decisions.

The join is mechanical and tool-owned (DE-006): each item is
matched by `itemKey` against the reply outcomes; a decision
mismatch fails `invalid_input` with nothing written. A `file`
item's `filed_issue_number`/`filed_issue_url` come from the
joined outcome (null when its issue failed — recording reality,
not an error); every other item renders null for both.

### D3: Round derivation is a directory scan

`readdir` on `.uf/artifacts/feedback-triage/`, matching
`^pr-(<N>)-round-(\d+)\.json$` against the same PR number;
round = max(M) + 1, default 1. Unparsable names are ignored; a
missing directory is round 1, not an error. Any other read
failure fails closed. This is the prose rule ("highest round
number and add 1 (not a file count — handles gaps from deleted
files)") as code.

### D4: Validation is a zod mirror, pinned to the samples

The rendered payload passes a zod validation mirroring
`schemas/feedback-triage/v1.1.0.schema.json`: required fields,
the `reviewer_role`/`classification`/`recommendation`/`decision`
enums, nullable fields, and `additionalProperties: false` at
item, summary, and payload level. A validation failure fails
closed with no file. Parity with the JSON Schema is pinned by a
vitest test that runs the checked-in samples through the same
validation — `samples/v1.1.0/sample-feedback-triage.json` and
the 1.0.0 sample pass; the `invalid-*` samples fail. The pin is
transitive (zod ↔ samples, samples ↔ JSON Schema via the Go CI
test in `internal/schemas`); adding Ajv for a direct check was
rejected as a new runtime dependency for one call site. The
envelope is validated by construction: its fields are constants
plus already-validated inputs.

Summary counting is derived, never passed: `total_items`,
`accepted`, `modified`, `rejected`, `asked`, `filed`, `replied`
count the decisions; `tier1_count`/`tier2_count` count the
tiers; `divisor_agents_invoked` is the deduplicated, lexically
sorted union of the items' `divisor_agents_used`. Items render
in input order — the order the command presents them, matching
Section 4.6's posting order.

### D5: Atomic write with the state-write precedent

The write follows `writeFeedbackState` and
`write_unleash_summary`: `mkdirSync` recursive `0o700`, symlink
refusal on the directory and the artifact path (`lstat`), a
`0o600` temp file (`${path}.${pid}.${Date.now()}.tmp`), then
`renameSync`. The path is derived from a validated integer
`pr_number` under the fixed artifact root, so traversal is
structurally impossible; the resolved path is still
prefix-checked against the root (belt and braces, summary-tool
precedent). Filesystem failures fail closed as `artifact_error`.

### D6: `clear_progress` moves into the tool

On a successful write the tool clears the `uf.address-feedback`
progress state through the same branch-keyed state path the
`clear_progress` tool uses. A clear failure is a warning —
`progress_cleared: false` plus the error in `warnings` — because
the artifact is already durable; this matches the
`feedback-reply-tools` D6 semantics for checkpoint writes. The
current prose clears only after the write, so behavior is
preserved exactly; the judgment-free trailing call just changes
owner (DE-005). `FeedbackReplyDeps` gains an optional
`clearProgress` hook (injectable for tests); the factory wires
the production state-path removal.

### D7: Error contract unchanged

Failures reuse the feedback tools' envelope: `status: "error"`,
`error: true`, `code`, `message`, `retryable` — `invalid_input`
for validation (including the decision cross-check), a new
`artifact_error` for artifact-directory and file I/O (round-scan
read failures, write failures, symlink refusal), and
`internal_error` for anything unexpected. Success returns
`pr_number`, `path`, `round`, `bytes`, the summary `counts`,
`progress_cleared`, `warnings`, and a rendered `summary` line
the command prints verbatim (DE-003).

### D8: Registration, mirrors, and contract tests

`createWriteFeedbackTriageArtifactTool(deps)` sits beside the
other feedback tool factories in uf-workflow with
`args: WriteFeedbackTriageArtifactInputSchema.shape`
(input-schema parity); the plugin entry registers
`write_feedback_triage_artifact`. The command, plugin, and lib
copies under `internal/scaffold/assets/` stay byte-identical —
no new asset-list entries (D1). The Go contract tests move their
4.7 pins to the tool wording: the `feedback triage artifact`
clause of `TestCommandContracts_AddressFeedback` and the
`checkpoint and clear` clause of
`TestCommandContracts_FeedbackReplyTools` re-pin to the call and
the tool-owned clear; the removed-text list gains the envelope
template and the hand-written round/atomic-write prose.

### D9: Section 4.7 becomes intent plus mapping

The rewritten section: build the items from the triage data with
their Section 4.6 comment identity, pass the
`post_feedback_replies` success data when it exists, call
`write_feedback_triage_artifact` once, print the returned
`summary` line verbatim, and STOP on an error envelope. The
no-success Tier 2 guidance (`tier2_unavailable: true`, the
`author-decides` compatibility labeling after human handling)
stays in prose — it is judgment about what to pass. The
Progress-Tracking `reply-comments` entry changes its trailing
note from "cleared after Phase 4.7" to "cleared by
`write_feedback_triage_artifact` after the artifact is written".
One rendered-output example (DE-002) shows the result shape.

## Risks / Trade-offs

- **Mirror drift is transitive, not direct.** The zod mirror is
  pinned to the samples, and the samples to the JSON Schema by
  the Go CI test — a schema change that skips the samples
  escapes the vitest pin. Accepted: that gap is already owned by
  the Go-side sample validation, and schema changes are
  spec-governed.
- **`replies_result` forward compatibility.** Validating only
  the consumed fields tolerates result evolution, but a renamed
  consumed field fails closed as `invalid_input` — visible, not
  silent. Preferred over strict whole-object validation, which
  would break the artifact on unrelated result additions.
- **The module grows** (~300 lines). Accepted per D1; the
  alternative (a second module) costs registration churn and
  splits one command's mechanics across files.
- **Tool-owned clear means an older plugin never clears.** If a
  deployment runs a plugin predating the tool, the command's
  tool call fails visibly and stops before any clear — the same
  failure class as every prior extraction; the state file is
  disposable, so a stuck progress record is recoverable by hand.
- **Artifact bytes are unchanged for the same inputs**, so no
  consumer can observe the writer change; the only observable
  difference is the command's terminal output (tool-rendered
  summary line) and the deterministic guarantee itself.
