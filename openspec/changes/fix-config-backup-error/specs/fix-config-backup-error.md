## MODIFIED Requirements

### Requirement: Config backup write error MUST be checked

`InitFile` MUST check the return value of the backup write
and return an error if it fails, before attempting to
overwrite the live config file.

Previously: the backup write error was discarded with `_ =`.

#### Scenario: Backup write fails before live config overwrite
- **GIVEN** a `.uf/config.yaml` exists and requires an update
- **WHEN** writing the `.bak` backup file fails (e.g., disk
  full, permission denied)
- **THEN** `InitFile` MUST return a non-nil error wrapping
  the message `"write backup config"`
- **AND** the live config file MUST NOT be modified

#### Scenario: Backup write succeeds, live write proceeds
- **GIVEN** a `.uf/config.yaml` exists and requires an update
- **WHEN** writing the `.bak` backup file succeeds
- **THEN** `InitFile` MUST proceed to overwrite the live
  config file as before
- **AND** `InitResult.Updated` MUST be true on success

## ADDED Requirements

### Requirement: Regression test MUST cover backup write failure

A test named `TestInitFile_BackupWriteFailureAbortsUpdate`
MUST exist in `internal/config/init_test.go`.

The test MUST:
- Inject a `WriteFile` stub via `InitOptions.WriteFile` that
  returns an error for `.bak` paths and delegates to
  `writeFileAtomic` for all other paths
- Assert that `InitFile` returns a non-nil error
- Assert the error message contains `"write backup config"`
- Assert the original config file content is unchanged after
  the failed call
