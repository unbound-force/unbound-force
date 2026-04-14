## Context

The OpenSpec commands (`opsx-propose.md`,
`opsx-apply.md`, `opsx-explore.md`, `opsx-archive.md`)
are created by `openspec init --tools opencode`, not
by `uf init`. They are OpenSpec-owned files.

The `/uf-init` slash command applies UF-specific
customizations to files created by external tools.
The `opsx/workflow-phase-boundaries` change added
customization sections for Speckit commands but not
for OpenSpec commands.

## Goals / Non-Goals

### Goals

- Add a Guardrails section to `/opsx-propose` via
  `/uf-init` customization
- Prevent the agent from implementing code changes
  or running /finale during a /opsx-propose invocation

### Non-Goals

- No guardrails for `/opsx-apply` (it IS the
  implementation command — it should modify code)
- No guardrails for `/opsx-explore` (read-only by
  nature)
- No guardrails for `/opsx-archive` (cleanup command)

## Decisions

### D1: Apply via /uf-init, not direct edit

Since `opsx-propose.md` is owned by OpenSpec (created
by `openspec init`), we add the guardrails via a
`/uf-init` customization step — the same approach used
for Speckit command guardrails.

### D2: Guardrails content

The guardrails section states:

```markdown
## Guardrails

- **NEVER implement code changes** — this command
  creates artifacts ONLY (proposal, design, specs,
  tasks)
- **NEVER commit, push, or create PRs** — those are
  /finale's responsibility
- **NEVER run /opsx-apply or /cobalt-crush** — the
  user decides when to implement
- After artifacts are complete, STOP and prompt the
  user to run /opsx-apply or /cobalt-crush
```
