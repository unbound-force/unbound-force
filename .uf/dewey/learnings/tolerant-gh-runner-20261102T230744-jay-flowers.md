---
tag: tolerant-gh-runner
author: jay-flowers
category: pattern
created_at: 2026-11-02T23:07:44Z
identity: tolerant-gh-runner-20261102T230744-jay-flowers
tier: draft
---

The durable pattern for consuming external-CLI JSON in OpenCode plugins (established by review-dispatch's runGhCommandTolerant and replicated for uf_finale_ci_poll on branch opsx/fix-uf-finale-ci-poll-gh-schema, 2026-11-02): invoke through an argument-array Bun.spawn runner with a per-invocation AbortSignal timeout that resolves {stdout, stderr, exitCode} without throwing; treat stdout that parses as a JSON array as authoritative data on ANY exit code; detect terminal shapes via narrow stderr markers pinned to the verified CLI version (gh's "no checks reported" for checkless PRs; "could not resolve to a PullRequest" for unresolvable PRs — deliberately NOT the looser prior-art alternation /not found|no pull request|could not find|.../, which misclassifies unrelated not-found stderrs such as HTTP 404 resource errors as non-retryable PR_NOT_FOUND and suppresses poll-loop retries); classify from the CLI's authoritative vocabulary field (bucket) with unknown values fail-closed as failed and surfaced verbatim so vocabulary drift is visible instead of burning a poll timeout; and preserve the tool's output contract by mapping wire fields onto the existing shape (conclusion sourced from state, url from link) so downstream consumers and golden-shape tests stay untouched. When this helper family is copied module-locally between plugins, both copies carry twin-pointer JSDoc naming each other with an extraction trigger ("a third consumer MUST trigger extraction into .opencode/lib/"), and the scaffold twins under internal/scaffold/assets/opencode/plugins/ are synced byte-identical (verified with diff).
