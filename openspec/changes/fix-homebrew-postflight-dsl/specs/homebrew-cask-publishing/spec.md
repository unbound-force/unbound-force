## ADDED Requirements

### Requirement: Transform Deprecated Cask Hook Syntax

The release publishing process MUST transform the generated
`unbound-force` Cask's expected legacy `postflight` hook into a
`postflight_steps` block before copying the Cask to the Homebrew tap.
The transformation MUST preserve the existing quarantine-removal and
`uf` symlink behaviors using supported declarative install steps.

The transformer MUST accept exactly this legacy block:

```ruby
postflight do
  if OS.mac?
    system_command "/usr/bin/xattr", args: ["-dr", "com.apple.quarantine", "#{staged_path}/unbound-force"]
    # Create uf symlink alias for daily-use convenience (FR-002/FR-009).
    system_command "/bin/ln", args: ["-sf", "#{staged_path}/unbound-force", "#{HOMEBREW_PREFIX}/bin/uf"]
  end
end
```

It MUST replace that block with exactly this supported declarative
block:

```ruby
postflight_steps do
  on_macos do
    run "/usr/bin/xattr",
      args: ["-dr", "com.apple.quarantine", "{{staged_path}}/unbound-force"]
    symlink "unbound-force", "{{HOMEBREW_PREFIX}}/bin/uf", overwrite: true
  end
end
```

#### Scenario: Transform the expected generated Cask
- **GIVEN** a generated `unbound-force` Cask containing exactly one
  literal legacy post-install block
- **WHEN** the release pipeline prepares the Cask for tap publication
- **THEN** the output MUST contain exactly one literal replacement block
- **AND** the output MUST NOT contain `postflight do`
- **AND** the output MUST retain install steps that remove quarantine
  from `unbound-force` and create the `uf` symlink.

### Requirement: Fail Closed on Unexpected Generated Cask Content

The transformation MUST fail before tap publication when the expected
legacy hook is absent, occurs more than once, or does not match the
known generated structure. The failure MUST identify that generated
Cask hook transformation could not be completed.

#### Scenario: Reject a Cask without the expected legacy hook
- **GIVEN** a generated Cask that does not contain the expected legacy
  post-install hook
- **WHEN** the release pipeline prepares the Cask for tap publication
- **THEN** the transformation MUST fail
- **AND** the workflow MUST NOT copy or publish that Cask.

#### Scenario: Reject an ambiguous Cask hook
- **GIVEN** a generated Cask containing multiple matching legacy hooks
- **WHEN** the release pipeline prepares the Cask for tap publication
- **THEN** the transformation MUST fail
- **AND** the workflow MUST NOT select a hook arbitrarily.

### Requirement: Cover Transformer Decision Contracts

The transformer unit suite MUST cover all five defined transformation
contracts: exact single-hook replacement, absent-hook rejection,
duplicate-hook rejection, near-match rejection, and preservation of
surrounding Cask content on successful replacement. The success case
MUST compare against exact expected Cask output. Each rejection case
MUST assert a contextual error and that no output file is created or
changed. This test-contract target MUST NOT modify existing CI coverage
thresholds.

#### Scenario: Verify complete transformer contract coverage
- **GIVEN** fixture Casks for each defined transformation contract
- **WHEN** the transformer unit suite runs
- **THEN** each contract MUST have an explicit test case
- **AND** the success fixture output MUST match exactly
- **AND** each rejection fixture MUST leave output absent or unchanged.

### Requirement: Verify Transformed Cask Content

The release pipeline MUST validate the transformed Cask before tap
publication. Validation MUST assert the supported stanza is present,
the deprecated stanza is absent, and both required install behaviors
are represented. The macOS release job MUST run Homebrew static
validation and an isolated staged-tap install/uninstall smoke test of
the transformed Cask. Existing signed-checksum validation and release
gates MUST remain enforced.

#### Scenario: Publish a validated transformed Cask
- **GIVEN** the transformation succeeds for the generated Cask
- **WHEN** the macOS release job validates and installs the staged Cask
- **THEN** publication MAY continue only when all transformation
  assertions, Homebrew validation, and smoke-test assertions pass
- **AND** existing checksum patching and validation MUST still run.

#### Scenario: Stop publication after failed semantic validation
- **GIVEN** a transformed staged Cask that fails Homebrew static
  validation or the staged-tap install smoke test
- **WHEN** the macOS release job prepares the tap update
- **THEN** the job MUST fail before copying the Cask to the published
  tap path or executing `git push`
- **AND** the failed staged Cask MUST NOT be published.

### Requirement: Preserve Release-Publication Ordering

The release workflow MUST preserve the order of signed-checksum
patching, hook transformation, Homebrew validation, staged-tap smoke
testing, tap copy, and tap publication. Automated workflow coverage
MUST verify that transformer or validation failure prevents tap
publication.

#### Scenario: Verify fail-closed workflow ordering
- **GIVEN** the release workflow definition
- **WHEN** its release-publishing path is tested
- **THEN** the test MUST verify transformation and semantic validation
  occur before the tap copy and `git push` operations
- **AND** the test MUST verify a failure in either preceding stage makes
  publication unreachable.

## MODIFIED Requirements

- None.

## REMOVED Requirements

- None.
