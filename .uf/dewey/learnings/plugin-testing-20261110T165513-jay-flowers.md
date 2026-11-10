---
tag: plugin-testing
author: jay-flowers
category: pattern
created_at: 2026-11-10T16:55:13Z
identity: plugin-testing-20261110T165513-jay-flowers
tier: draft
---

Embed security-relevant contract assertions inside the test double itself so every test using the double verifies them on every call. In unbound-force PR 1007, .opencode/test/feedback-replies.test.ts's fakeGh executor asserts the temp-file permissions (0o600 file under 0o700 directory) and post-call directory removal on every gh invocation that carries --input/--body-file, so all 61 tests in the suite double as injection-safety and temp-file-lifecycle regression tests without repeating those assertions per test.
