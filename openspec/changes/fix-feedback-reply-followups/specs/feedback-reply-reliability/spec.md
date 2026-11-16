<!-- capability: feedback-reply-reliability
     change: fix-feedback-reply-followups
     requirement ids: FRR-FR-001 to FRR-FR-003 allocated here -->

## ADDED Requirements

### Requirement: One Shared gh Failure Classification

**ID: FRR-FR-001** This requirement MUST be satisfied.

The uf-workflow plugin's feedback tools and the
review-dispatch plugin's fetch tools MUST derive their gh
failure codes and retryability from one shared classifier,
`classifyGhFailure` in `.opencode/lib/gh-error.ts`. The
classifier MUST evaluate, in order, first-match-wins: a
missing gh binary (`gh_not_found`, not retryable); a timeout
signal or timeout text (`timeout`, retryable); rate-limit
text (`rate_limited`, retryable); authentication text
(`gh_auth_failed`, not retryable); gh exit code 4 as a
fallback authentication signal (`gh_auth_failed`, not
retryable); anything else (`api_error`, not retryable).
Stderr text keys MUST take precedence over exit code 4. Both
consumers MUST be pinned to the shared table by tests that
run identical inputs through each consumer's public surface.

#### Scenario: Text outranks exit code

- **GIVEN** a gh failure with stderr "HTTP 429: API rate limit
  exceeded" and exit code 4
- **WHEN** the failure is classified
- **THEN** the code MUST be `rate_limited`
- **AND** it MUST be retryable

#### Scenario: Exit code 4 without matching text

- **GIVEN** a gh failure with exit code 4 and stderr that
  matches no text key
- **WHEN** the failure is classified by either consumer
- **THEN** the code MUST be `gh_auth_failed`
- **AND** it MUST NOT be retryable

#### Scenario: Timeout classification

- **GIVEN** a gh failure whose stderr contains "timed out"
- **WHEN** the failure is classified
- **THEN** the code MUST be `timeout`
- **AND** it MUST be retryable

### Requirement: No Duplicate Issue on Ambiguous Outcome

**ID: FRR-FR-002** This requirement MUST be satisfied.

`createIssue` in `.opencode/lib/feedback-replies.ts` MUST NOT
retry `gh issue create` after a classified `timeout`: the
issue may have been created server-side before the client
gave up, and a retry could file it twice. On such a timeout
the call MUST fail immediately with an error whose message
directs the operator to check the repository for the issue
before re-running. Failures that are retryable and
unambiguous (for example transient server errors) MUST still
retry once.

#### Scenario: Timeout does not retry

- **GIVEN** a FILE item whose first `gh issue create` attempt
  times out
- **WHEN** `post_feedback_replies` files the issue
- **THEN** `gh issue create` MUST have been invoked exactly
  once
- **AND** the item's error message MUST direct the operator
  to check the repository for the issue before re-running

#### Scenario: Transient failure still retries

- **GIVEN** a FILE item whose first `gh issue create` attempt
  fails with a transient server error
- **WHEN** `post_feedback_replies` files the issue
- **THEN** `gh issue create` MUST have been invoked exactly
  twice

### Requirement: Crash-Recovery Re-Run Passes the Full List

**ID: FRR-FR-003** This requirement MUST be satisfied.

`.opencode/commands/uf.address-feedback.md` MUST instruct
that, on crash-recovery re-entry, the full confirmed item
list is passed to `post_feedback_replies` again, unchanged —
including items whose replies were already posted — and MUST
NOT instruct passing only the already-posted subset. The
instruction MUST state that the tool skips posted replies and
resolved threads from its per-item records, so a re-run
posts pending replies without duplicating comments, issues,
or thread resolutions.

#### Scenario: Instruction names the full list

- **GIVEN** the crash-recovery guidance in
  `.opencode/commands/uf.address-feedback.md`
- **WHEN** it is read
- **THEN** it MUST say to pass the full confirmed item list
  again, unchanged, including items whose replies were
  already posted

#### Scenario: Re-run completes pending work

- **GIVEN** a prior `post_feedback_replies` run posted some
  replies and crashed before the rest
- **WHEN** the full confirmed item list is passed again
- **THEN** the already-posted replies MUST NOT be posted
  again
- **AND** the pending replies MUST be posted
- **AND** threads a prior run left open MUST still be
  resolved
