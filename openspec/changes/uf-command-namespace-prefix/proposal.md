## Why

The 10 slash commands embedded in the `uf` binary and
scaffolded by `uf init` have no consistent namespace
prefix. Users cannot distinguish uf-managed commands
from project-specific or other-tool commands when
browsing the command list.

Other tools in the ecosystem already use clear
namespacing: `speckit.*` (9 commands), `muti-mind.*`
(12 commands), `opsx-*` (4 commands). The uf-owned
commands are the only group without a namespace,
making ownership ambiguous and the command list harder
to navigate.

Fixes: https://github.com/unbound-force/unbound-force/issues/302

## What Changes

All 10 uf-embedded slash commands are renamed from
their current unprefixed names to `uf.` dot-notation:

| Current | New |
|---------|-----|
| `/address-feedback` | `/uf.address-feedback` |
| `/agent-brief` | `/uf.agent-brief` |
| `/cobalt-crush` | `/uf.cobalt-crush` |
| `/constitution-check` | `/uf.constitution-check` |
| `/finale` | `/uf.finale` |
| `/review-council` | `/uf.review-council` |
| `/review-pr` | `/uf.review-pr` |
| `/triage-issue` | `/uf.triage-issue` |
| `/uf-init` | `/uf.init` |
| `/unleash` | `/uf.unleash` |

The scaffold engine gains a migration map to
automatically clean up orphaned old-name files when
`uf init` is re-run.

## Capabilities

### New Capabilities
- `migration-map`: Scaffold engine maps old command
  paths to new paths and removes orphans on re-run

### Modified Capabilities
- `uf-init-scaffold`: Embedded command assets use new
  `uf.*` filenames; `isToolOwned()` logic unchanged
  (already covers all `opencode/commands/*` paths)
- `doctor-hints`: InstallHint strings reference
  `/uf.agent-brief` instead of `/agent-brief`
- `scaffold-hints`: Divisor hint references
  `/uf.review-council` instead of `/review-council`

### Removed Capabilities
- None. Old command names cease to be scaffolded but
  orphan cleanup handles the transition.

## Impact

### Files renamed (20)
- 10 files in `.opencode/commands/`
- 10 files in `internal/scaffold/assets/opencode/commands/`

### Go source (hard references)
- `internal/scaffold/scaffold.go`: path check (line
  338), warning message (line 920), hint string
  (line 1588)
- `internal/scaffold/scaffold_test.go`:
  `expectedAssetPaths` (10 entries), 20+ individual
  test assertions
- `internal/doctor/checks.go`: 7 `InstallHint`
  strings
- `internal/doctor/doctor_test.go`: hint assertion

### Documentation (~30 files)
- `AGENTS.md`, `QUICKSTART.md`, `README.md`
- `docs/usage.md`, `docs/architecture.md`,
  `docs/heroes.md`
- Agent files, convention packs, skills
- Schema samples

### Cross-command references
- Commands that reference other uf commands by name
  (e.g., `/unleash` → `/uf.cobalt-crush`)
- Out-of-scope-owned files that cross-reference uf
  commands (e.g., `/opsx-propose` → `/uf.unleash`)

### Not impacted
- Hero identity `cobalt-crush` in Go orchestration
  code (hero name, not command name)
- Historical CHANGELOG entries
- Replicator, Gaze, OpenSpec, Speckit, Muti-Mind
  commands (different ownership)

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

This change renames command files (artifacts) without
altering their content or behavior. The artifact
envelope format, inter-hero communication, and
self-describing outputs are unaffected. The scaffold
engine's migration map is a build-time concern, not
a runtime coupling.

### II. Composability First

**Assessment**: PASS

Heroes remain independently installable. The `uf.`
prefix actually improves composability by making it
clear which commands come from `uf` vs. other tools,
reducing confusion when multiple tools are deployed
together. No hero gains a mandatory dependency on
another.

### III. Observable Quality

**Assessment**: N/A

This change does not alter any hero's output format,
provenance metadata, or machine-parseable artifacts.
It is a naming/organizational change only.

### IV. Testability

**Assessment**: PASS

All renamed commands remain testable in isolation.
The scaffold engine's migration map will be covered
by existing drift-detection tests (updated to use
new paths). The `TestCanonicalSources_AreEmbedded`
test continues to enforce that embedded assets match
canonical sources.

### V. Security by Default

**Assessment**: N/A

No dependencies added, no input validation changes,
no privilege changes. Pure naming refactor.
