## Why

Three related process gaps discovered during this session:

1. **Spec-phase code changes** (issue #92): During a
   `/speckit.clarify` run, the agent updated the spec
   correctly but then also modified Go source code and
   tests. No guardrail in the clarify command (or any
   other spec-phase command) prevents this. The spec
   pipeline phases (specify, clarify, plan, tasks,
   analyze, checklist) should only produce spec
   artifacts, never source code.

2. **Skipped code review** (issue #94): `/unleash`
   skips the code review step when all implementation
   tasks are already marked `[x]` on the first run.
   The resumability detection treats "all tasks
   complete + tests pass" as "everything done" and
   jumps to retrospective. Unlike spec review (which
   writes `<!-- spec-review: passed -->` to tasks.md),
   code review has no persistent marker, so its
   completion state is lost.

3. **Speckit commands embedded in uf**: The 9
   `speckit.*.md` command files are embedded as scaffold
   assets in the `uf` binary, but they belong to the
   Speckit domain. Spec 027 externalized Speckit
   scripts and templates to `specify init`, but the
   command files were left behind. This creates
   ownership confusion — who controls the command
   content? — and prevents the guardrails from being
   applied via `/uf-init` (the project-specific
   customization mechanism).

Additionally, 74 stray files exist in subdirectories
(`internal/scaffold/.opencode/`, `cmd/unbound-force/.uf/`,
etc.) from tools running in the wrong working directory.
These need cleanup and prevention.

## What Changes

### New Capabilities

- **Phase Discipline rule**: New constitutional rule
  in Development Workflow requiring each pipeline phase
  to produce only its designated artifacts.
- **Workflow Phase Boundaries**: New AGENTS.md
  behavioral constraint enumerating which phases may
  write source code and which may not.
- **Code review marker**: `/unleash` writes
  `<!-- code-review: passed -->` to tasks.md after the
  review council approves, parallel to the existing
  spec-review marker.
- **`/uf-init` guardrail injection**: The `/uf-init`
  slash command injects a `## Guardrails` section into
  all 9 speckit command files after `specify init`
  creates them. AI-assisted, idempotent (checks for
  existing section before appending).
- **`/uf-init` custom command creation**: The `/uf-init`
  slash command creates the 4 UF-custom speckit commands
  (analyze, checklist, clarify, taskstoissues) that
  upstream `specify init` does not provide.

### Modified Capabilities

- **`/unleash` command**: Add code-review marker write
  after Step 6 approval. Update resumability detection
  (Step 2) to check for code-review marker.
- **`/uf-init` command**: Add speckit command
  customization sections (guardrail injection + custom
  command creation).
- **Scaffold engine**: Remove 9 `speckit.*.md` files
  from embedded assets. Add to `knownNonEmbeddedFiles`.

### Removed Capabilities

- 9 `speckit.*.md` files removed from scaffold assets
  (externalized to `specify init` + `/uf-init`).
- 74 stray files deleted from subdirectories.

## Impact

- `.specify/memory/constitution.md` — 1 new bullet
- `AGENTS.md` — 1 new subsection + Recent Changes
- `.opencode/command/unleash.md` — code-review marker
- `.opencode/command/uf-init.md` — guardrail injection
  + custom command creation sections
- `internal/scaffold/assets/opencode/command/` — remove
  9 speckit files, sync 2 modified files (unleash,
  uf-init)
- `internal/scaffold/scaffold_test.go` — update
  expectedAssetPaths, knownNonEmbeddedFiles
- `cmd/unbound-force/main_test.go` — update file count
- `.gitignore` — stray prevention patterns
- Delete 74 stray files from 4 subdirectories

## Constitution Alignment

### I. Autonomous Collaboration

**Assessment**: PASS

Phase boundaries strengthen artifact integrity. Each
phase produces well-defined outputs that downstream
phases consume. Externalizing speckit commands to their
owning tool (specify) reinforces autonomous ownership.

### II. Composability First

**Assessment**: PASS

Removing speckit commands from uf's scaffold assets
makes uf less coupled to Speckit. Each tool owns its
own command files. `/uf-init` applies project-specific
customizations on top — composable layering.

### III. Observable Quality

**Assessment**: PASS

The code-review marker makes review completion
observable and persistent. Without it, code review
state is inferred from CI pass/fail — which is
unreliable (no CI, tests flaky, etc.).

### IV. Testability

**Assessment**: PASS

Updated scaffold tests verify the externalization is
correct (expectedAssetPaths reduced, knownNonEmbeddedFiles
expanded). Drift detection catches any re-introduction.
