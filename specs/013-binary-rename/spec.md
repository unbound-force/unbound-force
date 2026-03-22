---
spec_id: "013"
title: "Binary Rename"
phase: 3
status: draft
depends_on:
  - "[[specs/003-specification-framework/spec]]"
  - "[[specs/011-doctor-setup/spec]]"
---

# Feature Specification: Binary Rename

**Feature Branch**: `013-binary-rename`
**Created**: 2026-03-22
**Status**: Draft
**Input**: User description: "Rename unbound CLI binary
to unbound-force with uf alias to avoid NLnet Labs
Unbound DNS resolver name collision"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - CLI Without Name Collision (Priority: P1)

A developer installs the Unbound Force CLI tool on a
machine that also has the NLnet Labs Unbound DNS resolver
installed (via `brew install unbound`). The two tools
must coexist without conflict. When the developer types
the Unbound Force command, they always get the Unbound
Force CLI -- never the DNS resolver.

**Why this priority**: This is the core problem. Without
solving the name collision, developers cannot reliably
use the CLI. Every other user story depends on the binary
having an unambiguous name.

**Independent Test**: Install both the NLnet Labs Unbound
DNS resolver and the Unbound Force CLI on the same
machine. Verify that `unbound-force --help` and
`uf --help` produce the Unbound Force help output, and
that `unbound --help` still produces the DNS resolver
help output (unaffected).

**Acceptance Scenarios**:

1. **Given** a machine with NLnet Labs Unbound DNS
   resolver installed, **When** the developer installs
   the Unbound Force CLI, **Then** both tools are
   accessible by their respective names with no
   conflict.
2. **Given** the Unbound Force CLI is installed,
   **When** the developer runs `unbound-force --help`,
   **Then** the output shows the full Unbound Force
   command list (init, doctor, setup, version).
3. **Given** the Unbound Force CLI is installed,
   **When** the developer runs `uf --help`, **Then**
   the output is identical to `unbound-force --help`.
4. **Given** a fresh machine with neither tool
   installed, **When** the developer installs only
   the Unbound Force CLI, **Then** both
   `unbound-force` and `uf` are available in the PATH.

---

### User Story 2 - Scaffold Output References (Priority: P1)

When a developer runs `uf init` (or `unbound-force init`)
in a project directory, the scaffolded files (agent
personas, convention packs, commands) reference the
correct binary name. No scaffolded file references the
bare `unbound` command name. Developers following the
instructions in scaffolded files always invoke the
correct binary.

**Why this priority**: Scaffold output is the primary
way new projects learn the CLI commands. If scaffolded
agent files say "run `unbound init --divisor`" but the
binary is named `unbound-force`, developers get confused
or invoke the DNS resolver by mistake.

**Independent Test**: Run `uf init` in a fresh temporary
directory. Search all generated files for the string
`unbound `. Verify zero matches. Search for
`unbound-force` or `uf` references. Verify they exist
where CLI commands are referenced.

**Acceptance Scenarios**:

1. **Given** a fresh directory with no scaffold files,
   **When** the developer runs `uf init`, **Then** all
   generated files reference `uf` or `unbound-force`
   for CLI commands, not the bare `unbound`.
2. **Given** a project with existing scaffold files
   from the old `unbound` binary, **When** the
   developer runs `uf init`, **Then** tool-owned files
   are updated with the new binary name references
   while user-owned files are preserved.

---

### User Story 3 - Homebrew Distribution (Priority: P2)

The Unbound Force CLI is distributed via Homebrew with
the correct binary name and alias. The Homebrew formula
installs `unbound-force` as the primary binary and
creates a `uf` symlink. The formula name does not
conflict with the existing `unbound` Homebrew formula
(NLnet Labs DNS resolver).

**Why this priority**: Distribution is how most
developers get the CLI. Without a correct Homebrew
formula, the name collision persists for every new
installation. This depends on US1 (the binary rename
itself) being complete first.

**Independent Test**: Run `brew install` for the Unbound
Force formula. Verify both `unbound-force` and `uf` are
in the PATH. Verify `brew install unbound` still
installs the DNS resolver without conflict.

**Acceptance Scenarios**:

1. **Given** the Homebrew tap is configured, **When**
   the developer runs the install command for the
   Unbound Force formula, **Then** both
   `unbound-force` and `uf` binaries are available.
2. **Given** the Unbound Force CLI is installed via
   Homebrew, **When** the developer also installs
   `unbound` (DNS resolver) via Homebrew, **Then**
   both tools coexist without conflict.
3. **Given** a tagged release exists, **When** the
   release pipeline runs, **Then** the Homebrew formula
   is automatically updated with the correct binary
   name and symlink.

---

### User Story 4 - Doctor and Setup Commands (Priority: P2)

The `uf doctor` and `uf setup` commands reference the
correct binary name in all their output -- install hints,
fix suggestions, and progress messages. When a developer
reads "Run: uf setup" in a doctor warning, the command
works as shown.

**Why this priority**: Doctor and setup are the primary
onboarding commands. Their output directly tells
developers what to type. Incorrect binary names in hints
cause frustration and confusion.

**Independent Test**: Run `uf doctor` in a project with
intentionally missing tools. Verify all hint text
references `uf` commands. Run `uf setup --dry-run` and
verify all output references `uf`.

**Acceptance Scenarios**:

1. **Given** a project missing scaffolded files,
   **When** the developer runs `uf doctor`, **Then**
   the fix hint says `Run: uf init` (not
   `unbound init`).
2. **Given** a project missing the Swarm plugin,
   **When** the developer runs `uf doctor`, **Then**
   the fix hint says `Run: uf setup` (not
   `unbound setup`).
3. **Given** a developer running `uf setup`, **When**
   setup completes the scaffolding step, **Then** the
   progress message references `uf init` (not
   `unbound init`).

---

### User Story 5 - Cross-Repo Documentation (Priority: P3)

All living documentation across the Unbound Force
GitHub organization references the correct binary name.
Developers reading guides in any repo (meta, gaze,
website) see `uf` or `unbound-force` commands, never the
bare `unbound` command for the Unbound Force CLI.

**Why this priority**: Documentation consistency prevents
confusion but is not blocking for functionality. The CLI
works correctly regardless of what the docs say. This
is a polish step that can happen after the core rename.

**Independent Test**: Search all Markdown files in the
meta repo, gaze repo, and website repo for the pattern
`unbound init`, `unbound doctor`, `unbound setup`.
Verify zero matches in living documents (AGENTS.md,
README.md, agent files, website content). Completed
specs (under `specs/003-*`, `specs/005-*`,
`specs/011-*`) are excluded as historical records.

**Acceptance Scenarios**:

1. **Given** the meta repo's living documentation
   (AGENTS.md, README.md, agent files), **When** a
   developer searches for bare `unbound` CLI
   references, **Then** zero matches are found (only
   `unbound-force` or `uf` appear).
2. **Given** the gaze repo's agent files and
   documentation, **When** a developer searches for
   bare `unbound` CLI references, **Then** zero
   matches are found.
3. **Given** the website repo's content pages,
   **When** a developer searches for bare `unbound`
   CLI references, **Then** zero matches are found
   (only `uf` or `unbound-force` appear).
4. **Given** completed specs in the meta repo (all
   specs under `specs/` with `status: complete` or
   implementation completed, excluding
   `specs/013-binary-rename/` itself), **When** a
   developer reads them, **Then** they retain their
   original `unbound` references as historical records
   (not modified).

---

### Edge Cases

- What happens when a developer has the old `unbound`
  binary installed via `go install`? The old binary
  remains in `$GOPATH/bin/unbound`. Running
  `go install ./cmd/unbound-force/` installs the new
  binary alongside it. The developer should remove the
  old binary manually. The `uf doctor` command SHOULD
  detect and warn about the stale binary.
- What happens when `uf init` is run in a project that
  was previously scaffolded with the old `unbound` name?
  Tool-owned scaffold files are updated with the new
  references. User-owned files (custom convention packs,
  custom agent files) are not modified.
- What happens when a developer types `unbound init`
  after the rename? If the NLnet Labs DNS resolver is
  installed, they get a DNS resolver error. If neither
  tool is installed, they get "command not found." The
  CLI does not provide a compatibility shim for the old
  name.
- What happens to CI pipelines that reference the old
  binary name? They break and must be updated. The
  binary rename is a breaking change for any script or
  pipeline referencing the old `unbound` name.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The CLI binary MUST be named
  `unbound-force` as its primary executable name.
- **FR-002**: A `uf` alias MUST be provided as a
  symlink to `unbound-force`, installed alongside the
  primary binary.
- **FR-003**: Both `unbound-force --help` and
  `uf --help` MUST produce identical output showing
  all available commands (init, doctor, setup, version).
- **FR-004**: The help output MUST indicate the alias
  relationship (e.g., "unbound-force (alias: uf)").
- **FR-005**: The `uf init` command MUST scaffold files
  that reference `uf` or `unbound-force` for CLI
  commands, not the bare `unbound`.
- **FR-006**: The `uf doctor` output MUST reference
  `uf` in all install hints, fix suggestions, and
  remediation instructions.
- **FR-007**: The `uf setup` output MUST reference
  `uf` in all progress messages and instructions.
- **FR-008**: The release pipeline MUST produce binaries
  named `unbound-force` for all target platforms.
- **FR-009**: The Homebrew formula MUST install
  `unbound-force` as the primary binary and create a
  `uf` symlink.
- **FR-010**: The Homebrew formula MUST NOT conflict
  with the existing `unbound` formula (NLnet Labs DNS
  resolver).
- **FR-011**: `go install ./cmd/unbound-force/` MUST
  produce a binary named `unbound-force` in
  `$GOPATH/bin/`.
- **FR-012**: The project MUST provide a `make install`
  target that builds the `unbound-force` binary and
  creates the `uf` symlink in `$GOPATH/bin/`.
- **FR-013**: All living documentation in the meta repo
  (AGENTS.md, README.md, unbound-force.md, agent
  persona files) MUST reference `uf` or `unbound-force`,
  not bare `unbound`.
- **FR-014**: Completed architectural specs (historical
  records under `specs/`) MUST NOT be modified for the
  rename.
- **FR-015**: All scaffold assets embedded in the
  binary MUST reference `uf` or `unbound-force` for
  CLI commands, not bare `unbound`.

### Key Entities

- **Primary Binary Name**: `unbound-force` -- the
  canonical executable name, used in Homebrew formulas,
  `go install`, and formal documentation.
- **Alias**: `uf` -- a symlink to `unbound-force`,
  used for daily command-line convenience. Produces
  identical behavior and output.
- **Stale Binary**: Any `unbound` binary in
  `$GOPATH/bin/` from a previous `go install` of the
  old `cmd/unbound/` directory. Should be detected and
  warned about by `uf doctor`.

## Assumptions

- The NLnet Labs Unbound DNS resolver (`brew install
  unbound`) is a well-established project that will not
  change its name. The collision is permanent unless
  the Unbound Force CLI changes its name.
- Developers primarily use the short alias `uf` for
  daily work. The full `unbound-force` name appears in
  formal contexts (Homebrew formulas, `go install`,
  documentation).
- The rename is a one-time breaking change. All
  downstream consumers (CI pipelines, scripts, other
  repos) must update their references.
- Cross-repo documentation updates (gaze, website) are
  coordinated but can proceed independently after the
  meta repo rename is complete.
- The Homebrew tap repository
  (`unbound-force/homebrew-tap`) manages the formula.
  GoReleaser auto-publishes to the tap on tagged
  releases.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: On a machine with both the NLnet Labs
  DNS resolver and the Unbound Force CLI installed,
  running `unbound-force --help` produces the Unbound
  Force help output and `unbound --help` produces the
  DNS resolver output -- zero ambiguity between the
  two tools.
- **SC-002**: Zero living documentation files across
  the meta repo, gaze repo, and website repo contain
  bare `unbound init`, `unbound doctor`, or
  `unbound setup` references (only `uf` or
  `unbound-force` variants appear).
- **SC-003**: `uf init` in a fresh directory produces
  scaffold output where 100% of CLI command references
  use `uf` or `unbound-force`.
- **SC-004**: `brew install unbound-force/tap/unbound-force`
  succeeds and both `unbound-force` and `uf` are
  available in the PATH.
- **SC-005**: All existing tests pass with the renamed
  binary -- zero regressions.
- **SC-006**: `make install` produces both
  `unbound-force` and `uf` in `$GOPATH/bin/` in a
  single command.
