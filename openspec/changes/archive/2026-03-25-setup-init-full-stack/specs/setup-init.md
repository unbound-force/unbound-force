## ADDED Requirements

### Requirement: setup-install-mxf

`uf setup` SHOULD install the Mx F Manager hero binary
via Homebrew when not present, and MUST degrade
gracefully when Homebrew is unavailable.

#### Scenario: mxf not installed

- **GIVEN** `mxf` is not in PATH and Homebrew is available
- **WHEN** `uf setup` runs
- **THEN** `mxf` is installed via `brew install unbound-force/tap/mxf`

#### Scenario: mxf already installed

- **GIVEN** `mxf` is already in PATH
- **WHEN** `uf setup` runs
- **THEN** the step reports "already installed"

#### Scenario: no Homebrew

- **GIVEN** `mxf` is not in PATH and Homebrew is
  not available
- **WHEN** `uf setup` runs
- **THEN** the step is skipped with a GitHub releases
  download link

#### Scenario: dry-run

- **GIVEN** `--dry-run` is set
- **WHEN** `uf setup` runs
- **THEN** the step reports "Would install" without
  executing

### Requirement: setup-install-gh

`uf setup` SHOULD install the GitHub CLI via Homebrew
when not present, and MUST degrade gracefully when
Homebrew is unavailable.

#### Scenario: gh not installed

- **GIVEN** `gh` is not in PATH and Homebrew is available
- **WHEN** `uf setup` runs
- **THEN** `gh` is installed via `brew install gh`

#### Scenario: gh already installed

- **GIVEN** `gh` is already in PATH
- **WHEN** `uf setup` runs
- **THEN** the step reports "already installed"

#### Scenario: no Homebrew

- **GIVEN** `gh` is not in PATH and Homebrew is not
  available
- **WHEN** `uf setup` runs
- **THEN** the step is skipped with a download link
  to https://cli.github.com

#### Scenario: dry-run

- **GIVEN** `--dry-run` is set
- **WHEN** `uf setup` runs
- **THEN** the step reports "Would install" without
  executing

### Requirement: setup-install-openspec

`uf setup` SHOULD install the OpenSpec CLI via bun or
npm when not present, and MUST degrade gracefully when
neither is available.

#### Scenario: openspec not installed

- **GIVEN** `openspec` is not in PATH and bun or npm
  is available
- **WHEN** `uf setup` runs
- **THEN** `openspec` is installed via
  `bun add -g @fission-ai/openspec@latest` (preferred)
  or `npm install -g @fission-ai/openspec@latest`
  (fallback)

#### Scenario: openspec already installed

- **GIVEN** `openspec` is already in PATH
- **WHEN** `uf setup` runs
- **THEN** the step reports "already installed"

#### Scenario: no Node.js available

- **GIVEN** `openspec` is not in PATH and Node.js is
  not available
- **WHEN** `uf setup` runs
- **THEN** the OpenSpec step is skipped

#### Scenario: npm/bun install fails

- **GIVEN** `openspec` is not in PATH and npm is
  available but install fails (EACCES, network error)
- **WHEN** `uf setup` runs
- **THEN** the step reports "failed" with actionable
  hint

#### Scenario: dry-run

- **GIVEN** `--dry-run` is set
- **WHEN** `uf setup` runs
- **THEN** the step reports "Would install" without
  executing

### Requirement: setup-dewey-init

`uf setup` MUST run `dewey init` after the Dewey
binary and embedding model are installed, if `.dewey/`
does not exist.

#### Scenario: dewey installed, workspace missing

- **GIVEN** `dewey` is in PATH and `.dewey/` does not
  exist
- **WHEN** `uf setup` runs
- **THEN** `dewey init` is executed and `.dewey/` is
  created

#### Scenario: workspace already exists

- **GIVEN** `.dewey/` already exists
- **WHEN** `uf setup` runs
- **THEN** the step reports "already initialized"

#### Scenario: dewey not installed

- **GIVEN** `dewey` is not in PATH
- **WHEN** `uf setup` runs
- **THEN** the step is skipped

#### Scenario: dewey init fails

- **GIVEN** `dewey` is in PATH and `.dewey/` does not
  exist and `dewey init` returns an error
- **WHEN** `uf setup` runs
- **THEN** the step reports "failed" and dewey index
  is skipped

#### Scenario: dry-run

- **GIVEN** `--dry-run` is set
- **WHEN** `uf setup` runs
- **THEN** the step reports "Would run: dewey init"
  without executing

### Requirement: setup-dewey-index

`uf setup` MUST run `dewey index` after the Dewey
workspace is initialized.

#### Scenario: workspace exists

- **GIVEN** `.dewey/` exists and `dewey` is in PATH
- **WHEN** `uf setup` runs
- **THEN** `dewey index` is executed

#### Scenario: workspace does not exist

- **GIVEN** `.dewey/` does not exist
- **WHEN** `uf setup` runs
- **THEN** the dewey index step is skipped

#### Scenario: dewey not installed

- **GIVEN** `dewey` is not in PATH
- **WHEN** `uf setup` runs
- **THEN** the dewey index step is skipped

#### Scenario: index fails

- **GIVEN** `.dewey/` exists and `dewey index` returns
  an error (e.g., Ollama not running)
- **WHEN** `uf setup` runs
- **THEN** the step reports "failed" with hint to run
  `ollama serve`

#### Scenario: dry-run

- **GIVEN** `--dry-run` is set
- **WHEN** `uf setup` runs
- **THEN** the step reports "Would run: dewey index"
  without executing

### Requirement: init-sub-tools

`uf init` MUST initialize Dewey when the binary is
available and `.dewey/` does not exist. Errors are
reported as warnings, not hard failures.

#### Scenario: dewey available, no workspace

- **GIVEN** `dewey` is in PATH and `.dewey/` does not
  exist
- **WHEN** `uf init` runs
- **THEN** `dewey init` and `dewey index` are executed
  after scaffolding

#### Scenario: dewey not available

- **GIVEN** `dewey` is not in PATH
- **WHEN** `uf init` runs
- **THEN** sub-tool initialization is skipped and
  printSummary suggests running `uf setup`

#### Scenario: dewey init fails during uf init

- **GIVEN** `dewey` is in PATH and `.dewey/` does not
  exist and `dewey init` returns an error
- **WHEN** `uf init` runs
- **THEN** scaffolding succeeds, printSummary shows a
  warning for the failed sub-tool init, and `uf init`
  returns no error

#### Scenario: DivisorOnly mode

- **GIVEN** `uf init --divisor` is used
- **WHEN** Dewey is available
- **THEN** sub-tool initialization is skipped

### Requirement: init-next-steps

`uf init` printSummary MUST show actionable next steps
including constitution creation, `uf doctor`, and
workflow commands.

#### Scenario: full toolchain available

- **GIVEN** all tools are installed
- **WHEN** `uf init` completes
- **THEN** printSummary shows: constitution, doctor,
  speckit, and opsx next steps

#### Scenario: tools missing

- **GIVEN** some tools are not installed
- **WHEN** `uf init` completes
- **THEN** printSummary shows `uf setup` as the first
  next step

## MODIFIED Requirements

### Requirement: scaffold-options-expanded

The `scaffold.Options` struct MUST include `LookPath`
and `ExecCmd` fields for testable sub-tool
initialization. These MUST be defaulted to production
implementations if nil, and the defaulting MUST happen
at the top of `Run()` before any code path that calls
`initSubTools()`.

## REMOVED Requirements

None.
