---
tag: review-council
category: pattern
created_at: 2026-05-07T23:20:36Z
identity: review-council-1
tier: draft
---

The spec review council consistently identified a missing regression test requirement (TC-006) when reviewing a bug fix proposal. Three out of five reviewers independently flagged that the tasks.md only verified existing tests pass but did not include tasks to write new regression tests that exercise the exact failure scenario (bare \r input). The lesson: when proposing a bug fix, always include explicit regression test tasks in the initial proposal. The regression test should reproduce the original failure — for example, injecting strings.NewReader("y\r") as stdin to verify the prompt correctly handles carriage-return-only line endings. A test that passes with both old and new code (like strings.NewReader("y\n")) is not a regression test. Additionally, verify existing test coverage claims by actually searching the test files rather than assuming coverage exists.
