---
tag: gh-cli
author: jay-flowers
category: pattern
created_at: 2026-11-02T19:45:27Z
identity: gh-cli-20261102T194527-jay-flowers
tier: draft
---

When a spec or change re-sources an existing field from a new wire vocabulary (e.g. ci_checks `status ← gh state`, uppercase SUCCESS/PENDING, replacing fixtures that spoke lowercase CheckRun-style "completed"), enumerate every downstream comparison written against the old vocabulary before freezing the consumer side. A frozen output *shape* ("{name, status, conclusion} unchanged") is not a frozen value *vocabulary*: the spec's own scenario output (status: "SUCCESS") fed into the untouched consumer (`check.status !== "completed"`, review-dispatch/index.ts:8030) yields ci_pending=true on fully green PRs. Verified in opsx/fix-uf-finale-ci-poll-gh-schema spec review: the consumer and the fixtures (fetch-pr-review-context.test.ts:271-274) still speak the never-verified lowercase vocabulary. The companion failure mode: exit-code-intolerance facts verified for one capability (gh pr checks exits 8 pending / 1 failing, JSON still on stdout) must be applied to *every* call site of that command in the same change, or the stated goal ("eliminating the degradation") silently holds only for green PRs.
