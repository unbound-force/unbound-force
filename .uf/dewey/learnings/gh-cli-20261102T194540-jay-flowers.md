---
tag: gh-cli
author: jay-flowers
category: gotcha
created_at: 2026-11-02T19:45:40Z
identity: gh-cli-20261102T194540-jay-flowers
tier: draft
---

A field-list fix alone does not stop `gh pr checks` degradation at call sites that use a throwing runner (execSync or runGhCommand): gh exits 8 for pending and 1 for failing checks with the authoritative JSON array still on stdout, and throwing runners discard that stdout and treat the non-zero exit as failure. Every `gh pr checks` call site needs BOTH layers fixed — the verified field list (name,state,bucket,link) AND exit-code tolerance (stdout that parses as a JSON array is data on ANY exit code, per runGhCommandTolerant / the extend-fetch-pr-review-context pattern). In fix-uf-finale-ci-poll-gh-schema, the uf_finale_ci_poll poller got both layers (design D2) but the review-dispatch ci_checks sub-call got only the field swap, so every non-green PR would still degrade to empty ci_checks despite the design goal claiming full restoration — caught at spec review only because the delta's sole positive scenario was the green PR. When writing or reviewing a gh-schema fix: enumerate every call site and check its runner's exit-code behavior, not just its --json field list; and write the acceptance scenario for the non-green (exit 8/1) states, which is where the signal actually matters.
