## ADDED Requirements

### Requirement: mxf included in RPM package

The `nfpms:` configuration MUST include the `mxf`
binary in the `unbound-force` RPM package, installed
to `/usr/bin/mxf`.

#### Scenario: RPM contains mxf binary

- **GIVEN** a tagged release is pushed
- **WHEN** GoReleaser generates the RPM package
- **THEN** the RPM contains both `/usr/bin/unbound-force`
  and `/usr/bin/mxf`

---

### Requirement: mxf included in Homebrew Formula

The `brews:` install block MUST include
`bin.install "mxf"` so that the Homebrew Formula
installs both binaries.

#### Scenario: Formula installs mxf

- **GIVEN** a user runs
  `brew install unbound-force/tap/unbound-force`
- **WHEN** Homebrew installs the Formula
- **THEN** both `unbound-force` and `mxf` are available
  in the PATH

---

## MODIFIED Requirements

### Requirement: installMxF verifies presence

Previously: `installMxF()` attempted
`brew install unbound-force/tap/mxf` (a non-existent
package).

Now: `installMxF()` MUST verify that `mxf` is in PATH
via `LookPath`. If present, report "already installed".
If absent, report as a note that `mxf` is bundled with
`unbound-force` and suggest reinstalling the parent
package.

`installMxF()` MUST NOT attempt to run
`brew install unbound-force/tap/mxf`.

#### Scenario: mxf found in PATH

- **GIVEN** `mxf` is available in PATH
- **WHEN** `uf setup` runs the Mx F step
- **THEN** the step reports "already installed"

#### Scenario: mxf not found in PATH

- **GIVEN** `mxf` is NOT available in PATH
- **WHEN** `uf setup` runs the Mx F step
- **THEN** the step reports "not found" with a hint to
  install `unbound-force` (which bundles `mxf`)

---

### Requirement: mxf install hints reference parent

Previously: `homebrewInstallCmd("mxf")` returned
`brew install unbound-force/tap/mxf`.

Now: Install hints for `mxf` MUST reference the
`unbound-force` package:
- Homebrew: `brew install unbound-force/tap/unbound-force`
- Generic: `Bundled with unbound-force`

#### Scenario: doctor hint for mxf

- **GIVEN** `mxf` is not in PATH
- **WHEN** `uf doctor` generates an install hint
- **THEN** the hint references `unbound-force` (not a
  non-existent `mxf` package)

---

## REMOVED Requirements

None.
