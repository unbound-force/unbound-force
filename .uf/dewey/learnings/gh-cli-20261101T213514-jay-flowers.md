---
tag: gh-cli
author: jay-flowers
category: gotcha
created_at: 2026-11-01T21:35:14Z
identity: gh-cli-20261101T213514-jay-flowers
tier: draft
---

gh CLI `pr checks --json` exposes name/state/bucket/link only (gh 2.91.0); `conclusion` is not one of its fields, and requesting an unsupported --json field makes gh itself fail closed at request time with `Unknown JSON field: \"...\"` before any output is emitted — a wrong field name baked into a tool's gh invocation fails deterministically on every run, not intermittently. Issue #895 (uf_finale_ci_poll) shows the wrong expectation can originate upstream of the code: the finale-token-reduction spec codifies `failedChecks: {name, conclusion, url}` as the tool's output vocabulary and the uf.finale command prose repeats `conclusion`, so the unverified field name propagated from spec into the tool's gh field request. Treat field names a spec asserts about an external CLI's JSON as testable claims — verify them empirically at authoring time (the discipline already recorded for bucket vocabulary and the zero-checks shape), and when one proves wrong, thread the correction through spec, command prose, and tool together (per learning spec-review-20261029T060640). The repo already ships a schema-aligned consumer of the same command (fetch_pr_checks: name/state/bucket/link) to consolidate against.
