## Why

Issue #1022 (epic #1060, Release plan: v0.22.0, Tier 2) — the
follow-up that `feedback-reply-tools` (#1007) deferred: its design
listed "rendering the feedback-triage artifact in a tool" as a
Non-Goal and tracked the extraction here.

Section 4.7 of `.opencode/commands/uf.address-feedback.md` still
has the agent hand-write the feedback-triage artifact. The section
is ~86 lines, of which ~53 are a JSON envelope template the agent
copies and fills. The rest is mechanical procedure:

- **Round numbering** — scan `.uf/artifacts/feedback-triage/` for
  the highest `pr-<N>-round-<M>.json` and add 1 (an enumeration
  loop with gap tolerance).
- **FILE/REPLY routing** — take each FILE item's
  `filed_issue_number`/`filed_issue_url` from the
  `post_feedback_replies` structured result, count FILE items
  under `filed` and REPLY items under `replied` (mechanical
  substitution from a structured result the previous step
  already returned).
- **Summary counting** — total/accepted/modified/rejected/asked/
  filed/replied/tier1/tier2 and the `divisor_agents_invoked`
  union, all derived from the item list.
- **Atomic write** — temp file, then rename, with `600` files
  under `700` directories.
- **Progress cleanup** — after the artifact is written, call
  `clear_progress` (a fixed follow-on call with no judgment
  between it and the write).

Every step is a fixed procedure with the same output every time
(DE-001), and the chain from the `post_feedback_replies` result
through the write to `clear_progress` has no judgment in it
(DE-005). The council review of #1007 flagged exactly this.

## What Changes

- New `uf-workflow` tool `write_feedback_triage_artifact`. It
  takes the judgment products — the PR identity (`pr_number`,
  `pr_url`, `branch`), the artifact context (`commit`,
  `backlog_item_id`), the per-item triage records (each carrying
  the same `comment_kind`/`comment_id` identity Section 4.6
  already passes to `post_feedback_replies`), and, when replies
  were posted, the `post_feedback_replies` success data — and in
  one call it: joins FILE items to their filed issues by comment
  key and cross-checks the decisions, computes the summary
  counts, scans the next round number, renders the feedback-triage
  1.1.0 envelope, validates the rendered artifact against the
  1.1.0 payload shape, writes it atomically (temp file then
  rename, `0o600` files under `0o700` directories, symlink
  refusal), clears the `uf.address-feedback` progress state on
  success, and returns the path, round, byte count, and a
  rendered summary line.
- The operation lives in `.opencode/lib/feedback-replies.ts`
  (extended, not a new module — same envelope and error contract,
  same scaffold registration), with the tool factory and
  registration in `uf-workflow` beside `post_feedback_replies`.
- `/uf.address-feedback` Section 4.7 shrinks to intent and the
  parameter mapping: one `write_feedback_triage_artifact` call,
  one rendered-output example (DE-002), and the error-envelope
  STOP. The no-success Tier 2 guidance (`tier2_unavailable`, the
  `author-decides` compatibility labeling) stays in prose — it is
  judgment about what to pass, not how to write. The
  Progress-Tracking note for `reply-comments` moves its
  "cleared after Phase 4.7" wording to the artifact tool.
- The schema is unchanged: `schemas/feedback-triage/v1.1.0.schema.json`
  stays as written; the tool's payload validation is a zod mirror
  of it, pinned to the checked-in samples (valid samples pass,
  `invalid-*` samples fail).

## Capabilities

### New Capabilities

- `feedback-triage-artifact`: rendering and writing the
  feedback-triage 1.1.0 artifact for `/uf.address-feedback` —
  round numbering, FILE/REPLY routing from the
  `post_feedback_replies` result, summary counting, validation,
  atomic write, and terminal progress cleanup.

### Modified Capabilities

None in `openspec/specs`. This change supersedes the artifact
clauses FRT-FR-005 places on `/uf.address-feedback` emission
(the emission itself is unchanged; the writer moves from agent
prose to the tool). `feedback-reply-tools` is not archived, so
its spec is left as written and this note records the
supersession boundary: FRT-FR-005 governs the schema and the
emitted decisions; `feedback-triage-artifact` governs how the
artifact gets written.

### Removed Capabilities

None.

## Impact

- **Live plugin and lib**:
  `.opencode/lib/feedback-replies.ts` (new schemas, round scan,
  render, validation, atomic write, clear-progress hook,
  `writeFeedbackTriageArtifact`) and
  `.opencode/plugins/uf-workflow/index.ts` (factory +
  registration beside the other feedback tools).
- **Command**: `.opencode/commands/uf.address-feedback.md`
  Section 4.7 and the Progress-Tracking `reply-comments` note.
  Sections 4.5, 4.6, 1.2.1, and 4.4 are untouched.
- **Scaffold assets**: the command, plugin, and lib mirrors
  under `internal/scaffold/assets/` stay byte-identical
  (`TestEmbeddedAssets_MatchSource`,
  `TestCommandContracts_CanonicalScaffoldParity`). The lib module
  is already on the tool-owned lists in `scaffold.go` and
  `plugin_activation.go`, so no asset-list entries are added.
- **Tests**: `.opencode/test/feedback-replies.test.ts` (round
  scan with gaps, FILE/REPLY routing join and decision
  cross-check, absent replies result, counting, validation
  failure writes nothing, atomic write, permissions, symlink
  refusal, clear-progress semantics, envelopes);
  `golden-output-shape.test.ts` (documented output shape);
  `uf-workflow.test.ts` and `scratch-smoke.test.ts` (name
  lists); the zod mirror is pinned against
  `schemas/feedback-triage/samples/` (valid pass, `invalid-*`
  fail). DE-007 seam verification names the three layers.
- **Go contract tests**: `internal/scaffold/command_contract_test.go`
  — the `feedback triage artifact` clause of
  `TestCommandContracts_AddressFeedback` and the
  `checkpoint and clear` clause of
  `TestCommandContracts_FeedbackReplyTools` move to the
  tool-owned wording; removed-text checks gain the retired 4.7
  prose (the envelope template and the hand-written round/atomic
  steps).
- **Behavior change**: the artifact bytes are unchanged for the
  same inputs (same envelope, same fields, same round
  derivation); the path to them becomes deterministic, and
  progress cleanup now happens inside the tool with
  warning-not-failure semantics on clear failure (the artifact
  is already durable at that point).
- **Instruction delta (DE-004)**: Section 4.7 goes from 86 lines
  (including the 53-line envelope JSON template) to roughly 35;
  the exact token delta is measured with `instruction-lint
  --base` during implementation and recorded in the CHANGELOG
  entry.
- **Release tracking**: Tier 2 of the v0.22.0 release plan
  (#1060); must close before tagging.
- **CI replication**: `make plugin-test` (vitest suite, 94
  percent statements coverage floor) and
  `go test -race -count=1 ./internal/scaffold/...` (drift and
  contract gates).
- **Documentation impact**: CHANGELOG.md entry. The change is
  user-facing (the command's terminal step now reports a
  tool-rendered summary line), so the documentation gate
  applies: a documentation issue is filed in this repo before PR
  merge.

## Constitution Alignment

Assessed against the Unbound Force org constitution (v1.4.0 — six
active principles).

### I. Autonomous Collaboration

**Assessment**: PASS

The artifact was already the async contract between
`/uf.address-feedback` and its consumers (Mx F, Cobalt-Crush,
Muti-Mind); this change only changes its writer. The rendered
envelope keeps producer identity, version, timestamp, artifact
type, and schema version, so consumers interpret it without
consulting the producer. The tool's result carries the path,
round, and counts, so the calling agent reports outcomes from
structured data instead of re-deriving them.

### II. Composability First

**Assessment**: PASS

No new dependency and no cross-hero requirement. The operation
extends the existing standalone lib module and reuses its
envelope and error contract; the plugin remains usable
standalone. The zod mirror of the 1.1.0 schema adds no runtime
package (zod is already pinned).

### III. Observable Quality

**Assessment**: PASS

The artifact keeps its machine-parseable shape and provenance
metadata, now guaranteed by validation before write instead of
trusted to template copying. The tool returns a structured
result (path, round, bytes, counts, warnings) plus a rendered
summary line, matching the `post_feedback_replies` observable
contract. Failures use the existing structured error envelope
(`status: "error"`, `error: true`, `code`, `message`,
`retryable`).

### IV. Testability

**Assessment**: PASS

Every behavior is testable in isolation through the existing
vitest doubles: the round scan runs against a temp directory,
the FILE/REPLY join against fixture reply results, the
validation against the checked-in schema samples, and the write
against the filesystem with permission and symlink assertions.
No network, no `gh`, no external services.

### V. Security by Default

**Assessment**: PASS

No new trust boundary. The tool writes only inside
`.uf/artifacts/feedback-triage/` with the path derived from a
validated integer PR number, refuses symlinked targets, and
creates files `0o600` under directories `0o700`. Untrusted
comment text never reaches a shell — the artifact is written
through the filesystem, not `gh`.

### VI. Deterministic Execution

**Assessment**: PASS

This change is the determinism rule applied to itself: ~86
lines of mechanical instruction (DE-001) and a judgment-free
chain off a structured result (DE-005) move into a tested tool.
Round derivation, counting, joining, rendering, and the write
are pure functions of the inputs plus one directory scan;
`clear_progress` becomes a tool-owned constant follow-on.
Judgment products (triage records, context values) stay at the
tool boundary as validated parameters (DE-006).
