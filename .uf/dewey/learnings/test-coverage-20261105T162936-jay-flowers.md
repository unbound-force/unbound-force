---
tag: test-coverage
author: jay-flowers
category: pattern
created_at: 2026-11-05T16:29:36Z
identity: test-coverage-20261105T162936-jay-flowers
tier: draft
---

A tasks.md entry that enumerates test cases by name can under-count the outcomes its own design decisions specify. Before marking such a task complete, diff the enumeration against every fail/warn/pass branch the design names. In stop-shipping-plugin-lock (PR 957), task 3.3 listed five doctor test cases while design D4 specified six behaviors (mismatch, unreadable install, undeclared, missing manifest, malformed manifest, not-installed deferral) — the malformed-manifest Warn became the rewritten check's sole untested outcome, caught only by maintainer feedback. Enumeration-by-name in tasks is convenient for reviewers but silently drops branches unless cross-checked against the design's outcome list.
