---
tag: gh-cli
author: jay-flowers
category: gotcha
created_at: 2026-11-02T23:07:40Z
identity: gh-cli-20261102T230740-jay-flowers
tier: draft
---

gh CLI's --json flag validates the requested field list against the installed version's schema and aborts the whole command (non-zero exit, no data emitted) when a requested field does not exist: gh 2.91.0 (2026-04-22) `gh pr checks --json` supports only bucket, completedAt, description, event, link, name, startedAt, state, and workflow — requesting conclusion or detailsUrl fails with `Unknown JSON field: "conclusion"`. Compounding this, gh's exit codes carry state, not just success: pending checks exit 8, failing checks exit 1, and a checkless PR exits 1 with EMPTY stdout and "no checks reported" on stderr (gh never emits `[]`), so a throwing execSync consumer breaks on any non-green PR even with correct field names. Any plugin that shells out to gh MUST pin its --json field list to the verified schema of the deployed CLI and pin the exact argv in golden tests, so future schema drift fails loudly in the suite instead of silently in production. Discovered fixing unbound-force issue #895 (uf_finale_ci_poll, branch opsx/fix-uf-finale-ci-poll-gh-schema, 2026-11-02).
