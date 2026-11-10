## Context

`/uf.address-feedback` posts replies after the author confirms
them (Section 4.5 gate) and offers to resolve addressed threads
(Section 4.6 gate). Between and after those gates the command
carried shell choreography the agent re-derived every run:
temp-file handling, endpoint selection, `gh issue create` with one
retry, cache-flag bookkeeping, and a GraphQL mutation whose
`threadId` input the command never captured. Precedent for the fix
is the intent-level tool: `uf_delegate_command`,
`dispatch_review_runs`, and `uf_finale_create_pr` for the temp-file
lifecycle.

## Goals / Non-Goals

**Goals:**

- One tool call posts all confirmed replies and resolves the
  approved threads (#966).
- Thread node IDs are resolved inside the tool.
- A re-run after a partial failure resumes per item: no duplicate
  comment or issue, and threads a prior run left open get resolved.
- Issue filing is a tool with one retry and a cache duplicate check
  (#998).
- The artifact records FILE and REPLY as decisions (#999).

**Non-Goals:**

- Changing Phase 1.3 ingestion or the Section 1.2.1 PR-head
  checkout (a concurrent change owns 1.2.1).
- Rendering the feedback-triage artifact in a tool. It stays
  agent-written in 4.7; extraction is tracked in
  unbound-force/unbound-force#1022.
- Duplicate detection against GitHub issue search. The duplicate
  check is the feedback cache, as #998 specifies.

## Decisions

### D1: One tool for file, post, and resolve

FILE replies need the issue number, which is mechanical
substitution, so filing inside the reply tool removes the last
judgment-free hop between gates. `post_feedback_replies` takes, per
item, `comment_kind` (`review` or `issue`), `comment_id`,
`decision`, the confirmed `body`, and for `file` items an
`issue: {title, body}`; plus `pr_number`, optional `repo`, and
`resolve_threads`. For a `file` item it renders the reply as
`Filed as #<n>: <body>`, so the reply-content template for FILE is
tool-rendered (DE-003).

`file_feedback_issue` is the standalone #998 surface with the same
cache record, so a finding filed on its own is reused by a later
reply batch (a write→read seam).

### D2: Both gates before the call

The resolve gate (4.6) moves before posting, so its answer is a
parameter (`resolve_threads`) and the human decisions stay at the
tool boundary (DE-006). The option labels, gate count, and
per-gate contract fragments are unchanged.

### D3: Ordering, retry, and failure-stop as constants

- Items run in input order. A FILE item files first; its issue is
  persisted before the reply is posted.
- `gh issue create` gets one retry. Output with no issue URL is not
  retried (the issue may exist). A FILE item whose issue fails is
  not posted or resolved, and the next item proceeds, as AFS-FR-002
  requires.
- A failed post stops further posting: an outage or rate limit
  would fail the rest too. Remaining items report `pending`.
- Resolution runs for every addressed review item (accept, modify,
  file, reply) whose reply is posted, including replies posted by
  an earlier run. `resolveReviewThread` is idempotent, so it gets
  one retry on retryable errors. A failed resolution does not stop
  the others.
- Posting is not retried in-call: a timed-out POST may have landed,
  and the cache cannot tell. The re-run is the retry.

### D4: Thread IDs from a paginated reviewThreads lookup

One GraphQL query per page (100 threads, up to 50 pages) maps every
comment `databaseId` in a thread to the thread's node ID and
`isResolved`. Already-resolved threads are recorded without a
mutation. A comment with no thread reports `thread_not_found`.

### D5: Tool-owned cache record

The tool reads and writes `.uf/feedback/pr-<N>/state.json`,
preserving every key the command writes and owning only
`replies["<kind>:<id>"]`: `comment_posted`, `reply_id`,
`reply_url`, `issue_number`, `issue_url`, `thread_id`,
`thread_resolved`, `decision`, `updated_at`. Writes are atomic
(temp file and rename), files `0o600`, directories `0o700`, and
symlinked targets are refused. A malformed cache fails the call
with `state_error` (the cache is disposable: delete and re-run).

### D6: Checkpoint by the tool

The tool records `reply-comments` through the same
`recordAutoCheckpoint` path the checkpoint tool uses, and only when
every reply is posted and every approved thread resolved. An
incomplete run records nothing, so a resumed run re-enters Section
4.6. `clear_progress` moves after the 4.7 artifact. A checkpoint
write failure is a warning in the result, since the GitHub
mutations already happened.

### D7: Error contract

Whole-call failures return
`{status: "error", error: true, code, message, retryable}`: the
uf-workflow discriminant plus the fetch-tool contract, with the
fetch-tool codes (`gh_not_found`, `gh_auth_failed`, `rate_limited`,
`timeout`, `api_error`) and `invalid_input`, `repo_unresolved`,
`state_error`, `internal_error`. Per-item failures sit in the
success payload as `items[].error` with the same three fields.

### D8: Lib module plus factories

The chain lives in `.opencode/lib/feedback-replies.ts` (schemas,
pure helpers, and the two operations, with an injectable gh
executor); `uf-workflow` registers the factories and supplies the
checkpoint recorder. The tool `args` are the validation schemas'
shapes, so the model-facing schema and the execute-path validation
cannot drift. The lib module ships as a tool-owned scaffold asset.

### D9: Schema 1.1.0 as a strict superset

`v1.1.0.schema.json` adds `file` and `reply` to the decision enum,
optional `filed_issue_number` and `filed_issue_url`, and optional
`filed` and `replied` summary counts. Every 1.0.0 artifact
validates under 1.1.0. A reader pinned to 1.0.0 rejects the new
decisions instead of misreading them. Samples live in
`samples/v1.1.0/`, following the review-verdict v2 layout.

## Risks / Trade-offs

- The resolve question is now asked before posting, so the author
  decides resolution without seeing posting results. A failed post
  leaves its thread open regardless, so nothing is resolved without
  a posted reply.
- A POST that succeeded server-side but failed client-side is
  re-posted on the re-run. The previous prose had the same window.
- The command still writes the artifact by hand (Non-Goal); its
  FILE fields now come straight from the tool's result.
