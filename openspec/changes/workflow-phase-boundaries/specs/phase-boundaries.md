## ADDED Requirements

### Requirement: Phase Discipline Constitution Rule

The constitution's Development Workflow section MUST
include a Phase Discipline rule requiring each pipeline
phase to produce only its designated artifacts.
Implementation code MUST NOT be written during
specification phases.

#### Scenario: Agent runs /speckit.clarify

- **GIVEN** the agent is executing `/speckit.clarify`
- **WHEN** the agent identifies a spec update needed
- **THEN** the agent writes ONLY to spec.md and files
  within the feature directory (specs/NNN-*/)
- **AND** does NOT modify Go source, tests, agents,
  commands, or config files

---

### Requirement: AGENTS.md Workflow Phase Boundaries

AGENTS.md Behavioral Constraints MUST include a
Workflow Phase Boundaries subsection that maps each
pipeline phase to its allowed output:

- Specify/Clarify/Plan/Tasks/Analyze/Checklist →
  spec artifacts only
- Implement → source code changes
- Review → findings and minor fixes only

---

### Requirement: Code Review Marker in /unleash

The `/unleash` command MUST write a
`<!-- code-review: passed -->` marker to tasks.md after
the review council approves in Step 6 (code review).

The resumability detection (Step 2) MUST check for this
marker to determine if code review has been completed.

#### Scenario: /unleash resumes after implementation

- **GIVEN** all tasks are marked `[x]` in tasks.md
- **AND** `<!-- code-review: passed -->` is NOT present
- **WHEN** `/unleash` runs resumability detection
- **THEN** it resumes at Step 6 (code review), NOT
  Step 7 (retrospective)

#### Scenario: /unleash resumes after code review

- **GIVEN** all tasks are marked `[x]` in tasks.md
- **AND** `<!-- code-review: passed -->` IS present
- **WHEN** `/unleash` runs resumability detection
- **THEN** it skips to Step 7 (retrospective)

---

### Requirement: Externalize Speckit Commands

All 9 `speckit.*.md` command files MUST be removed from
`internal/scaffold/assets/opencode/command/`. They are
no longer deployed by `uf init`.

The 5 upstream commands (speckit.specify, speckit.plan,
speckit.tasks, speckit.implement, speckit.constitution)
are created by `specify init`.

The 4 UF-custom commands (speckit.analyze,
speckit.checklist, speckit.clarify,
speckit.taskstoissues) are created by `/uf-init`.

#### Scenario: uf init no longer deploys speckit commands

- **GIVEN** an engineer runs `uf init` in a fresh dir
- **WHEN** the scaffold engine processes embedded assets
- **THEN** no `speckit.*.md` files are deployed
- **AND** the 6 UF-custom commands (cobalt-crush,
  constitution-check, finale, review-council, uf-init,
  unleash) ARE still deployed

#### Scenario: specify init creates upstream commands

- **GIVEN** `specify init` runs (delegated by `uf init`)
- **WHEN** Speckit scaffolds its commands
- **THEN** 5 upstream `speckit.*.md` files are created
  in `.opencode/command/`

---

### Requirement: /uf-init Creates UF-Custom Commands

The `/uf-init` slash command MUST create the 4
UF-custom speckit commands if they do not exist:
- `speckit.analyze.md`
- `speckit.checklist.md`
- `speckit.clarify.md`
- `speckit.taskstoissues.md`

Creation is idempotent — if the file exists, skip it.

#### Scenario: /uf-init creates missing custom commands

- **GIVEN** `specify init` created 5 upstream commands
- **AND** the 4 UF-custom commands do not exist
- **WHEN** the engineer runs `/uf-init`
- **THEN** the 4 UF-custom command files are created

#### Scenario: /uf-init skips existing custom commands

- **GIVEN** all 4 UF-custom commands already exist
- **WHEN** the engineer runs `/uf-init`
- **THEN** the existing files are not overwritten

---

### Requirement: /uf-init Injects Guardrails

The `/uf-init` slash command MUST inject a
`## Guardrails` section into all 9 `speckit.*.md`
command files. The guardrails state that the command
may only write to files within the `specs/NNN-*/`
feature directory.

Injection is idempotent — if the `## Guardrails`
section already exists, skip it.

#### Scenario: /uf-init adds guardrails to new commands

- **GIVEN** speckit command files exist without a
  `## Guardrails` section
- **WHEN** the engineer runs `/uf-init`
- **THEN** a `## Guardrails` section is appended to
  each speckit command file

#### Scenario: /uf-init skips existing guardrails

- **GIVEN** speckit command files already have a
  `## Guardrails` section
- **WHEN** the engineer runs `/uf-init`
- **THEN** the existing guardrails are not duplicated

---

### Requirement: Stray File Cleanup and Prevention

All stray `.opencode/` and `.uf/` directories inside
`internal/scaffold/`, `cmd/unbound-force/`, and other
subdirectories MUST be deleted.

The `.gitignore` MUST include patterns that prevent
these directories from being tracked inside
subdirectories in the future.

#### Scenario: Stray directories cleaned up

- **GIVEN** stray `.opencode/` and `.uf/` directories
  exist inside `internal/scaffold/` and
  `cmd/unbound-force/`
- **WHEN** this change is applied
- **THEN** all stray files are deleted (74 files)

## MODIFIED Requirements

### Requirement: Scaffold Test Assertions

The `expectedAssetPaths` array MUST be updated to
remove the 9 externalized speckit command entries.
The `knownNonEmbeddedFiles` map MUST include all 9
speckit command filenames. The file count assertion
in `main_test.go` MUST be updated.

### Requirement: Scaffold Asset Synchronization

The modified command files (`unleash.md`, `uf-init.md`)
MUST have their scaffold asset copies synchronized.

## REMOVED Requirements

- 9 `speckit.*.md` files removed from scaffold assets.
