## ADDED Requirements

### Requirement: Build from Source subsection in README

The README `## Getting Started` section MUST include a "Build from
Source" subsection positioned after the Fedora/RHEL subsection. The
subsection MUST document:

1. Prerequisites (Go 1.25+, make)
2. Clone command (`git clone`)
3. Install command (`make install`)
4. What the install target does (builds binary + creates `uf` symlink)
5. Verification step (`uf version`)

The subsection SHOULD note that `$GOPATH/bin` must be on `$PATH`.

#### Scenario: User builds unbound-force from source via README

- **GIVEN** a user has Go 1.25+ and make installed
- **WHEN** they follow the "Build from Source" README instructions
  (clone the repo, run `make install`)
- **THEN** both `unbound-force` and `uf` commands are available on
  their `$PATH` and `uf version` outputs version information

### Requirement: Build from Source subsection in QUICKSTART.md

The QUICKSTART.md `## Install` section MUST include a "Build from
Source" subsection positioned after the "Fedora / RHEL (dnf --
minimal)" subsection and before "## For Project Maintainers". The
subsection MUST document the same information as the README
subsection.

#### Scenario: User builds unbound-force from source via QUICKSTART

- **GIVEN** a user reads QUICKSTART.md for installation options
- **WHEN** they reach the "Build from Source" subsection
- **THEN** they find prerequisites, build command, and verification
  step consistent with the README subsection and with other UF
  component documentation

#### Scenario: User without $GOPATH/bin on PATH

- **GIVEN** a user has Go installed but `$GOPATH/bin` is not on
  their `$PATH`
- **WHEN** they read the "Build from Source" subsection
- **THEN** they see a note about ensuring `$GOPATH/bin` is on
  their `$PATH`

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
<!-- scaffolded by uf vdev -->
