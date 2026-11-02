---
tag: gate-automation
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:17:54Z
identity: gate-automation-20261102T041754-jay-flowers
tier: draft
---

Automation-wiring issues for gate commands must verify observability contracts in the consumer's context, not only the command's own test suite. unbound-force issue #584 (wiring `uf gate` as FullSend `validation_loop.script` and documenting pre-commit/pre-push hook usage) explicitly includes "verify exit code semantics work correctly in the harness loop" and "test JSON output parsing by the harness" — acceptance criteria that exercise the command's exit-code taxonomy (0 pass / 1 check failure / 2 internal error, per PR #582) and machine-parseable JSON output (Constitution Principle III, Observable Quality) at the integration point where the gate actually executes. A wiring issue that only adds harness config or hook documentation without consumer-side verification of exit codes and structured-output parsing leaves the gate's behavior in CI and hook contexts unverified. This complements learning/issue-triage-20261101T222235-jay-flowers (security-control completion framing) with the operational-readiness criterion: consumer-context verification IS the acceptance bar for wiring work.
