## ADDED Requirements

### Requirement: OpenPackage post-walk delegation

FR-001 [MUST]: When `opkg` is detected on PATH, `uf init`
MUST run `opkg install` for each package (review-council
and, unless `--divisor` mode is active, workflows) after
the embedded asset walk completes. The embedded assets are
always written first (baseline); `opkg` then overlays
harness-specific content on top. Each package is installed
separately; if one fails, the loop stops and the summary
reports which package failed with a re-run hint.

#### Scenario: opkg available and install succeeds
- **GIVEN** `opkg` is on PATH and responds to `opkg install`
- **WHEN** `uf init` is executed
- **THEN** all embedded assets are written to disk first
- **AND** `opkg install <pkg>` runs once per package in
  the target directory
- **AND** the summary reports `opkg installed (<packages>)`

#### Scenario: opkg available but install fails
- **GIVEN** `opkg` is on PATH but `opkg install` returns
  a non-zero exit code
- **WHEN** `uf init` is executed
- **THEN** all embedded assets are written as normal
  (fallback baseline remains)
- **AND** the summary reports `opkg failed (<package>:
  <detail> — re-run uf init to retry)`

#### Scenario: opkg not on PATH
- **GIVEN** `opkg` is not on PATH
- **WHEN** `uf init` is executed
- **THEN** all embedded assets are written as normal
- **AND** no opkg-related line appears in the summary

### Requirement: ExecCmdInDir injectable

FR-003 [MUST]: `scaffold.Options` MUST include an
`ExecCmdInDir` field with signature
`func(dir string, name string, args ...string) ([]byte, error)`
that runs a command with the working directory set. The
default implementation wraps `exec.Command` with `cmd.Dir`.

#### Scenario: ExecCmdInDir is injectable for tests
- **GIVEN** a test creates `scaffold.Options` with a custom
  `ExecCmdInDir` function
- **WHEN** `Run()` calls `openPackageInstall()`
- **THEN** the custom function is called instead of the
  default `exec.Command` wrapper

### Requirement: SkipOpenPackage test flag

FR-004 [MUST]: `scaffold.Options` MUST include a
`SkipOpenPackage` boolean field. When set to `true`,
`openPackageInstall()` MUST return immediately without
checking for `opkg` or running any commands.

#### Scenario: SkipOpenPackage prevents delegation
- **GIVEN** `SkipOpenPackage` is `true` in Options
- **WHEN** `Run()` is called
- **THEN** `openPackageInstall()` returns no results
- **AND** all embedded assets are written normally

### Requirement: DivisorOnly package selection

FR-005 [MUST]: When `DivisorOnly=true`, `openPackageInstall()`
MUST install only the `review-council` package (not
`workflows`).

#### Scenario: DivisorOnly installs only review-council
- **GIVEN** `DivisorOnly` is `true` and `opkg` is on PATH
- **WHEN** `openPackageInstall()` runs
- **THEN** the install command includes only
  `<target>/.openpackage/packages/review-council`
- **AND** `workflows` is not included

### Requirement: OpenPackage source trees

FR-006 [MUST]: Two OpenPackage source trees MUST be embedded
in the binary and deployed to `.openpackage/packages/` during
`uf init`: `review-council/` and `workflows/`, each with
valid `openpackage.yml` manifests, multi-harness frontmatter
in agent/command files, and the standard opkg directory
convention:
```
packages/<name>/
  openpackage.yml
  README.md
  agents/<name>/*.md
  commands/<name>/*.md
  rules/<name>/*.md     (review-council only)
  mcp.jsonc             (review-council only)
```

#### Scenario: review-council package structure
- **GIVEN** the `.openpackage/packages/review-council/` directory
- **WHEN** its contents are enumerated
- **THEN** it contains `openpackage.yml` with name
  `@unbound-force/review-council`
- **AND** 9 Divisor agent files under
  `agents/review-council/`
- **AND** 2 command files under `commands/review-council/`
- **AND** 3 rule files under `rules/review-council/`
- **AND** `mcp.jsonc` and `README.md`

#### Scenario: workflows package structure
- **GIVEN** the `.openpackage/packages/workflows/` directory
- **WHEN** its contents are enumerated
- **THEN** it contains `openpackage.yml` with name
  `@unbound-force/workflows` and a dependency on
  `@unbound-force/review-council` at `^0.1.0`
- **AND** 1 agent file under `agents/workflows/`
- **AND** 14+ command files under `commands/workflows/`
- **AND** `README.md`

### Requirement: installOpkg setup step

FR-007 [SHOULD]: `uf setup` SHOULD include a step that
attempts `brew install openpackage`. On failure (formula
not published, Homebrew absent), the step MUST return
`"skipped"` with a manual-install hint, not a hard error.

#### Scenario: opkg already installed
- **GIVEN** `opkg` is already on PATH
- **WHEN** `uf setup` reaches the opkg step
- **THEN** it reports `"already installed"`

#### Scenario: Homebrew available but formula missing
- **GIVEN** Homebrew is available but `openpackage`
  formula does not exist
- **WHEN** `uf setup` reaches the opkg step
- **THEN** it reports `"skipped"` with a manual-install
  hint
- **AND** no error propagates to cause setup to fail

### Requirement: OpenPackagePlatforms passthrough

FR-008 [MUST]: `openPackageInstall()` MUST always pass
`--platforms` to each `opkg install` invocation. When
`Options.OpenPackagePlatforms` is empty, the default
value `"opencode"` MUST be used. When non-empty, the
provided value is used as-is.

#### Scenario: default platforms when flag not set
- **GIVEN** `OpenPackagePlatforms` is empty
- **WHEN** `openPackageInstall()` runs
- **THEN** each `opkg install` call includes
  `--platforms opencode`

#### Scenario: custom platforms flag passed to opkg
- **GIVEN** `OpenPackagePlatforms` is set to
  `"opencode,cursor"`
- **WHEN** `openPackageInstall()` runs
- **THEN** each `opkg install` call includes
  `--platforms opencode,cursor` as trailing arguments

### Requirement: Summary reporting

FR-009 [MUST]: When opkg delegation runs (success or
failure), the scaffold summary MUST include a sub-tool
result entry showing the outcome. Action values:
`"installed"` (success), `"failed"` (error with package
name and re-run hint), or absent (opkg not on PATH).

#### Scenario: opkg result in summary
- **GIVEN** `opkg` is on PATH and install succeeds
- **WHEN** `printSummary()` renders sub-tool results
- **THEN** a line appears: `✓ opkg installed (<packages>)`

### Requirement: Error detail truncation

FR-010 [SHOULD]: When `opkg install` fails, the error
detail in the subToolResult SHOULD be truncated to 200
characters to prevent summary output overflow. The
failing package name MUST be included in the detail.

#### Scenario: Long error is truncated
- **GIVEN** `opkg install` fails with a 500-character error
- **WHEN** the result detail is constructed
- **THEN** the detail is truncated to 200 characters
  followed by `"…"`
- **AND** the failing package base name appears first

## MODIFIED Requirements

(None)

## REMOVED Requirements

(None)
