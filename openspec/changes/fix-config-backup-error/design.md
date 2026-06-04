## Context

`internal/config/init.go:InitFile` handles the create-or-update
lifecycle of `.uf/config.yaml`. When an update is needed, it
writes a `.bak` backup before overwriting the live file. The
backup write uses `_ = opts.WriteFile(...)`, silently dropping
any error.

The `InitOptions.WriteFile` field is already a function
injection point used by tests. This makes the fix trivially
testable without any refactoring of the external interface.

## Goals / Non-Goals

### Goals
- Surface backup write failures as returned errors.
- Prevent live config overwrite when backup has failed.
- Add a regression test using the existing injection point.

### Non-Goals
- Changing the backup file naming or location.
- Adding logging or warnings as an alternative to aborting
  (aborting is the safer and simpler choice for this case).
- Modifying any other error handling in `InitFile`.

## Decisions

**Abort on backup failure, do not warn-and-continue.**
The backup's only purpose is to protect the user's config
before a destructive write. If the backup cannot be written,
continuing the overwrite defeats its purpose entirely.
Returning an error and leaving the live file untouched is the
only safe behavior.

**Error message: `"write backup config: %w"`.**
Consistent with the existing error message style in `InitFile`
(`"write updated config: %w"`, `"write config: %w"`). Callers
can match on `"write backup config"` in tests and error
messages distinguish this failure from the live-write failure.

**No new types, interfaces, or packages.**
The fix is a one-line change from `_ =` to `if err := ...; err
!= nil { return nil, fmt.Errorf(...) }`. The existing
`InitOptions.WriteFile` injection point is sufficient for the
regression test.

## Risks / Trade-offs

**Risk**: A backup write failure that was previously invisible
will now surface as an error to the user. In theory this could
expose a pre-existing environment problem (e.g., a read-only
`.uf/` directory) that the user was unaware of. This is
acceptable — surfacing the problem is better than silently
corrupting the config.

**Trade-off**: None. The fix is strictly safer than the
original code.
