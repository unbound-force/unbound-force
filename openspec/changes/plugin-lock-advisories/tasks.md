<!--
  [P] marks tasks eligible for parallel execution.
  Item numbers in parentheses map to the findings in issue #991.
-->

## 1. Drift sweeps and acceptance tests

- [x] 1.1 (#1) Extend `TestPluginTestNpmPin_TwinsCISecurityPin` and
  `TestOsvRecipeDoc_PinsMatchCanonicalJob` to sweep
  `.github/workflows/ci_scheduled.yml`, and add
  `TestOsvPluginTransitive_ScheduledStepsMatchCanonical` (design D1)
- [x] 1.2 (#11) Record the recipe sweep's pins-not-flags scope in
  the test doc comment and the recipe Gotchas (design D5)
- [x] 1.3 (#2) Add `TestSCFR004_OfflineInstallFailure` asserting the
  install command and the pinned npm 11 escape hatch
- [x] 1.4 [P] (#9) Use `containsPath` in
  `TestSCFR003_RunPreservesTargetDevDependencies`

## 2. Doctor

- [x] 2.1 (#3) Run the tracked-lock git call under `LC_ALL=C` and
  line-anchor the not-a-repository match, with locale and spoofing
  test cases (design D2)
- [x] 2.2 [P] (#4) Add the `os.ReadFile` nil fallback to
  `pluginManifestHasDevDependencies`, with a test

## 3. Recorded decisions and traceability

- [x] 3.1 (#5) Record the GOTOOLCHAIN patch-staleness window in
  `docs/osv-transitive-scan.md` (design D4)
- [x] 3.2 (#6) Record the decision to keep the remediation constants
  as twinned pairs (design D3). DEFERRED: the shared internal
  package itself; it would retire the source-regex drift assertion
  in `TestRemediationConstants_CarryNpm11EscapeHatch`, which is a
  governance-surface change outside this stricter-only follow-up
- [x] 3.3 (#7) Fold the scheduled job into
  `stop-shipping-plugin-lock` tasks.md 4.4 and its "Lock-based
  scanners" risk bullet
- [x] 3.4 (#10) Reword the `stop-shipping-plugin-lock` SC-FR-004
  offline scenario to invariant wording
- [x] 3.5 (#8) Describe the daily transitive job in the AGENTS.md CI
  table's Scheduled row
- [x] 3.6 CHANGELOG entry

## 4. Verification

- [x] 4.1 `make check` passes (lint, test, build, coverage-gate,
  plugin-test). Locally: lint, test, build, coverage-gate green;
  plugin-test's real-git TS tests hit the 5s vitest timeout under
  host load (load average ~20) and passed when re-run on their own.
  No TS sources changed; CI is the authoritative run
