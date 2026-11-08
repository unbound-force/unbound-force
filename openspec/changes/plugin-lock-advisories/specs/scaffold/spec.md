## Scaffold delta: plugin-lock-advisories

Follow-up to `stop-shipping-plugin-lock` (issue #991). That change
is unarchived, so its SC-FR-004 offline-install scenario is
reworded in place in its own delta (invariant wording: runner
class and failure mode, not the mutable npx version). This delta
allocates no SC-FR ids.

## ADDED Requirements

### Requirement: Throwaway-lock scan pin sweep

Each throwaway-lock OSV scan copy MUST match its twins. The copies
run by this repository (`osv_plugin_transitive` in
`.github/workflows/ci_security.yml` and
`.github/workflows/ci_scheduled.yml`) MUST declare the same npm
runner, `GOTOOLCHAIN`, and osv-scanner pins as the Makefile
`PLUGIN_TEST_NPM` twin and the copyable recipe in
`docs/osv-transitive-scan.md`, and the two CI copies MUST carry
byte-identical `steps:` blocks. Drift MUST fail a test.

#### Scenario: Scheduled job drifts from the canonical job

- **GIVEN** `ci_scheduled.yml` pins a different osv-scanner
  version than `ci_security.yml`
- **WHEN** the Go test suite runs
- **THEN** a pin-sweep test fails and names the scheduled workflow

### Requirement: Offline activation acceptance test

A scenario-named test MUST cover SC-FR-004. The test
(`TestSCFR004_*`) MUST exercise the SC-FR-004 offline install failure end to end through `Run()` and
assert that no plugin source is auto-discoverable and that the
remediation names both the lockless install command and the pinned
npm 11 escape hatch.

#### Scenario: Offline install surfaces both remedies

- **GIVEN** the activation install fails
- **WHEN** `uf init` reports the failure
- **THEN** the output names
  `npm install --ignore-scripts --omit=dev --no-package-lock`
- **AND** the output names the pinned npm 11 runner

### Requirement: Locale-stable tracked-lock classification

The tracked-lock check MUST be locale-stable. `uf doctor` MUST
classify "not a git repository" independently of the user's locale, and a `TargetDir`
that embeds the signature MUST NOT be classified as outside a
repository.

#### Scenario: Path embeds the signature

- **GIVEN** git fails with a message whose path contains
  `fatal: not a git repository` mid-line
- **WHEN** `uf doctor` runs the tracked-lock check
- **THEN** the check warns with the git failure instead of passing
