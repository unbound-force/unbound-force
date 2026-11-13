<!-- capability: ref-resolution
     change: ref-lookup
     Resolve a ref name to an immutable commit object name.
     Design rationale: design.md (D1-D4). -->

Requirement numbering: RR-FR-001 through RR-FR-005 are allocated
for this change. RR-FR-006 through RR-FR-020 are reserved for
future ref-resolution changes. This is the capability's first
delta; no other range is allocated. The sibling
artifact-provenance capability numbers its requirements AP-FR-NNN
(specs/artifact-provenance/spec.md in this change).

## ADDED Requirements

### Requirement: Ref Resolution

RR-FR-001: `refs.Lookup(dir, ref)` MUST return the object name of
the commit `ref` points at in the repository at `dir`, after
`refs.Normalize`. Annotated tags MUST be peeled to their commit.

RR-FR-002: Lookup MUST accept full object names of both SHA-1
(40 hex digits) and SHA-256 (64 hex digits) repositories, and
MUST reject any other git output with an error.

RR-FR-003: Lookup MUST pass the ref to git as one argument after
`--end-of-options` and MUST NOT invoke a shell, so no ref value
can be interpreted as an option or a command.

RR-FR-004: Lookup MUST select the repository by `dir` alone; ambient
`GIT_*` environment variables MUST NOT redirect it.

RR-FR-005: Every Lookup error MUST name the ref, and when git fails
the error MUST include git's diagnostic output.

#### Scenario: Branch name in any accepted spelling
- **GIVEN** a repository whose `main` branch points at commit C
- **WHEN** Lookup is called with `main`, `refs/heads/main`, or
  ` main\n`
- **THEN** it returns C's object name

#### Scenario: Annotated tag
- **GIVEN** an annotated tag `v1` on commit C
- **WHEN** Lookup is called with `v1`
- **THEN** it returns C's object name, not the tag object's

#### Scenario: SHA-256 repository
- **GIVEN** a repository initialized with
  `--object-format=sha256`
- **WHEN** Lookup resolves `main`
- **THEN** it returns the 64-digit object name of the commit

#### Scenario: Hostile ref name
- **GIVEN** a ref such as `--output=<path>`, `main; touch <path>`
  or `$(touch <path>)`
- **WHEN** Lookup is called with it
- **THEN** it returns an error and `<path>` is not created

#### Scenario: Ambient GIT_DIR
- **GIVEN** `GIT_DIR` points at another repository
- **WHEN** Lookup resolves `main` in `dir`
- **THEN** it returns the commit from `dir`

#### Scenario: Unknown ref
- **GIVEN** no ref named `does-not-exist`
- **WHEN** Lookup is called with it
- **THEN** the error names the ref and carries git's message
