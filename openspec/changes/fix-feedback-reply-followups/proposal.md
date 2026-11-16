## Why

The v0.22.0 follow-up review of the feedback-reply tooling
(issue #1023, epic #1060, Tier 2) confirmed three reliability
defects in the unattended feedback and review-dispatch flows:

1. **Crash-recovery wording names the wrong subset** (MEDIUM).
   `.opencode/commands/uf.address-feedback.md` Section 1.7 says
   on crash-recovery re-entry to "pass items whose replies were
   already posted to `post_feedback_replies` again unchanged" —
   it names only the already-posted subset as what to pass.
   Followed literally, a re-run passes no pending items: the
   tool reads its per-item records, sees the passed items are
   already posted, does nothing, and records the
   reply-comments checkpoint with pending replies never
   posted. The tool itself is idempotent (its cache skips
   posted replies and resolved threads); only the instruction
   is wrong.

2. **gh exit code 4 is classified two ways** (MEDIUM).
   `classifyGhError` in `feedback-replies.ts` maps exit 4 to
   `gh_auth_failed` (not retryable), matching gh's documented
   convention ("if a command requires authentication, the exit
   code will be 4"). `ghFailureError` in
   `review-dispatch/index.ts` maps the same exit 4 to
   `rate_limited` (retryable), with a comment calling it gh's
   rate-limit exit code. One failure class, two consumers,
   opposite retry decisions, and nothing pins them together.

3. **`createIssue` retries on timeout and can file a duplicate
   issue** (LOW). The single retry in `createIssue` fires on
   any classified-retryable error, including `timeout`. A
   timeout is an ambiguous outcome — the issue may have been
   created server-side before the client gave up — so the
   retry can file the same issue twice. `createIssue` already
   declines to retry unparseable output for exactly this
   reason; timeout needs the same guard.

Fixes: #1023
Epic: unbound-force/unbound-force#1060 (Tier 2)

## What Changes

- **One shared gh failure classifier** — new module
  `.opencode/lib/gh-error.ts` exporting `classifyGhFailure`,
  the single decision table: signal and stderr-text keys first
  (missing binary, timeout, rate limit, authentication text),
  then gh's documented exit-code 4 as a fallback
  authentication signal (`gh_auth_failed`, not retryable).
  `classifyGhError` in `feedback-replies.ts` and
  `ghFailureError` in `review-dispatch/index.ts` become thin
  adapters over it, keeping their existing message templates.
  The review-dispatch side changes semantics for exactly one
  input class — exit 4 with no recognizable stderr text now
  reports `gh_auth_failed` / not retryable instead of
  `rate_limited` / retryable. Every existing review-dispatch
  test that pins exit 4 also pins rate-limit stderr text, so
  no pinned behavior changes.
- **`createIssue` no longer retries on timeout** — on a
  classified `timeout` the retry is skipped and the failure
  names the ambiguity (the issue may exist; check the repo
  before re-running), mirroring the existing
  unparseable-output guard. Other retryable failures (for
  example transient 5xx) still retry once.
- **Crash-recovery wording corrected** — the Section 1.7
  sentence in `.opencode/commands/uf.address-feedback.md` now
  says to pass the full confirmed item list again, unchanged,
  including items whose replies were already posted, and
  states what the tool's idempotence guarantees on a re-run.
- **Scaffold registration and mirrors** — `gh-error.ts` joins
  the scaffolded shared-lib module set (asset mirror,
  `isToolOwned` and `isDivisorAsset` lists,
  `stageReviewPluginSources`); every edited `.opencode/` file
  is mirrored byte-identically under
  `internal/scaffold/assets/opencode/`.

## Impact

- Affected files: `.opencode/lib/gh-error.ts` (new),
  `.opencode/lib/feedback-replies.ts`,
  `.opencode/plugins/review-dispatch/index.ts`,
  `.opencode/commands/uf.address-feedback.md`, their scaffold
  mirrors, the scaffold registration lists
  (`internal/scaffold/scaffold.go`,
  `internal/scaffold/plugin_activation.go`,
  `internal/scaffold/scaffold_test.go`), and tests
  (`.opencode/test/gh-error.test.ts` new;
  `feedback-replies.test.ts` and `fetch-pr-checks.test.ts`
  extended).
- No change to any tool's input schema, to retry counts for
  non-timeout failures, to quality gates, or to the four-phase
  address-feedback workflow.
- Out of scope by the epic's Tier-2 boundary: every other
  follow-up item in #1023 (the timeout-knob decision-table
  tests, the review LOWs, the "documented sites" count fix,
  and the IL-030/IL-031 budget reduction).
- Supersession boundary: `feedback-reply-tools` and
  `feedback-loop-scope` are unarchived changes, so their spec
  deltas are left as written. This change's FRT-FR-002-adjacent
  behavior (issue-creation failure handling) and AFS-FR-002
  clause-level overlap are recorded here rather than by
  rewriting those deltas.

## Constitution Alignment

Assessed against the Unbound Force org constitution (v1.4.0).

### I. Autonomous Collaboration

Unattended runs make their retry decisions from error
envelopes, not conversation. With one classifier, a
crash-recovered feedback run and a review-dispatch fetch tool
now report — and act on — the same gh failure identically, so
an operator reading either artifact gets one truth. The
crash-recovery fix keeps a recovered run completable without
human intervention: the tool's idempotence already exists; the
instruction now matches it.

### II. Composability First

The decision table moves from two private copies into one
shared lib module both plugins delegate to. Neither consumer
gains a dependency on the other's internals; the module stands
alone with its own test suite, and the scaffold registration
keeps a freshly scaffolded repo byte-compatible with this one.

### III. Observable Quality

Exit-4-without-text failures now carry the true classification
(`gh_auth_failed`, not retryable) in every tool's
machine-parseable error envelope, backed by a decision-table
regression test. The duplicate-issue guard surfaces the
ambiguity instead of hiding it behind a silent second attempt.

### IV. Testability

`classifyGhFailure` is a pure function over a signals record;
both adapters are covered by table tests that pin the mapping
through each consumer's public surface, and the new
`gh-error.test.ts` asserts adapter parity so the two consumers
cannot drift again.

### V. Security by Default

Untrusted text still reaches `gh` only through temporary files
and argument arrays; this change touches classification and
retry policy only. No new dependencies; no new network or
filesystem surface.

### VI. Deterministic Execution

The classification decision table — mechanical procedure —
moves from duplicated per-plugin prose-in-code to one
executable, tested module, which is exactly the tool-over-prose
direction the principle requires.
