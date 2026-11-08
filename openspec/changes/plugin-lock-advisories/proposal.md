# Proposal: plugin-lock-advisories

## Why

The review-council gate on PR #957 (`stop-shipping-plugin-lock`,
round 3) approved with eleven advisory findings. Per the
review-council rule they shipped as recorded deferrals and are
tracked in issue #991. The largest gap: the daily
`osv_plugin_transitive` job in `.github/workflows/ci_scheduled.yml`
is a third copy of the throwaway-lock scan pins (`npx -y
npm@11.6.2`, `GOTOOLCHAIN=go1.27.0`, `osv-scanner@v2.7.0`) that no
drift test sweeps, so it can silently fall out of step with the
push/PR job and the recipe hero repositories copy.

## What Changes

- **Pin sweeps cover the scheduled job.** The npm twin-pin test and
  the recipe-doc pin test also sweep `ci_scheduled.yml`, and a new
  parity test requires the scheduled job's `steps:` block to match
  the canonical `ci_security.yml` job byte for byte. Gates only get
  stricter.
- **SC-FR-004 gets a scenario-named acceptance test**
  (`TestSCFR004_OfflineInstallFailure`) that asserts the plugin
  sources stay out of auto-discovery and that the remediation names
  both the lockless install and the pinned npm 11 escape hatch.
- **Locale-stable git classification** in the `uf doctor`
  tracked-lock check: git runs under `LC_ALL=C`, and the
  not-a-repository signature must start a line of git's output, so
  translated messages and a `TargetDir` that embeds the phrase can
  no longer flip the classification.
- **Nil `ReadFile` fallback** in `pluginManifestHasDevDependencies`,
  matching its sibling seams.
- **Recorded decisions**: the accepted GOTOOLCHAIN patch-staleness
  window (docs and design), the pins-versus-flags scope of the
  recipe sweep (test comment and docs), and the decision to keep the
  four remediation constants as two twinned pairs rather than a
  shared package (design D3).
- **Traceability fold-back** into `stop-shipping-plugin-lock`
  (tasks 4.4 and the "Lock-based scanners" risk bullet), invariant
  wording for its SC-FR-004 offline scenario, and the AGENTS.md CI
  table's Scheduled row.
- `TestSCFR003_RunPreservesTargetDevDependencies` uses the shared
  `containsPath` helper.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `scaffold`: adds drift-sweep and acceptance-test requirements for
  the throwaway-lock scan surfaces and the SC-FR-004 offline
  scenario; amends the unarchived `stop-shipping-plugin-lock`
  SC-FR-004 scenario wording in place.

### Removed Capabilities
- None.

## Impact

- Test and doc changes plus two small `internal/doctor` fixes. No
  CI workflow, pin, threshold, or flag is changed.
- `uf doctor` now invokes `env LC_ALL=C git ...` for the
  tracked-lock check; `env` is POSIX and doctor already refuses
  Windows.

## Constitution Alignment

### I. Autonomous Collaboration

**Assessment**: PASS

No new writers or ownership changes; decisions are recorded in
the design for future agents.

### II. Composability First

**Assessment**: PASS

No new dependencies or packages; the shared-package refactor is
deliberately not taken (design D3).

### III. Observable Quality

**Assessment**: PASS

Drift between the three scan surfaces now fails a named test, and
the tracked-lock check's classification no longer depends on the
user's locale.

### IV. Testability

**Assessment**: PASS

Every behavior change carries a test: the locale and spoofing
cases, the nil seam, the scheduled-job sweeps, and the SC-FR-004
scenario.

### V. Security by Default

**Assessment**: PASS

Supply-chain pin copies gain enforcement; the toolchain staleness
window is recorded rather than left implicit.

### VI. Deterministic Execution

**Assessment**: PASS

Pin parity is enforced by executable tests instead of review prose.
