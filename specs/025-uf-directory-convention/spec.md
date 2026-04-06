# Feature Specification: Unified .uf/ Directory Convention

**Feature Branch**: `025-uf-directory-convention`  
**Created**: 2026-04-06  
**Status**: Draft  
**Input**: User description: "Consolidate all per-repo tool directories under .uf/ and rename .opencode/unbound/ to .opencode/uf/"

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Project Initialization Creates .uf/ (Priority: P1)

When an engineer runs `uf init` in a project, all
per-repo tool data is organized under a single `.uf/`
directory instead of scattered across `.dewey/`,
`.hive/`, `.unbound-force/`, `.muti-mind/`, and `.mx-f/`.
Today, each tool creates its own dotfile directory at
the repo root, cluttering the project with 5 separate
hidden directories. After this change, `uf init` creates
a single `.uf/` directory with subdirectories for each
tool, making project structure cleaner and tool data
discoverable in one location.

**Why this priority**: Init is the per-repo entry point.
Every project encounters this on first setup. A clean
directory structure reduces confusion for new
contributors who see multiple mysterious dotfiles.

**Independent Test**: Run `uf init` in a fresh directory
with Dewey and Replicator installed. Verify `.uf/` is
created with `dewey/`, `replicator/`, `config.yaml`
subdirectories. Verify no `.dewey/`, `.hive/`, or
`.unbound-force/` directories are created.

**Acceptance Scenarios**:

1. **Given** a fresh project directory with no `.uf/`,
   **When** the engineer runs `uf init`, **Then** the
   scaffold creates `.uf/` with `config.yaml` and
   delegates to `dewey init` and `replicator init` which
   create `.uf/dewey/` and `.uf/replicator/` respectively.
2. **Given** `.uf/` already exists with correct
   subdirectories, **When** the engineer runs `uf init`,
   **Then** existing directories are preserved and the
   result reports "already configured."
3. **Given** old directories (`.dewey/`, `.hive/`,
   `.unbound-force/`) exist, **When** the engineer runs
   `uf init`, **Then** old directories are ignored (not
   detected, not migrated, not deleted) and `.uf/` is
   created fresh.
4. **Given** the Dewey or Replicator binary is not in
   PATH, **When** the engineer runs `uf init`, **Then**
   the corresponding `.uf/` subdirectory is not created
   and the result reports the tool is not available.

---

### User Story 2 — Health Checks Use .uf/ Paths (Priority: P1)

When an engineer runs `uf doctor`, all health checks
reference `.uf/` paths instead of the old scattered
directories. Today, doctor checks `.dewey/` for Dewey
workspace, `.hive/` for Replicator cells, and
`.unbound-force/` for workflow config. After this change,
doctor checks `.uf/dewey/`, `.uf/replicator/`, and
`.uf/config.yaml` respectively.

**Why this priority**: Doctor is how engineers diagnose
their environment. If doctor checks the wrong paths,
it produces misleading diagnostics.

**Independent Test**: Run `uf doctor` in a project with
`.uf/dewey/` and `.uf/replicator/` present. Verify all
checks pass and reference `.uf/` paths in the output.
Verify no references to `.dewey/` or `.hive/` appear.

**Acceptance Scenarios**:

1. **Given** `.uf/dewey/` and `.uf/replicator/` exist,
   **When** the engineer runs `uf doctor`, **Then**
   the Dewey workspace check and Replicator `.hive/`
   check both report PASS using `.uf/` paths.
2. **Given** `.uf/replicator/` does not exist, **When**
   the engineer runs `uf doctor`, **Then** the check
   reports WARN with hint `Run: uf init`.
3. **Given** `uf doctor` runs with `--format=json`,
   **When** the output is parsed, **Then** all path
   references use `.uf/` (not `.dewey/` or `.hive/`).

---

### User Story 3 — Convention Packs at .opencode/uf/packs/ (Priority: P1)

When the scaffold engine deploys convention packs via
`uf init`, they are placed at `.opencode/uf/packs/`
instead of `.opencode/unbound/packs/`. Today, 9 pack
files live at `.opencode/unbound/packs/`. After this
change, the same 9 files live at `.opencode/uf/packs/`.
All agent files that reference pack paths are updated.

**Why this priority**: Convention packs are loaded by
every agent at the start of every task. If the path is
wrong, agents cannot find their coding standards.

**Independent Test**: Run `uf init` in a fresh directory.
Verify `.opencode/uf/packs/go.md` exists. Verify no
`.opencode/unbound/` directory is created.

**Acceptance Scenarios**:

1. **Given** a fresh project, **When** the engineer runs
   `uf init`, **Then** convention packs are deployed to
   `.opencode/uf/packs/` (not `.opencode/unbound/packs/`).
2. **Given** all 13 agent files reference
   `.opencode/uf/packs/`, **When** an agent loads its
   convention pack, **Then** the pack is found at the
   new path.
3. **Given** the scaffold engine processes embedded
   assets, **When** pack files are deployed, **Then**
   they are written to `opencode/uf/packs/` in the
   target directory.

---

### User Story 4 — Setup Uses .uf/ Paths (Priority: P2)

When an engineer runs `uf setup`, all references to
per-repo directories use `.uf/` paths. Today, setup
references `.dewey/` and delegates `dewey init` and
`replicator setup`. After this change, setup references
`.uf/dewey/` and `.uf/replicator/` in its output and
delegates to tools that create the new paths.

**Why this priority**: Setup is per-machine and runs
infrequently. The path references in setup output are
informational, not structural. Init (US-1) and Doctor
(US-2) are higher impact.

**Independent Test**: Run `uf setup` and verify no
`.dewey/`, `.hive/`, or `.unbound-force/` references
appear in the output.

**Acceptance Scenarios**:

1. **Given** an engineer runs `uf setup`, **When** the
   output references tool directories, **Then** all
   paths use `.uf/` notation.

---

### User Story 5 — Orchestration and Hero Data Under .uf/ (Priority: P2)

The orchestration engine stores workflow state and
artifacts under `.uf/workflows/` and `.uf/artifacts/`
(was `.unbound-force/workflows/` and
`.unbound-force/artifacts/`). Muti-Mind stores backlog
data under `.uf/muti-mind/` (was `.muti-mind/`). Mx F
stores metrics under `.uf/mx-f/` (was `.mx-f/`).

**Why this priority**: These paths are used by internal
Go packages and CLI commands. Changing them is
mechanical but affects multiple packages.

**Independent Test**: Run the orchestration engine's
test suite. Verify all workflow and artifact paths
reference `.uf/`. Verify the Muti-Mind CLI defaults to
`.uf/muti-mind/backlog`.

**Acceptance Scenarios**:

1. **Given** the orchestration engine creates a workflow,
   **When** the workflow state is persisted, **Then**
   the JSON file is written to `.uf/workflows/`.
2. **Given** an artifact is produced by a hero, **When**
   it is saved, **Then** it is written to
   `.uf/artifacts/`.
3. **Given** the `mutimind` CLI runs, **When** it
   accesses backlog data, **Then** the default path is
   `.uf/muti-mind/backlog`.
4. **Given** the `mxf` CLI collects metrics, **When**
   it writes data, **Then** the default path is
   `.uf/mx-f/data`.

---

### Edge Cases

- What happens when a project has both old (`.dewey/`)
  and new (`.uf/dewey/`) directories? The tools use
  `.uf/dewey/` exclusively. Old directories are ignored
  — no detection, no warning, no migration.
- What happens when `uf init` runs but `.uf/` parent
  directory does not exist? Init creates `.uf/` via
  `os.MkdirAll` before creating subdirectories.
- What happens when a `.gitignore` references old
  directories? The `.gitignore` is updated to reference
  `.uf/` instead of the old paths.
- What happens when agent files reference the old
  `.opencode/unbound/packs/` path? All agent files
  are updated to `.opencode/uf/packs/`. The old path
  is not checked or supported.
- What happens when the hero contract validation script
  checks `.unbound-force/hero.json`? The script is
  updated to check `.uf/hero.json`.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `uf init` MUST create `.uf/` as the root
  per-repo directory, with `config.yaml` initialized
  inside it.
- **FR-002**: `uf init` MUST delegate to `dewey init`
  which creates `.uf/dewey/` (not `.dewey/`).
- **FR-003**: `uf init` MUST delegate to
  `replicator init` which creates `.uf/replicator/`
  (not `.hive/`).
- **FR-004**: `uf init` MUST NOT create, detect, or
  interact with old directories (`.dewey/`, `.hive/`,
  `.unbound-force/`, `.muti-mind/`, `.mx-f/`).
- **FR-005**: `uf init` MUST deploy convention packs to
  `.opencode/uf/packs/` (not `.opencode/unbound/packs/`).
- **FR-006**: `uf init` MUST configure `opencode.json`
  with updated Dewey serve command referencing the new
  workspace path.
- **FR-007**: `uf doctor` MUST check `.uf/dewey/` for
  Dewey workspace existence (not `.dewey/`).
- **FR-008**: `uf doctor` MUST check `.uf/replicator/`
  for Replicator data existence (not `.hive/`).
- **FR-009**: `uf doctor` MUST check `.uf/config.yaml`
  for workflow configuration (not
  `.unbound-force/config.yaml`).
- **FR-010**: `uf doctor` install hints MUST reference
  `.uf/` paths.
- **FR-011**: All agent files referencing convention
  pack paths MUST use `.opencode/uf/packs/` (not
  `.opencode/unbound/packs/`).
- **FR-012**: All scaffold embedded assets MUST be
  moved from `opencode/unbound/packs/` to
  `opencode/uf/packs/` in the asset directory tree.
- **FR-013**: The orchestration engine MUST use
  `.uf/workflows/` and `.uf/artifacts/` for workflow
  state and artifact storage (not
  `.unbound-force/workflows/` and
  `.unbound-force/artifacts/`).
- **FR-014**: The `mutimind` CLI MUST default to
  `.uf/muti-mind/` for backlog and artifact paths
  (not `.muti-mind/`).
- **FR-015**: The `mxf` CLI MUST default to
  `.uf/mx-f/` for metrics and impediment paths
  (not `.mx-f/`).
- **FR-016**: The hero contract validation script MUST
  check `.uf/hero.json` (not `.unbound-force/hero.json`).
- **FR-017**: The hero manifest schema MUST reference
  `.uf/hero.json` as the canonical location.
- **FR-018**: The `.gitignore` MUST be updated to
  ignore `.uf/` instead of `.dewey/`, `.unbound-force/`,
  `.muti-mind/`, and `.mx-f/`.
- **FR-019**: All existing tests MUST continue to pass
  after the migration.
- **FR-020**: The `opencode.json` at the repo root MUST
  be updated with the new Dewey serve command path.

### Key Entities

- **Tool Directory**: A per-repo directory managed by
  a specific tool (Dewey, Replicator, Muti-Mind, Mx F).
  Lives under `.uf/<tool-name>/`.
- **Convention Pack**: A Markdown file defining coding
  or content standards. Lives under
  `.opencode/uf/packs/`.
- **Scaffold Asset**: An embedded file deployed by
  `uf init`. Paths in the asset tree mirror the
  target deployment paths.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Zero references to `.dewey/`, `.hive/`,
  `.unbound-force/`, `.muti-mind/`, or `.mx-f/` remain
  in production source code — verified by text search
  across all Go source and live Markdown agent/command
  files (excluding historical spec documents).
- **SC-002**: Zero references to `.opencode/unbound/`
  remain in production source code, agent files, or
  scaffold assets — verified by text search.
- **SC-003**: `uf init` in a fresh directory creates
  `.uf/` with appropriate subdirectories and
  `.opencode/uf/packs/` with all 9 convention pack
  files — verified by running init and checking
  the filesystem.
- **SC-004**: `uf doctor` reports all checks using
  `.uf/` paths — verified by running doctor and
  checking output for old path references.
- **SC-005**: All existing tests pass after the
  migration — verified by running the full test suite.
- **SC-006**: The scaffold drift detection tests pass
  — verified by running scaffold tests specifically.

## Dependencies & Assumptions

### Dependencies

- **Dewey workspace path change**
  (`unbound-force/dewey#33`): Dewey must accept
  `.uf/dewey/` as its workspace directory. `dewey init`
  must create `.uf/dewey/`. `dewey serve` must accept
  the new path. Hard blocker.
- **Replicator directory change**
  (`unbound-force/replicator#9`): Replicator must use
  `.uf/replicator/` for per-repo data and
  `~/.config/uf/replicator/` for per-system data.
  `replicator init` must create `.uf/replicator/`.
  Hard blocker.

### Assumptions

- The Dewey `serve` command accepts a flag to specify
  the workspace path (e.g., `--workspace .uf/dewey` or
  `--vault .uf/dewey`). The exact flag name depends on
  what Dewey implements in #33.
- No backward compatibility is provided. Old directories
  are ignored. Users upgrading `rm -rf` old directories
  and re-run `uf init`.
- Historical spec documents under `specs/` are not
  updated — they are archival records of past work.
  Only production code, live agent/command files,
  config files, and documentation (AGENTS.md) are
  updated.
- The `.opencode/uf/packs/` rename requires moving
  the scaffold asset directory from
  `internal/scaffold/assets/opencode/unbound/packs/`
  to `internal/scaffold/assets/opencode/uf/packs/`.
  This is a directory rename in the embedded filesystem.
