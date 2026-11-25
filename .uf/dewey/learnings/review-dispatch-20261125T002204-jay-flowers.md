---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-25T00:22:04Z
identity: review-dispatch-20261125T002204-jay-flowers
tier: draft
as_of: "unbound-force@c9241969995f590083f3f56f0a0347168c4d56f9"
---

When a new context-acquisition step precedes an aggregate dispatch tool, fold the acquisition into the dispatch tool instead of instructing callers through a call-then-forward chain: the dispatch runs the acquisition internally whenever the session-metadata field is absent, an explicitly supplied state always wins, and an acquisition failure degrades fail-open to a recorded warning (the review proceeds without the context). Keep the standalone acquisition tool registered only for paths that genuinely dispatch without the aggregate tool, and pin the collapsed orchestration end to end with a command-eval assertion that the standalone tool is never called in the command flow (tool_not_called). Verified in unbound-force intent-satisfaction-check: dispatch_review_runs internal acquisition (review finding F3, .opencode/plugins/review-dispatch/index.ts:8320-8350, schema at :6826-6844) with the eval fixture asserting tool_not_called: [acquire_intent_context] (.eval/fixtures/intent-satisfaction-check/assertions.yaml).
