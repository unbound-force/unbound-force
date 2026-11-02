---
tag: gh-cli-contracts
author: jay-flowers
category: gotcha
created_at: 2026-11-01T21:35:25Z
identity: gh-cli-contracts-20261101T213525-jay-flowers
tier: draft
---

uf_finale_ci_poll (unbound-force/unbound-force issue #895) failed at runtime with 'Unknown JSON field: \"conclusion\"' while its full test suite passed, because the tool invoked `gh pr checks --json name,state,conclusion,detailsUrl` — a field set that subcommand never exposed. `gh pr checks --json` emits name/state/bucket/link; `conclusion` belongs to `gh pr view --json statusCheckRollup` and `gh run view --json` check runs. The plugin's unit and golden-shape tests fabricated `conclusion`/`detailsUrl` in their gh mocks, so mock-based tests pinned the assumed schema instead of the real one — CI stayed green (94% coverage floor included) while the tool failed on every poll iteration and /uf.finale's CI-watch step degraded to manual polling. Rule: when a tool shells out to an external CLI and parses its `--json` output, the mock must be derived from that exact subcommand's documented field set (or a captured real response), not from a sibling command's schema or from memory; where feasible, add a smoke test that executes the real command with the exact field list so field-name drift fails in CI instead of production. Related prior learning: spec-review-20261029T060640-jay-flowers (gh pr view exposes no baseRepository field — same field-existence gotcha class).
