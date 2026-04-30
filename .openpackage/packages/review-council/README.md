# @unbound-force/review-council

Nine reviewer personas (**The Divisor**) audit code for security,
architecture, testing, operations, intent drift, and documentation
coverage. Ships OpenCode slash commands (`/review-council`,
`/review-pr`), shared convention packs (`default`, `severity`), and
optional Dewey MCP config.

## Install

```bash
opkg install @unbound-force/review-council
```

Or use **`uf init`** — when `opkg` is on your `PATH`, it installs this
package automatically (see `openspec/changes/opkg-delegate/proposal.md`).

## Personas

| Agent | Focus |
|:---|:---|
| divisor-guard | Intent drift, constitution, zero-waste |
| divisor-architect | Structure, boundaries, coupling |
| divisor-adversary | Security, abuse, resilience |
| divisor-sre | Operations, rollout, observability |
| divisor-testing | Tests, coverage, edge cases |
| divisor-scribe | Technical documentation |
| divisor-herald | Blog and announcements |
| divisor-envoy | PR and communications |
| divisor-curator | Docs gaps, website issues |

## Contents

- `agents/review-council/divisor-*.md` — OpenCode/Cursor/Claude metadata
- `commands/review-council/*.md` — slash commands
- `rules/review-council/*.md` — convention packs
- `mcp.jsonc` — Dewey MCP template (optional)
<!-- scaffolded by uf vdev -->
