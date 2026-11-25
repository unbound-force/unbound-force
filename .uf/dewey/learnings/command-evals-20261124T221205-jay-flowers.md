---
tag: command-evals
author: jay-flowers
category: gotcha
created_at: 2026-11-24T22:12:05Z
identity: command-evals-20261124T221205-jay-flowers
tier: draft
as_of: "unbound-force@c9241969995f590083f3f56f0a0347168c4d56f9"
---

When a fix loop moves a tool call inside another tool (here acquire_intent_context into dispatch_review_runs), command prose that changes from 'call X' to 'do not call X' must be cross-checked against every .eval/fixtures/*/assertions.yaml that asserts tool_called or tool_output_matches on X. Eval fixtures are report-only and not part of make check, so the stale assertions stay green locally and only fail in nightly ci_eval. Grep .eval/fixtures for the moved tool name whenever a command contract retires a direct tool call.
