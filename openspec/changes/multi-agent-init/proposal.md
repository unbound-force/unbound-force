## Why

`uf init` deploys all artifacts under `.opencode/`, which
only OpenCode reads. Teams using Claude Code or Cursor get
no benefit from the scaffold unless they manually discover
and reference the files. There is no mechanism to tell
non-OpenCode agents where convention packs and workflow
instructions live.

Additionally, repositories that already have framework
commands via a plugin (speckit, openspec) do not need uf
to deploy those same commands. A lighter deployment mode
that provides only convention packs — the unique value of
uf — would reduce onboarding friction.

## What Changes

### 1. `--packs-only` flag for `uf init`

A new filter mode lighter than `--divisor`. Deploys only
convention packs and custom stubs. No agents, no commands,
no schema directories.

**Implementation**: Add `isPacksOnlyAsset` predicate
function (similar to `isDivisorAsset`) that returns true
only for files under the convention pack path. When
`PacksOnly` is set in Options, skip all non-pack assets.

**What gets deployed with `--packs-only`:**

| File | Ownership |
|---|---|
| `.agents/packs/{lang}.md` | Tool-owned |
| `.agents/packs/{lang}-custom.md` | User-owned |
| `.agents/packs/default.md` | Tool-owned |
| `.agents/packs/default-custom.md` | User-owned |
| `.agents/packs/severity.md` | Tool-owned |

**What does NOT get deployed:**

- `.opencode/command/*` (commands)
- `.opencode/agents/*` (agent personas)
- `.opencode/skill/*` (skills)
- `openspec/schemas/*` (schema)

### 2. AGENTS.md pack reference section

After deploying packs, `uf init` appends a standardized
section to AGENTS.md (if the file exists) that tells all
agents — regardless of type — where to find convention
packs and workflow instructions.

**Proposed section:**

```markdown
## Convention Packs

This repository uses convention packs scaffolded by
unbound-force. Agents MUST read the applicable pack(s)
before writing or reviewing code.

- `.agents/packs/{lang}.md` — Language-specific rules
- `.agents/packs/{lang}-custom.md` — Project extensions
- `.agents/packs/default.md` — Language-agnostic rules
- `.agents/packs/severity.md` — Severity definitions

For spec-driven workflow commands, see
`.opencode/command/` (OpenCode) or follow the equivalent
instructions from AGENTS.md for other agents.
```

This section is user-owned (not overwritten on re-run)
and inserted only if AGENTS.md exists and does not already
contain a "Convention Packs" heading.

## Capabilities

### New Capabilities

- `--packs-only flag`: Deploys only convention packs
  without commands, agents, or schema.
- `AGENTS.md pack reference`: Automatically appended
  section telling all agents where packs live.

### Modified Capabilities

- `Options struct`: New `PacksOnly bool` field.
- `Run function`: Check `PacksOnly` before `DivisorOnly`.
  `PacksOnly` is the most restrictive filter.
- CLI `init` command: Accept `--packs-only` flag.

### Removed Capabilities

None.

## Impact

- No breaking changes. Existing `uf init` and
  `uf init --divisor` behavior is unchanged.
- `--packs-only` is additive — a new filter mode.
- AGENTS.md modification is conditional and idempotent.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

No change to artifact-based communication. Packs remain
self-describing Markdown files.

### II. Composability First

**Assessment**: PASS

This change directly serves composability. `--packs-only`
lets teams adopt convention packs without the review
council or swarm agents. Each capability is independently
installable.

### III. Observable Quality

**Assessment**: N/A

No change to output formats or quality metrics.

### IV. Testability

**Assessment**: PASS

`isPacksOnlyAsset` will have unit tests following the
existing pattern for `isDivisorAsset`.
