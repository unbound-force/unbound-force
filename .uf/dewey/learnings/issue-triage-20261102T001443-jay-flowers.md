---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:14:43Z
identity: issue-triage-20261102T001443-jay-flowers
tier: draft
---

CHANGELOG entries that name their unit-test surface — the test file plus the enumerated covered behaviors — make a change's regression-protection claims verifiable from the change record alone. The rebase-upstream-before-push entry (unbound-force/unbound-force CHANGELOG.md, tracked by issue #810) lists `.opencode/test/uf-finale.test.ts` covering upstream remote preference, origin-only fallback, symbolic-ref failure falling back to `main`, the offline `BASE_SYNC_FETCH` envelope, the `BASE_SYNC_COUNT` envelope, and ahead/behind counts; each enumerated behavior maps 1:1 to an actual test in the file's `describe("uf_finale_base_state")` block (verified by grep during triage). Triagers and reviewers can confirm coverage claims with one targeted grep of the named file, and any gap between claimed and actual tests surfaces immediately. Authoring rule: when a change record summarizes new tool behavior, name the test file and the behaviors it pins.
