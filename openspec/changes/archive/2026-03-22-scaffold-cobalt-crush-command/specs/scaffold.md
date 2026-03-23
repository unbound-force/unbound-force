## ADDED Requirements

### Requirement: scaffold-cobalt-crush-command

The `uf init` command MUST deploy a
`cobalt-crush.md` command file to
`.opencode/command/cobalt-crush.md` in the target
project directory.

#### Scenario: fresh scaffold includes cobalt-crush

- **GIVEN** a fresh directory with no existing scaffold
- **WHEN** the developer runs `uf init`
- **THEN** the file `.opencode/command/cobalt-crush.md`
  is created with the Cobalt-Crush command instructions

#### Scenario: re-scaffold updates cobalt-crush

- **GIVEN** a project previously scaffolded with an
  older version of the `cobalt-crush.md` file
- **WHEN** the developer runs `uf init` again
- **THEN** the tool-owned `cobalt-crush.md` file is
  overwritten with the current version

### Requirement: scaffold-file-count-updated

The scaffold file count test assertion MUST reflect the
addition of the new scaffold asset.

#### Scenario: file count test passes

- **GIVEN** the scaffold engine with the new
  `cobalt-crush.md` asset added
- **WHEN** `go test -race -count=1 ./...` is run
- **THEN** all tests pass including the file count
  assertion

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
