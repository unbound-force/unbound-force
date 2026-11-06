---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-03T11:11:15Z
identity: test-quality-20261103T111115-jay-flowers
tier: draft
---

Aggregate statement-coverage floors cannot detect an uninvoked input mode. In PR #929 (bound-consolidate-dispatch-output), the new `detail: "summary"` input to consolidate_dispatch passed the 94% plugin coverage floor because the entire summary branch was simply never executed — the builder helpers were unit-tested directly, but no test invoked the tool boundary with the new input. When adding a new input mode to a tool command, the seam test (tool-level invocation with the new input, plus rejection of invalid values) is mandatory; coverage percentage is not a substitute for input-space coverage.
