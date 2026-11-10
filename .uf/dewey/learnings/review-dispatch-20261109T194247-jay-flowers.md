---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-09T19:42:47Z
identity: review-dispatch-20261109T194247-jay-flowers
tier: draft
---

A docs-only diff (paths under docs/, plus CHANGELOG.md, QUICKSTART.md, README.md) contains no instruction files, so the divisor-linguist Code Review pass is vacuous by design: it produces no findings and approves. Dispatch planners can exclude divisor-linguist from docs-only review councils without losing instruction-quality coverage. Corollary: when the pre-flight instruction-lint advisory check fails on such a PR (it runs repo-wide, or unscoped on a merge-base fallback), those findings are pre-existing repo-wide lint debt anchored outside the diff — reviewers must scope linter evidence to the PR's changed files and not report it as PR-caused findings.
