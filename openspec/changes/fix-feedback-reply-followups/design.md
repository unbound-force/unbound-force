# Design: fix-feedback-reply-followups

## Context

Three scoped follow-up fixes from issue #1023 (epic #1060,
Tier 2): the uf.address-feedback crash-recovery wording, the
gh exit-code-4 classification divergence between
feedback-replies and review-dispatch, and the createIssue
timeout retry that can file a duplicate issue.

## D1 — Shared module, full scaffold registration

**Decision**: new module `.opencode/lib/gh-error.ts`; both
consumers delegate to it.

The issue offered two remedies ("Share one classifier, or pin
both to the same mapping in a test"). A shared module is
preferred because it removes the divergence class instead of
detecting it: one decision table, one test surface, one
authoritative ordering. Alternatives rejected:

- *In-place fix + cross-pinning test*: leaves two decision
  tables that can drift independently below the pin's
  granularity (any input class the pin does not cover).
- *Housing the classifier in `uf-workflow-types.ts`*: that
  module is types-only; review-dispatch would grow a runtime
  dependency on a types module.
- *Housing it in `feedback-replies.ts`*: would make
  review-dispatch import a feedback-domain module — wrong
  ownership direction.
- *`cli-common.ts`*: repo-local dev-CLI helper, deliberately
  not scaffold-registered; a plugin dependency on it would
  break scaffolded repos.

Registration cost is mechanical and precedented (16 existing
lib modules registered the same way): mirror file,
`isToolOwned` and `isDivisorAsset` lists in `scaffold.go`,
`stageReviewPluginSources` in `plugin_activation.go`, and the
scaffold test tables. The embed directive is whole-directory,
so the mirror file itself needs no registration.

## D2 — Decision table order and exit-4 semantics

`classifyGhFailure(signals)` evaluates in this order; the
first match wins:

| # | Signal | Code | Retryable |
|---|--------|------|-----------|
| 1 | spawn `ENOENT` | `gh_not_found` | no |
| 2 | spawn `ETIMEDOUT`, `SIGTERM`, or text "timed out"/"timeout" | `timeout` | yes |
| 3 | text "rate limit" | `rate_limited` | yes |
| 4 | text "gh auth login" / "not logged in" | `gh_auth_failed` | no |
| 5 | exit code 4 | `gh_auth_failed` | no |
| 6 | anything else | `api_error` | no |

- **Text keys outrank the exit code.** gh's own documentation
  (`gh help exit-codes`) defines exit 4 as "if a command
  requires authentication" — but `gh api` also exits 4 on
  other HTTP 4xx responses, and real rate-limit responses
  carry recognizable stderr. Classifying on stderr text first
  keeps those cases correct (a rate limit reporting exit 4
  still classifies `rate_limited`), with exit 4 as the
  fallback authentication signal when no text key matches.
- **Timeout is retryable in the fetch sense** (a re-run may
  succeed) but NOT for issue creation — see D3.
- This ordering preserves every currently-pinned behavior of
  `classifyGhError` and every review-dispatch test (all its
  exit-4 pins carry rate-limit stderr). The one semantic
  change is review-dispatch's exit-4-without-text case, which
  moves from `rate_limited`/retryable to `gh_auth_failed`/
  not-retryable — the documented gh convention.

**Adapters** keep their message templates so operator-visible
error strings are unchanged except where semantics change:
- `classifyGhError(err)` (feedback-replies) extracts
  `{stderr text, exitCode, spawnCode, signal}` from the caught
  record and maps codes to its existing messages.
- `ghFailureError(stderr, exitCode)` (review-dispatch)
  classifies, annotates the Error with `{ghCode, retryable}`,
  and keeps its message shapes; `classifyError` reads the
  annotations instead of re-parsing text. review-dispatch's
  prologue direct checks (`which gh`, `gh auth status`) are
  untouched.

## D3 — createIssue timeout guard

On a classified `timeout`, `createIssue` MUST NOT retry: the
issue may have been created server-side before the client gave
up, so a retry can file it twice. It throws immediately with a
message that names the ambiguity and directs the operator to
check the repo for the issue before re-running — mirroring the
existing unparseable-output guard, which does not retry for
the same reason. The thrown error carries `retryable: false`
because the remediation requires a human check, not a blind
re-run. Non-timeout retryable failures (for example transient
5xx → `api_error`) keep the single retry; that behavior is
pinned by an existing test.

## D4 — Crash-recovery wording

One sentence in Section 1.7 of `uf.address-feedback.md`
(line ~302) is replaced: the instruction now says to pass the
full confirmed item list again, unchanged — including items
whose replies were already posted — and states the tool's
idempotence guarantees (skips posted replies and resolved
threads; posts pending replies; resolves threads left open).
No other section changes; the file's existing IL-030/IL-031
budget warnings (out-of-scope item in #1023) remain
warnings — the added text is ~15 tokens.

## Testing strategy

- `.opencode/test/gh-error.test.ts` (new): decision-table
  tests for `classifyGhFailure` covering every row plus
  precedence (text beats exit code), and parity tests running
  the same inputs through `classifyGhError` and
  `ghFailureError` to pin the adapters to the shared table.
- `feedback-replies.test.ts`: extend the `classifyGhError`
  table with a text-timeout row and an exit-4-plus-rate-limit
  precedence row; add "does not retry gh issue create on a
  timeout" (exactly one gh invocation; message directs the
  operator to check for the issue).
- `fetch-pr-checks.test.ts`: pin review-dispatch's changed
  class — exit 4 with non-matching stderr (HTTP 422 text) now
  yields `gh_auth_failed`, retryable false. (Stderr choice
  avoids the tool's `no checks reported` and `not found`
  patterns.)
- Existing exit-4 pins carrying rate-limit stderr in
  `fetch-failed-run-logs.test.ts`,
  `fetch-pr-review-context.test.ts`, and
  `fetch-pr-checks.test.ts` MUST keep passing unchanged.
- Go side: `TestEmbeddedAssets_MatchSource` and the scaffold
  registration tables cover mirror identity and the new
  module's registration.
