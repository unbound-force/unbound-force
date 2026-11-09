## Why

`/uf.address-feedback` Sections 4.5 and 4.6 told the agent to
hand-run a fixed chain of `gh` calls between two human gates: write
each reply to a temp file, pick the REST endpoint, post with
`--input`, flip `comment-posted` flags in the feedback cache, file
FILE items' issues with shell and one retry, and, in a separate
phase, run the GraphQL `resolveReviewThread` mutation. Addressed
threads were only sometimes resolved (#966):

- Resolution sat in its own phase behind its own gate. A session
  end, crash, or declined gate between 4.5 and 4.6 left threads
  open, and gh-hud reads open threads as unresolved feedback.
- The progress list ended at `reply-comments` "(then clear)", so a
  resumed run never re-entered resolution.
- Phase 1.3 ingests REST endpoints only. They return no GraphQL
  thread node IDs, so the mutation depended on the agent improvising
  a lookup.

The chain has no judgment in it (DE-005), and every step is
mechanical (DE-001). Issue filing (#998) is the same class of
prose. Separately, the feedback-triage artifact records FILE and
REPLY items as `reject` with a `filed:` or `reply:` prefix, because
schema 1.0.0 has no such decisions (#999).

## What Changes

- New `uf-workflow` tool `post_feedback_replies` (#966). It takes
  the judgment products (PR number, per-item reply text, decision,
  target comment, the FILE item's issue title and body, and the
  author's resolve decision) and, in one call, files FILE items'
  issues, posts every reply, maps review comments to thread node
  IDs through a paginated GraphQL `reviewThreads` lookup, and
  resolves the approved threads. Each mutation is recorded per item
  in `.uf/feedback/pr-<N>/state.json` before the next starts, so a
  re-run with the same items resumes, including resolution. The
  tool records the `reply-comments` checkpoint when nothing is left
  to do and returns a rendered summary line.
- New `uf-workflow` tool `file_feedback_issue` (#998): files one
  finding with one retry and returns the issue number and URL; an
  issue already recorded for the same comment is returned instead
  of filing twice. `post_feedback_replies` files through the same
  path and cache record.
- Both tools pass untrusted text to `gh` only through `0o600` temp
  files and argument arrays, and fail with the fetch-tool error
  contract (`error: true`, `code`, `message`, `retryable`) alongside
  `status: "error"`.
- `/uf.address-feedback` Sections 4.5 and 4.6 shrink to intent and
  the two human gates. Both gates are asked before the single tool
  call. `clear_progress` moves after the Phase 4.7 artifact.
- New `schemas/feedback-triage/v1.1.0.schema.json` (#999): `file`
  and `reply` decisions, optional `filed_issue_number` and
  `filed_issue_url` item fields, and optional `filed` and `replied`
  summary counts. 1.0.0 stays for historical reads; 1.1.0 is a
  strict superset of it. The command emits 1.1.0.

## Capabilities

### New Capabilities

- `feedback-reply-tools`: posting replies, filing issues, and
  resolving threads for `/uf.address-feedback`, and the 1.1.0
  feedback-triage record of FILE and REPLY items.

### Modified Capabilities

None in `openspec/specs`. This change supersedes the artifact clause
of AFS-FR-002 in `feedback-loop-scope` (FILE and REPLY recorded as
`reject` with a prefix) and moves its filing mechanics into the
tool; that change is not archived, so its spec is left as written
and this note records the supersession.

## Impact

- `.opencode/lib/feedback-replies.ts` (new) and
  `.opencode/plugins/uf-workflow/index.ts`, with scaffold mirrors;
  `internal/scaffold` asset lists and drift tests gain the new lib
  module.
- `.opencode/commands/uf.address-feedback.md` and its scaffold
  mirror (Sections 1.7, 4.5, 4.6, 4.7, Progress Tracking,
  Guardrail 8). Sections 1.2.1 and 4.4 are untouched.
- `schemas/feedback-triage/` (schema, samples, README) and
  `internal/schemas/ci_test.go`.
- Consumers of feedback-triage artifacts (Mx F, Muti-Mind) see
  `schema_version: "1.1.0"` and the new decisions; readers pinned
  to 1.0.0 reject them rather than misread them.
- Instruction delta (DE-004, `instruction-lint --base`):
  `uf.address-feedback.md` 1510 → 1509 lines, ~17068 → ~17215
  tokens (+147). Removed: three shell blocks (30 lines of `mktemp`,
  `gh issue create`, `gh api --input`, and GraphQL mutation text)
  and the cache-flag and crash-recovery bookkeeping prose. Added:
  the tool's parameter mapping and two rendered-output examples
  (DE-002).
