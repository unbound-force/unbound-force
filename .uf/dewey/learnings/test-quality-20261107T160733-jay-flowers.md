---
tag: test-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-07T16:07:33Z
identity: test-quality-20261107T160733-jay-flowers
tier: draft
---

When a CLI offers a report-scoping flag (e.g. --only PATH) over whole-repository analysis, pin the no-false-positives contract with a fixture whose scoped file references a VALID artifact that lives OUTSIDE the scope. If every reference in the scoped fixture points to genuinely missing targets, a regression that filters discovery/artifacts before analysis (instead of filtering findings after) produces byte-identical output and the suite stays green — while every scoped consumer (review personas linting changed files, preflight checks) starts emitting false missing-reference errors at full severity. Symptom to audit for: a test whose name claims "while analyzing the whole repo" (or equivalent) but whose assertions cannot distinguish whole-repo analysis from scoped discovery. Concrete pin: create an out-of-scope real skill/command, reference it from the scoped file, and assert no reference-error finding fires for it (or assert the scoped file's load-set token total includes the out-of-scope artifact).
