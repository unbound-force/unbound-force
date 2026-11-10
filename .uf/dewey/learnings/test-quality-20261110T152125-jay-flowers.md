---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-10T15:21:25Z
identity: test-quality-20261110T152125-jay-flowers
tier: draft
---

A regression test for a "MUST NOT report X" clause needs a fixture where X sits in the state that would trigger the report. In unbound-force PR #1005 (merge-plugin-runtime-deps), the doctor test for "user-added runtime dependency is not drift" installed left-pad at exactly its declared version, so the test passed identically under the specified shipped-only comparison and under a widened all-declared comparison — the exact regression the MUST forbids. The check iterated a fixed shipped-package list and never read the user package's install, making the fixture write invisible to the code under test. Discriminating fixture: put the guarded entity in the offending state (declared 1.3.0 vs installed 1.0.0, or declared-but-absent) and assert the check still passes without naming it.
