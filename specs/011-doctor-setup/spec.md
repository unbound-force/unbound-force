# Feature Specification: Doctor and Setup Commands

**Feature Branch**: `011-doctor-setup`
**Created**: 2026-03-21
**Status**: Draft
**Input**: User description: "Add unbound doctor and unbound setup commands for environment health checking and automated Swarm plugin installation"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Diagnose Environment Health (Priority: P1)

A developer runs `unbound doctor` in a project directory to
understand whether their Unbound Force development environment
is correctly configured. The command checks for required tools,
scaffolded files, hero availability, Swarm plugin status, MCP
server config, and agent/skill integrity. Each check produces
a pass, warning, or failure result with clear install
instructions when something is missing.

**Why this priority**: Without a diagnostic command, developers
have no way to know what is missing or broken in their
environment. This is the prerequisite for every other workflow
in the ecosystem -- if tools are missing, nothing else works.
This is also the foundation that `unbound setup` builds on.

**Independent Test**: Can be fully tested by running
`unbound doctor` in a project with intentionally missing
dependencies and verifying that each missing item produces
the correct failure message with install instructions.

**Acceptance Scenarios**:

1. **Given** a project with all tools installed and `unbound init`
   completed, **When** the developer runs `unbound doctor`,
   **Then** all checks pass and the exit code is 0.

2. **Given** a project where `gaze` is not installed, **When**
   the developer runs `unbound doctor`, **Then** the Core
   Tools group shows a warning for `gaze` with an install
   instruction appropriate to the detected package manager
   (e.g., `brew install unbound-force/tap/gaze` if Homebrew
   is detected).

3. **Given** a project where `opencode` is not installed,
   **When** the developer runs `unbound doctor`, **Then** the
   Core Tools group shows a failure for `opencode` with an
   install instruction appropriate to the detected package
   manager (e.g., `brew install anomalyco/tap/opencode` if
   Homebrew is detected, or
   `curl -fsSL https://opencode.ai/install | bash`
   otherwise).

4. **Given** a project where `unbound init` has not been run,
   **When** the developer runs `unbound doctor`, **Then** the
   Scaffolded Files group shows failures for missing
   directories with the hint `Run: unbound init`.

5. **Given** a project directory, **When** the developer runs
   `unbound doctor --format=json`, **Then** the output is
   valid JSON containing all check groups, results, and
   summary counts.

---

### User Story 2 - Validate Swarm Plugin and Include Swarm Doctor Output (Priority: P1)

A developer runs `unbound doctor` and the command checks
whether the OpenCode Swarm plugin is installed, configured
in `opencode.json`, and healthy. When the `swarm` binary is
found, `unbound doctor` shells out to `swarm doctor`, captures
its output, and embeds it verbatim as a subsection of the
report. When Swarm is not installed, the report shows a
failure with install instructions.

**Why this priority**: Swarm is a core coordination layer in
the Unbound Force stack. Developers need to know whether
their Swarm setup is healthy, and the `swarm doctor` output
provides the most authoritative health check. Embedding it
avoids duplicating Swarm's own validation logic.

**Independent Test**: Can be tested by running `unbound doctor`
with and without the `swarm` binary in PATH and verifying
that the Swarm Plugin section shows the correct output in
each case.

**Acceptance Scenarios**:

1. **Given** the Swarm plugin is installed and `swarm setup`
   has been run, **When** the developer runs `unbound doctor`,
   **Then** the Swarm Plugin section shows the `swarm` binary
   location, embeds the full `swarm doctor` output, and
   reports whether `.hive/` is initialized and whether
   `opencode-swarm-plugin` is listed in `opencode.json`
   plugins.

2. **Given** the Swarm plugin is not installed, **When** the
   developer runs `unbound doctor`, **Then** the Swarm Plugin
   section shows a failure with the install instruction
   `npm install -g opencode-swarm-plugin@latest` and the
   hint `Then run: unbound setup`.

3. **Given** the Swarm binary is installed but `swarm doctor`
   exits with a non-zero code, **When** the developer runs
   `unbound doctor`, **Then** the Swarm Plugin section shows
   a warning, includes the `swarm doctor` stderr/stdout
   output, and suggests `Run: unbound setup`.

4. **Given** the Swarm binary is installed but
   `opencode-swarm-plugin` is not listed in the `opencode.json`
   `plugin` array, **When** the developer runs
   `unbound doctor`, **Then** the Swarm Plugin section shows
   a warning for the missing plugin config entry with the
   hint `Fix: unbound setup`.

---

### User Story 3 - Automated Environment Setup (Priority: P2)

A developer runs `unbound setup` to automatically install all
required and recommended tools that `unbound doctor` would flag
as missing: OpenCode, Gaze, Swarm plugin, and their
configuration. The command installs OpenCode (via its install
script), Gaze (via Homebrew), the Swarm plugin (via npm), runs
`swarm setup`, configures the `opencode.json` plugin entry, and
initializes the project's `.hive/` directory. The command is
idempotent -- safe to run multiple times without breaking
existing configuration.

**Why this priority**: Manual installation of the full tool
chain requires multiple steps across different package managers
and config files. Automating this reduces onboarding friction
and eliminates common misconfiguration errors. This is secondary
to diagnosis because developers need to understand what is
wrong (doctor) before fixing it (setup).

**Independent Test**: Can be tested by running `unbound setup`
in a project with no Swarm installed and verifying that Swarm
is installed, configured, and functional afterward.

**Acceptance Scenarios**:

1. **Given** a machine with Homebrew and nvm-managed Node.js
   >= 18 but no Unbound Force tools, **When** the developer
   runs `unbound setup`, **Then** the command detects
   Homebrew and nvm, installs OpenCode (via Homebrew), Gaze
   (via Homebrew), `opencode-swarm-plugin` (via npm from the
   nvm-managed Node), runs `swarm setup`, adds
   `opencode-swarm-plugin` to the `opencode.json` `plugin`
   array, initializes `.hive/` in the project directory, and
   prints a success summary showing which managers were used.

2. **Given** Node.js is not installed, **When** the developer
   runs `unbound setup`, **Then** the command prints a failure
   message with Node.js install instructions
   (`brew install node` or https://nodejs.org) and exits
   without attempting further steps.

3. **Given** all tools are already installed and configured,
   **When** the developer runs `unbound setup`, **Then** the
   command detects the existing installations, skips all
   redundant steps, and reports that everything is already
   configured.

4. **Given** an `opencode.json` exists with MCP servers and
   other config, **When** `unbound setup` adds the Swarm
   plugin entry, **Then** all existing configuration is
   preserved and only the `plugin` array is added or updated.

5. **Given** `npm install -g opencode-swarm-plugin` fails,
   **When** the developer runs `unbound setup`, **Then** the
   command reports the npm error, suggests manual install
   steps, and does not proceed to `swarm setup`.

---

### User Story 4 - Colored Terminal Output with Install Guidance (Priority: P2)

The `unbound doctor` command produces colored terminal output
that uses visual indicators (green checkmarks, yellow warnings,
red failures) to help developers quickly scan results. Each
failure includes a brief install instruction and, where
appropriate, a link to detailed documentation.

**Why this priority**: Colored output significantly improves
readability for interactive terminal use. Install hints
directly in the output reduce the number of steps a developer
needs to take to fix their environment. This is secondary to
the diagnostic logic itself.

**Independent Test**: Can be tested by capturing terminal
output and verifying that pass/warn/fail indicators are
present and that install hints appear for each failed check.

**Acceptance Scenarios**:

1. **Given** a terminal that supports ANSI colors, **When** the
   developer runs `unbound doctor`, **Then** passed checks
   show a green checkmark, warnings show a yellow exclamation,
   failures show a red cross, and optional missing items
   show a gray circle.

2. **Given** a check fails for a missing tool, **When** the
   result is displayed, **Then** the output includes a brief
   install command (e.g., `brew install unbound-force/tap/gaze`)
   on the line below the failure, and a URL to detailed
   instructions when the install process is non-trivial
   (e.g., https://opencode.ai/docs for OpenCode).

3. **Given** a terminal that does not support colors (piped
   output or `NO_COLOR` env var set), **When** the developer
   runs `unbound doctor`, **Then** the output uses plain text
   indicators (`[PASS]`, `[WARN]`, `[FAIL]`) instead of
   colored symbols.

---

### User Story 5 - Machine-Readable JSON Output (Priority: P3)

The `unbound doctor` command supports a `--format=json` flag
that produces structured JSON output suitable for CI pipelines,
scripting, and automated environment validation.

**Why this priority**: JSON output enables integration with CI
systems and other tools, but most developers will use the
default text output interactively. This is a lower priority
convenience feature.

**Independent Test**: Can be tested by running
`unbound doctor --format=json` and validating the output
against the expected JSON structure.

**Acceptance Scenarios**:

1. **Given** any environment state, **When** the developer runs
   `unbound doctor --format=json`, **Then** the output is
   valid JSON containing a `groups` array (each with `name`
   and `results`), a `summary` object with `total`, `passed`,
   `warned`, `failed` counts, and each result includes
   `name`, `severity`, `message`, and optional
   `install_hint` and `install_url` fields.

2. **Given** the `--format=json` flag is set, **When** the
   command encounters a failed check, **Then** the JSON
   result for that check includes `install_hint` with the
   install command and `install_url` with the documentation
   link (if applicable).

---

### Edge Cases

- What happens when `swarm doctor` hangs or takes too long?
  The command MUST enforce a timeout (10 seconds) on the
  `swarm doctor` subprocess. If it times out, the Swarm
  Plugin section reports a warning with the message
  "swarm doctor timed out" and suggests running
  `swarm doctor` manually.

- What happens when `opencode.json` is malformed JSON?
  The MCP Server Config and Swarm Plugin sections report a
  warning that the file could not be parsed, and `unbound setup`
  refuses to modify the file, suggesting the developer fix
  the JSON syntax manually.

- What happens when the user runs `unbound doctor` outside
  a git repository? The command SHOULD still work for all
  checks except `.hive/` initialization (which requires git).
  The report notes that `.hive/` cannot be validated outside
  a git repository.

- What happens when `go version` output changes format?
  The version check SHOULD fail gracefully -- if the version
  string cannot be parsed, the check passes with a warning
  that the version could not be verified.

- What happens when `npm install -g` requires sudo?
  The `unbound setup` command runs npm without sudo. If the
  install fails due to permissions, the error message suggests
  using a Node version manager (nvm, fnm) or fixing npm
  global prefix permissions, with a link to the npm docs.

- What happens when the developer runs `unbound setup`
  without having run `unbound init` first? Setup runs
  `unbound init` automatically at the end if `.opencode/`
  is missing, completing the full onboarding in one command.

- What happens when Homebrew is not installed? Gaze install
  via `brew` is skipped with a warning. The warning provides
  alternative install instructions (download binary from
  GitHub releases). OpenCode install falls back to its curl
  install script. Swarm install is unaffected since it uses
  npm.

- What happens when multiple managers exist for the same
  tool category (e.g., both nvm and fnm for Node.js)?
  The command SHOULD prefer the manager whose environment
  is currently active (e.g., `NVM_DIR` is set and nvm shims
  are in PATH). If both are equally active, it SHOULD prefer
  the first detected in a stable precedence order documented
  in the implementation.

- What happens when a version manager is detected but the
  required tool version is not installed through it?
  `unbound doctor` reports the version gap with a
  manager-specific upgrade hint (e.g.,
  `goenv install 1.24.3`). `unbound setup` attempts the
  install through that manager.

## Clarifications

### Session 2026-03-21

- Q: Should `unbound setup` install all required/recommended
  tools or only Swarm? → A: All required and recommended
  tools (OpenCode, Gaze, Swarm plugin and configuration).
  Setup handles the full tool chain, not just Swarm.

- Q: Should `unbound setup` also run `unbound init` if
  `.opencode/` is missing? → A: Yes. Setup runs
  `unbound init` at the end if `.opencode/` is missing,
  completing the full onboarding flow in a single command.

- Q: Should doctor/setup detect the developer's existing
  version and package managers (goenv, nvm, pyenv, Homebrew)
  and tailor install hints and install methods accordingly?
  → A: Yes. Doctor detects active version managers and
  tailors hints to match; setup installs through detected
  managers (e.g., `nvm install` instead of raw `brew install
  node`, `goenv install` instead of `brew install go`).

- Q: Should `unbound doctor` display the detected environment
  (which managers are present) as a dedicated section?
  → A: Yes. Show as the first section ("Detected
  Environment") at the top of the report, listing all
  detected version and package managers so the developer can
  verify detection accuracy before reviewing tool checks.

- Q: Should `unbound doctor` report the install provenance
  of each detected tool (e.g., "go 1.24.3 via goenv")?
  → A: Yes. Each tool check shows version plus the detected
  manager that provides it (e.g., "1.24.3 via goenv",
  "22.15.0 via nvm").

## Requirements *(mandatory)*

### Functional Requirements

#### Doctor Command

- **FR-000**: Both `unbound doctor` and `unbound setup` MUST
  detect the developer's active version and package managers
  before performing checks or installs. Detection MUST
  include: `goenv` (Go), `nvm`/`fnm` (Node.js), `pyenv`
  (Python), `mise` (polyglot), Homebrew, and `bun`.
  Detection is performed by checking for each manager's
  binary in PATH and, where applicable, its environment
  variables (e.g., `GOENV_ROOT`, `NVM_DIR`, `FNM_DIR`).
  Detected managers MUST be used to tailor install hints in
  doctor output and install methods in setup execution.

- **FR-000a**: `unbound doctor` MUST display a "Detected
  Environment" section as the first group in the report,
  listing all detected version and package managers, their
  paths, and which tool categories they manage. This section
  is informational only (all items are pass severity) and
  allows the developer to verify detection accuracy before
  reviewing tool checks.

- **FR-001**: `unbound doctor` MUST check for required
  binaries (`go`, `opencode`) and report failure with install
  instructions tailored to the detected version manager
  (e.g., `goenv install 1.24.3` if goenv is detected, or
  `brew install go` if only Homebrew is available). When
  a binary is found, the check MUST report its version and
  the detected install provenance (e.g., "1.24.3 via
  goenv" or "0.2.15 via Homebrew").

- **FR-002**: `unbound doctor` MUST check for recommended
  binaries (`gaze`, `mxf`) and report warnings with install
  instructions when they are missing.

- **FR-003**: `unbound doctor` MUST check for optional
  binaries (`graphthulhu`, `node`, `gh`, `swarm`) and report
  their absence as informational (not failure) with install
  instructions.

- **FR-004**: `unbound doctor` MUST parse `go version` output
  and verify the Go version is >= 1.24. If the version is
  below 1.24, the check MUST fail with upgrade instructions
  appropriate to the detected manager (e.g.,
  `goenv install 1.24.3 && goenv global 1.24.3` if goenv
  is detected).

- **FR-005**: `unbound doctor` MUST parse `node --version`
  output and verify Node.js >= 18 when present. If below 18,
  the check MUST warn with upgrade instructions appropriate
  to the detected manager (e.g., `nvm install 22` if nvm is
  detected, `fnm install 22` if fnm is detected, or
  `brew install node` otherwise).

- **FR-006**: `unbound doctor` MUST verify that `unbound init`
  scaffolded files exist: `.opencode/agents/` (with at least
  one `.md` file), `.opencode/command/` (with at least one
  `.md` file), `.specify/` directory, and `AGENTS.md`.

- **FR-007**: `unbound doctor` MUST detect hero availability
  by checking for agent files in `.opencode/agents/` and
  binaries in PATH, consistent with the existing
  `DetectHeroes()` function.

- **FR-008**: When the `swarm` binary is found, `unbound doctor`
  MUST execute `swarm doctor` as a subprocess, capture its
  stdout and stderr, and embed the output verbatim in the
  report's Swarm Plugin section.

- **FR-009**: `unbound doctor` MUST enforce a 10-second
  timeout on the `swarm doctor` subprocess. If it exceeds
  the timeout, the check MUST report a warning.

- **FR-010**: `unbound doctor` MUST check whether `.hive/`
  exists in the target directory and report its status in
  the Swarm Plugin section.

- **FR-011**: `unbound doctor` MUST parse `opencode.json`
  (if present) and check that each MCP server's command
  binary exists in PATH.

- **FR-012**: `unbound doctor` MUST check whether
  `opencode-swarm-plugin` is listed in the `opencode.json`
  `plugin` array and report a warning if it is missing.

- **FR-013**: `unbound doctor` MUST validate YAML frontmatter
  in all `.md` files under `.opencode/agents/`, checking that
  `description` is present and non-empty.

- **FR-014**: `unbound doctor` MUST validate each `SKILL.md`
  file under `.opencode/skill/` or `.opencode/skills/`,
  checking that `name` and `description` frontmatter fields
  exist, and that `name` matches its parent directory name
  and the pattern `^[a-z0-9]+(-[a-z0-9]+)*$`.

- **FR-015**: Each failed or warned check MUST include an
  `install_hint` field containing a brief install command
  or fix instruction.

- **FR-016**: Each failed check for a non-trivial install
  SHOULD include an `install_url` field containing a link
  to detailed documentation.

- **FR-017**: `unbound doctor` MUST support a `--format` flag
  accepting `text` (default) or `json`.

- **FR-018**: `unbound doctor` MUST support a `--dir` flag to
  specify the target directory (defaults to current working
  directory).

- **FR-019**: `unbound doctor` text output MUST use colored
  indicators (green/yellow/red/gray) when the terminal
  supports ANSI colors, and plain text indicators
  (`[PASS]`/`[WARN]`/`[FAIL]`) when it does not.

- **FR-020**: `unbound doctor` MUST exit with code 0 when all
  checks pass or only warnings exist, and exit with code 1
  when any check fails.

#### Setup Command

- **FR-021**: `unbound setup` MUST install all required and
  recommended tools that are missing, using the developer's
  detected version and package managers. The install order
  MUST be: OpenCode, Gaze, Node.js check, Swarm plugin,
  Swarm configuration. Each step is skipped if the tool is
  already present.

- **FR-022**: `unbound setup` MUST install OpenCode via
  Homebrew (`brew install anomalyco/tap/opencode`) if
  Homebrew is detected, or via its install script
  (`curl -fsSL https://opencode.ai/install | bash`)
  otherwise. If the `opencode` binary is already in PATH,
  the step is skipped.

- **FR-023**: `unbound setup` MUST install Gaze via Homebrew
  (`brew install unbound-force/tap/gaze`) if the `gaze`
  binary is not in PATH and Homebrew is available. If
  Homebrew is not available, the step MUST warn and provide
  alternative install instructions (download binary from
  GitHub releases).

- **FR-024**: `unbound setup` MUST verify that Node.js >= 18
  is available before proceeding to Swarm installation. If
  Node.js is missing and a Node version manager is detected
  (nvm or fnm), setup MUST install Node.js through that
  manager. If no manager is detected and Homebrew is
  available, setup MUST use `brew install node`. If neither
  is available, setup MUST print install instructions and
  skip Swarm-related steps.

- **FR-025**: `unbound setup` MUST install the Swarm plugin
  via `npm install -g opencode-swarm-plugin@latest` if the
  `swarm` binary is not in PATH. If `bun` is detected
  instead of `npm`, setup SHOULD use
  `bun add -g opencode-swarm-plugin@latest`.

- **FR-026**: `unbound setup` MUST run `swarm setup` after
  installing the plugin to initialize Swarm's own
  configuration.

- **FR-027**: `unbound setup` MUST add
  `opencode-swarm-plugin` to the `opencode.json` `plugin`
  array if it is not already present, preserving all
  existing configuration.

- **FR-028**: `unbound setup` MUST create a minimal
  `opencode.json` with `$schema` if no `opencode.json`
  exists.

- **FR-029**: `unbound setup` MUST run `swarm init` in the
  target directory if `.hive/` does not exist.

- **FR-030**: `unbound setup` MUST be idempotent -- running it
  multiple times on an already-configured project MUST NOT
  break existing configuration or produce errors.

- **FR-031**: `unbound setup` MUST support a `--dir` flag to
  specify the target directory (defaults to current working
  directory).

- **FR-032**: When any step fails, `unbound setup` MUST
  report the specific error, suggest manual alternatives,
  and continue to subsequent independent steps where
  possible (e.g., OpenCode install failure should not
  prevent Gaze install attempt).

- **FR-033**: `unbound setup` MUST run `unbound init` at the
  end of the setup process if `.opencode/` does not exist in
  the target directory, completing the full project
  onboarding flow.

- **FR-034**: `unbound setup` MUST print a completion summary
  showing what was done (installed, skipped, failed) and
  suggesting next steps (including `unbound doctor` for
  verification).

### Key Entities

- **DetectedEnvironment**: The developer's detected version
  and package manager landscape. Includes which managers are
  present (goenv, nvm, fnm, pyenv, mise, Homebrew, bun),
  how each was detected (binary in PATH, environment
  variable), and the preferred install method for each tool
  category (Go, Node.js, general packages).

- **CheckResult**: A single diagnostic finding with name,
  severity (pass/warn/fail), human-readable message, optional
  detail (path, version, provenance manager), optional
  install hint (tailored to detected environment), and
  optional documentation URL.

- **CheckGroup**: A named collection of related CheckResults.
  The report contains 7 groups in order: Detected
  Environment, Core Tools, Swarm Plugin, Scaffolded Files,
  Hero Availability, MCP Server Config, Agent/Skill
  Integrity. Groups may include optional embedded output
  from external tools (e.g., `swarm doctor`).

- **Report**: The complete diagnostic output containing an
  ordered list of CheckGroups and a Summary with aggregate
  counts (total, passed, warned, failed).

- **Summary**: Aggregate counts of check results by severity,
  used for both display and exit code determination.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A developer with no prior Unbound Force
  experience can run `unbound doctor` and understand within
  30 seconds what is missing and how to install each
  dependency.

- **SC-002**: Running `unbound setup` on a machine with
  Node.js installed but no Swarm results in a fully
  functional Swarm environment in under 2 minutes
  (network speed permitting).

- **SC-003**: `unbound doctor` correctly identifies and
  reports all 8 checked binaries (go, opencode, gaze, mxf,
  graphthulhu, node, gh, swarm) with accurate install
  instructions for each.

- **SC-004**: `unbound doctor --format=json` output can be
  parsed by standard JSON tools and contains all the same
  information as the text output.

- **SC-005**: Running `unbound setup` twice in succession
  produces no errors, no duplicate config entries, and no
  modified files on the second run.

- **SC-006**: When `swarm doctor` output is embedded in the
  `unbound doctor` report, the embedded output faithfully
  reproduces what `swarm doctor` would show if run directly.

- **SC-007**: Every failed check in `unbound doctor` includes
  a copy-pasteable install command that, when executed,
  resolves the failure on the next `unbound doctor` run.

### Assumptions

- Developers are using macOS or Linux. Install hints are
  tailored to the detected version and package managers on
  the system rather than assuming a single method. When no
  manager is detected, hints fall back to Homebrew (macOS)
  or direct download instructions.

- The Swarm plugin distribution remains via npm
  (`opencode-swarm-plugin` package). If distribution
  changes, install hints in doctor will need updating.

- `swarm doctor` produces its output to stdout and uses exit
  code 0 for healthy, non-zero for issues. If Swarm changes
  this contract, the embedding logic will need updating.

- `opencode.json` follows the schema at
  https://opencode.ai/config.json with a `plugin` key
  accepting an array of npm package names.

- The `go version` output format follows the standard
  `go version goX.Y.Z <platform>` pattern.

- The `node --version` output format follows the standard
  `vX.Y.Z` pattern.

### Dependencies

- Existing `internal/orchestration.DetectHeroes()` function
  (reused for hero availability checks).

- `charmbracelet/lipgloss` library (already an indirect
  dependency, promoted to direct for colored output).

- The `swarm` CLI binary and its `doctor`, `setup`, and
  `init` subcommands (external dependency, not modified by
  this spec).
