## ADDED Requirements

None.

## MODIFIED Requirements

### Requirement: uf setup step count

Previously: `uf setup` runs 14 steps, with step 3 being
the Mx F binary check (`installMxF`).

New: `uf setup` MUST run 13 steps. The Mx F binary check
MUST be removed. Steps previously numbered 4-14 MUST be
renumbered to 3-13. Step progress output (e.g.,
`[3/13] GitHub CLI...`) MUST reflect the new count.

#### Scenario: Setup runs without mxf step

- **GIVEN** a user runs `uf setup`
- **WHEN** the setup reaches the tool installation steps
- **THEN** there is no step for Mx F, and the total step
  count displayed is 13

### Requirement: uf doctor core tools

Previously: `uf doctor` includes `mxf` in the `coreTools`
slice with `recommended: true`, producing a `[WARN]` when
the binary is not found.

New: `uf doctor` MUST NOT include `mxf` in the `coreTools`
slice. The doctor MUST NOT warn about a missing `mxf`
binary. The hero availability check for Mx F (via agent
file detection) MUST remain unchanged.

#### Scenario: Doctor does not warn about mxf

- **GIVEN** a user runs `uf doctor`
- **WHEN** the `mxf` binary is not in PATH
- **THEN** no warning or error about `mxf` appears in
  the Core Tools group

#### Scenario: Hero availability still detects Mx F

- **GIVEN** `.opencode/agents/mx-f-coach.md` exists
- **WHEN** `uf doctor` checks hero availability
- **THEN** Mx F is reported as available via agent file

## REMOVED Requirements

### Requirement: installMxF function

The `installMxF()` function in `internal/setup/setup.go`
MUST be removed. Reason: the `mxf` binary is not built
or distributed. The function always returns "not found"
with a misleading hint that the binary is bundled with
`unbound-force`.

### Requirement: mxf install hint in doctor

The install hint `"brew install unbound-force/tap/
unbound-force (mxf is bundled)"` MUST be removed from
doctor output. Reason: the `mxf` binary is not bundled
in the `unbound-force` Homebrew formula.

### Requirement: mxf test functions

The test functions `TestSetupRun_MxFMissing_BundledHint`
and `TestSetupRun_MxFPresent` MUST be removed alongside
the `installMxF` function they test. `mxf` entries in
LookPath stubs of remaining tests MUST be removed to
avoid dead test setup code.
