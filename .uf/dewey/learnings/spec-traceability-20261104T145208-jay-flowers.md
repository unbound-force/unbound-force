---
tag: spec-traceability
author: jay-flowers
category: pattern
created_at: 2026-11-04T14:52:08Z
identity: spec-traceability-20261104T145208-jay-flowers
tier: draft
---

When a spec delta mandates exact output-message content (path quoting, guidance text), the verification obligation lands with the delta even when the emitting code is pre-existing — codifying existing behavior into MUST text amplifies the obligation, it does not inherit its test coverage. Quick traceability check: grep the test tree for a distinctive keyword from the mandated message; zero matches means the normative clause is unenforced and a later refactor can drop the text without failing any test. Anchor the new assertions in the test cases that already exercise the emitting code path, and when a shared unexported helper (e.g. diffPathFailure behind validateDiffPath, serving both build_review_prompts and read_diff_chunk) produces the message, each consuming tool's error-path test should assert the mandated content.
