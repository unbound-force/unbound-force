## ADDED Requirements

### Requirement: Cleanup error logging

`cleanupRenamedCommands()` MUST log a warning for each
old-name file that exists but cannot be removed. The
warning MUST include the file path and the error message
from `os.Remove`.

#### Scenario: File removal fails due to permissions

- **GIVEN** a target directory contains an old-name
  command file `address-feedback.md` that is read-only
- **WHEN** `cleanupRenamedCommands()` runs
- **THEN** a warning is written to the output writer
  containing the file path and the permission error
- **AND** the function continues processing remaining
  files without aborting

#### Scenario: File removal succeeds

- **GIVEN** a target directory contains old-name command
  files `address-feedback.md` and `review-council.md`
- **WHEN** `cleanupRenamedCommands()` runs
- **THEN** both files are removed
- **AND** the returned slice contains both mapped output
  paths (`.opencode/commands/address-feedback.md`,
  `.opencode/commands/review-council.md`)

### Requirement: Stale command reference warning

`uf init` MUST warn when agent files in
`.opencode/agents/` contain references to old (pre-
namespace-prefix) command names. The warning MUST list
each affected file and the stale references found,
along with the correct replacement names.

#### Scenario: Agent file contains stale reference

- **GIVEN** `.opencode/agents/cobalt-crush-dev.md`
  contains the text `/review-council`
- **WHEN** `uf init` runs
- **THEN** a warning is printed listing
  `cobalt-crush-dev.md` with the stale reference
  `/review-council` and replacement `/uf.review-council`

#### Scenario: No agent files contain stale references

- **GIVEN** all agent files use the current `uf.`-prefixed
  command names
- **WHEN** `uf init` runs
- **THEN** no stale reference warning is printed

#### Scenario: No agent files exist

- **GIVEN** the target directory has no files in
  `.opencode/agents/`
- **WHEN** `uf init` runs
- **THEN** no stale reference warning is printed and no
  error occurs

### Requirement: Test coverage for cleanup path

`cleanupRenamedCommands()` MUST have test coverage for:
- Happy path (files exist and are removed)
- No-op path (no old files exist)
- Partial failure (some files removable, some not)
- Returned path format verification

`warnStaleCommandRefs()` MUST have test coverage for:
- Agent file with stale references
- Agent file with no stale references
- Empty agent directory

All tests MUST use `t.TempDir()` for filesystem
isolation and `bytes.Buffer` for output capture.
New functions (`cleanupRenamedCommands`,
`warnStaleCommandRefs`) MUST achieve ≥80% line
coverage as measured by `go test -coverprofile`.

#### Scenario: Test isolation

- **GIVEN** a test for `cleanupRenamedCommands()`
- **WHEN** the test creates files in `t.TempDir()`
- **THEN** no files outside the temp directory are read,
  modified, or deleted

## MODIFIED Requirements

### Requirement: cleanupRenamedCommands signature

`cleanupRenamedCommands()` MUST accept an `io.Writer`
parameter for warning output, in addition to the
existing `targetDir string` parameter.

Previously: `func cleanupRenamedCommands(targetDir string)
[]string` accepted only a target directory.

### Requirement: Scaffold summary includes migration info

The scaffold summary printed by `printSummary()` MUST
include migrated files in the file categories output.

Previously: The `Result.Migrated` field existed but
migrated files could be empty due to silent cleanup
failures. With error logging, the summary now
accurately reflects what was and was not cleaned up.

## REMOVED Requirements

None.
