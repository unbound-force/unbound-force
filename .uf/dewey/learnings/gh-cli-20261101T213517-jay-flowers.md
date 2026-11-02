---
tag: gh-cli
author: jay-flowers
category: gotcha
created_at: 2026-11-01T21:35:17Z
identity: gh-cli-20261101T213517-jay-flowers
tier: draft
---

gh validates requested `--json` field names against each subcommand's supported field set and fails with `Unknown JSON field` at invocation time, before emitting output — so a plugin whose expected schema drifts from the installed gh breaks at command time, not parse time. Concrete case (issue #895): `uf_finale_ci_poll` requests `gh pr checks --json name,state,conclusion,detailsUrl` and classifies pass/fail on REST-API-style conclusion values (SUCCESS/NEUTRAL/SKIPPED), but `gh pr checks` exposes `bucket`/`state`/`link` — the field set the host-side `fetch_pr_checks` tool already parses with fail-closed unknown-bucket handling. The vitest suite stays green because fixtures stub the exec with the drifted schema (golden-output-shape.test.ts even pins the wrong command string as its expected invocation), making the mismatch invisible until a live /uf.finale run. A second call site — review-dispatch's advisory ci_checks sub-call — requests the same nonexistent `conclusion` field but degrades silently to empty instead of failing. Fixes must thread the correction through every artifact layer that asserts the field: plugin code, the finale-token-reduction spec contract (`failedChecks: {name, conclusion, url}`), command docs, and golden output-shape tests — per learning/spec-review-20261029T060640-jay-flowers.
