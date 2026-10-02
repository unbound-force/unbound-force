## ADDED Requirements

### Requirement: Stealth init mode

`uf init` MUST support a `--stealth` flag that, when set, scaffolds all
tool-owned files required for pipeline operation while leaving the
working tree git-clean and all tracked files unmodified. The flag MUST
default to `false` so that non-stealth behavior is unchanged.

#### Scenario: clean tree after stealth init

- **GIVEN** a git repository with a clean working tree
- **WHEN** the user runs `uf init --stealth`
- **THEN** `git status --porcelain` returns no output

#### Scenario: already-tracked scaffolded paths are refused

- **GIVEN** a git repository that already tracks one or more scaffolded
  paths (`opencode.json`, `.opencode/`, `.specify/`, `openspec/`, or
  `.uf/`)
- **WHEN** the user runs `uf init --stealth`
- **THEN** the command reports a clear error naming the tracked paths
- **AND** writes no files

#### Scenario: non-git directory

- **GIVEN** a directory that is not inside a git repository
- **WHEN** the user runs `uf init --stealth`
- **THEN** the command exits non-zero with a non-empty error message
  referencing the git requirement
- **AND** writes no files to the working directory

#### Scenario: git not installed

- **GIVEN** a directory where the `git` binary is not installed or not
  on PATH
- **WHEN** the user runs `uf init --stealth`
- **THEN** the command exits non-zero with a specific, actionable error
  distinguishing git-unavailable from not-a-work-tree
- **AND** writes no files to the working directory

#### Scenario: normal init unchanged

- **GIVEN** a git repository
- **WHEN** the user runs `uf init` without `--stealth`
- **THEN** the resulting `.gitignore` and `AGENTS.md` contents are
  byte-identical to what the pre-change implementation produced for the
  same inputs

### Requirement: Tracked file immutability

Stealth mode MUST NOT modify any tracked file. This MUST cover edits
made by `uf init` itself (`ensureGitignore`,
`ensureAGENTSmdPackSection`) AND edits made by any sub-tool that
`uf init` installs or invokes. Restoration MUST preserve both content
bytes and mode/permission bits.

#### Scenario: file mode preserved on restore

- **GIVEN** a tracked file with non-default mode (e.g. `0755`) or
  restrictive permissions (e.g. `0600`) that a sub-tool modifies
- **WHEN** the user runs `uf init --stealth`
- **THEN** the file's bytes AND mode/permission bits are identical
  before and after

#### Scenario: pre-existing .gitignore

- **GIVEN** a repository with a pre-existing `.gitignore`
- **WHEN** the user runs `uf init --stealth`
- **THEN** the bytes of `.gitignore` are identical before and after

#### Scenario: sub-tool attempts to edit .gitignore

- **GIVEN** a sub-tool invoked during `uf init` that writes to
  `.gitignore`
- **WHEN** the user runs `uf init --stealth`
- **THEN** `.gitignore` is restored to its pre-init bytes

### Requirement: Local exclusion mechanism

Stealth mode MUST hide scaffolded files using a local-only git facility
(`.git/info/exclude`) and MUST NOT append to the tracked `.gitignore`.
The excluded paths MUST cover `.opencode/`, `.specify/`, `openspec/`,
`.uf/`, and `opencode.json`. The exclusion path MUST be resolved via
`git rev-parse --git-path info/exclude` so linked worktrees and
submodules write to the correct location. The exclusion list MUST be
complete for all output that `uf init` and its sub-tools create; any
newly-created untracked path during init MUST be appended to the
marker-delimited block. Appended paths MUST be emitted as literal
patterns (leading `#`/`!` escaped, no trailing whitespace).

#### Scenario: exclusion written locally

- **GIVEN** a git repository
- **WHEN** the user runs `uf init --stealth`
- **THEN** `.git/info/exclude` contains the scaffolded paths
- **AND** the tracked `.gitignore` contains no new unbound-force entry

#### Scenario: exclusion is idempotent

- **GIVEN** a git repository where `uf init --stealth` has already
  completed
- **WHEN** the user runs `uf init --stealth` again
- **THEN** `.git/info/exclude` contains no duplicate entries
- **AND** `.gitignore` and `AGENTS.md` remain byte-identical

#### Scenario: exclusion target unwritable

- **GIVEN** a git repository whose `.git/info/exclude` cannot be written
- **WHEN** the user runs `uf init --stealth`
- **THEN** the command reports a clear error
- **AND** writes no scaffolded files (or rolls back any already
  written), leaving `git status --porcelain` empty

### Requirement: Pipeline operability

Pipeline commands MUST continue to resolve scaffolded files at their
conventional in-repo locations when stealth mode is active. No
indirection (symlink, overlay, or virtual path) MUST be required for
path resolution.

#### Scenario: pipeline command locates scaffold

- **GIVEN** a stealth init has completed
- **WHEN** a pipeline command reads `.opencode/`, `.specify/`, or
  `openspec/`
- **THEN** the files are present and readable at their standard paths

### Requirement: Failure atomicity

Stealth mode MUST never leave the working tree in a half-applied state.
If any step fails, tracked files MUST be restored to their pre-init
state (bytes and mode/permission bits), any scaffolded files written
MUST be removed, only the marker-delimited exclusion block appended by
the current invocation MUST be removed (preserving any pre-existing
block), the index MUST be restored to its init-start state (paths
newly staged during this init — currently staged minus staged at init
start — MUST be un-staged, never a blanket `git reset`), and a clear
error MUST be reported.

#### Scenario: failure mid-init

- **GIVEN** a failure occurs part-way through `uf init --stealth`
- **WHEN** the command returns an error
- **THEN** `.gitignore` and `AGENTS.md` are byte-identical to their
  pre-init state
- **AND** `git status --porcelain` returns no output
- **AND** the command returns a non-nil error with a non-empty message

### Requirement: Observability and verification

Stealth mode MUST provide a `--check` verification path that reports
the expected working-tree state without persisting changes, and a
post-init summary that reports whether the tree is git-clean. Both the
`--check` output and the post-init summary MUST emit a
machine-parseable (JSON) result. `--check` MUST also be re-runnable
after init (still combined with `--stealth`), detecting a missing or
stale `.git/info/exclude` block and reporting which scaffolded paths
have become visible.

The JSON result MUST include provenance fields (`tool`, `version`,
`timestamp`, and — when available — `ref`/`commit`) per Constitution
III.

`--check` MUST exit non-zero when the tree is not clean or the
exclusion block is missing/stale, and MUST exit 0 only when the
invariant holds. The post-init summary MUST likewise signal (via a
non-zero exit or a JSON `"clean": false` field) when it detects
residual dirty paths.

#### Scenario: check reports cleanliness without writing

- **GIVEN** a git repository
- **WHEN** the user runs `uf init --stealth --check`
- **THEN** the command reports (in JSON) whether the tree would be
  clean
- **AND** writes no scaffolded or tracked files

#### Scenario: check detects a lost exclusion

- **GIVEN** a stealth init has completed and `.git/info/exclude` has
  since been wiped
- **WHEN** the user runs `uf init --stealth --check`
- **THEN** the command reports that scaffolded paths have become
  visible

#### Scenario: non-portability is documented

- **GIVEN** a user runs `uf init --help`
- **WHEN** the help text is inspected
- **THEN** the `--stealth` flag help documents that the exclusion is
  local-only and must be re-created with `uf init --stealth`

#### Scenario: normal init warns about stale stealth exclusions

- **GIVEN** a repository where a prior `uf init --stealth` left a
  marker-delimited `.git/info/exclude` block
- **WHEN** the user runs a normal `uf init` (without `--stealth`)
- **THEN** the command warns that the stale stealth exclusions will
  hide scaffolded files from `git add`

### Requirement: Flag interactions

The `--stealth` flag MUST compose deterministically with existing flags.
`--force` MUST NOT override the already-tracked scaffolded-path
refusal; tracked scaffolded paths are always refused.

#### Scenario: interaction matrix

- **GIVEN** a git repository
- **WHEN** `uf init --stealth --force` runs
- **THEN** scaffolding is written and hidden (overwrite existing
  *untracked* scaffolded files + hide)
- **AND** `uf init --stealth --divisor` scaffolds only the divisor
  subset and hides it
- **AND** `uf init --stealth --lang <lang>` honors the language and
  hides scaffolding
- **AND** `uf init --stealth --dry-run` suppresses `opencode.json`
  writes as before while still hiding scaffolded files
- **AND** `uf init --stealth --check` persists nothing
- **AND** `uf init --check` (without `--stealth`) is a clear error
- **AND** existing flags without `--stealth` retain their current
  behavior

#### Scenario: force does not bypass tracked-path refusal

- **GIVEN** a git repository that already tracks `opencode.json`
- **WHEN** the user runs `uf init --stealth --force`
- **THEN** the command refuses with a clear error naming the tracked
  path
- **AND** writes no files
