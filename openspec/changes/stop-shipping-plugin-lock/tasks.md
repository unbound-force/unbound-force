<!--
  [P] marks tasks eligible for parallel execution.
-->

## 1. Scaffold and activation

- [x] 1.1 Remove the embedded
  `internal/scaffold/assets/opencode/package-lock.json` asset, and
  drop `opencode/package-lock.json` from both tool-owned path
  lists in `internal/scaffold/scaffold.go`
- [x] 1.2 Add `package-lock.json` to the canonical
  `opencodeGitignore` lines and to this repository's
  `.opencode/.gitignore`
- [x] 1.3 Switch `installAndActivateReviewPlugins`
  (`internal/scaffold/plugin_activation.go`) to
  `npm install --ignore-scripts --omit=dev --no-package-lock
  --no-audit --no-fund`, and update the remediation text
- [x] 1.4 Stop tracking `.opencode/package-lock.json` in this
  repository, and make `make plugin-test` install with
  `npx -y npm@11.6.2 install --ignore-scripts --no-package-lock
  --no-audit --no-fund` (npm 11 because node 22's bundled
  npm 10.9.9 crashes in Arborist on the lockless vitest 5
  peer graph)
- [x] 1.5 Ship the scaffold `.opencode/package.json` without
  `devDependencies`, keep a target's existing `devDependencies`
  when `uf init` refreshes the manifest, and limit the doctor
  package checks to the runtime packages, so activation installs
  cleanly on npm 10.9.9 (design D2a)

## 2. Doctor

- [x] 2.1 Replace `checkManifestLockConsistency` with
  `checkManifestInstallConsistency` (`manifest-install`) in
  `internal/doctor/checks.go`, which compares each declared pin
  with the installed `node_modules/<name>/package.json` version,
  and update the remediation text

## 3. Tests

- [x] 3.1 Update `internal/scaffold/scaffold_test.go`:
  - exact-version manifest gate without the lock assertions;
  - `.gitignore` gate requires `package-lock.json` to be ignored;
  - mirror and fresh-target gates assert no lock is embedded
    or deployed;
  - asset inventory, surface pairs and tool-owned tables updated;
  - gitignore merge expectations updated.
- [x] 3.2 [P] Update `internal/scaffold/plugin_activation_test.go`
  and the stealth gitignore test for the new install command
- [x] 3.3 [P] Update `internal/doctor/review_plugins_test.go`:
  installed-version fixtures, plus mismatch, not-installed,
  missing-manifest, unreadable-install and undeclared cases
- [x] 3.4 Shift the CLI file-count window in
  `cmd/unbound-force/main_test.go` from [82-86] to [81-85]
  (one fewer deployed asset)

## 4. Verification

- [x] 4.1 `go test -race -count=1 ./...` passes, including the
  base-count path with no Node or npm on `PATH`
- [x] 4.2 `make plugin-test` passes and leaves no
  `.opencode/package-lock.json` behind
- [x] 4.3 File the documentation-gate issues: one in this repo,
  labeled `docs`, and one in `unbound-force/website`, both
  covering the new install command, the ignored lock, the
  compensating vulnerability-scan posture, and the rollout note
  for existing hero repositories
- [x] 4.4 Add the compensating `osv_plugin_transitive` job to
  `.github/workflows/ci_security.yml` (throwaway-lock OSV scan)
- [x] 4.5 Reconcile the sibling deltas for the lockless install:
  `ts-plugin-coverage-gate`, `init-override-stubs-gitignore-sync`,
  and `fix-triage-issue-dispatch-docs` (FR-007 lockstep set)
- [x] 4.6 Remove `TestPluginPackageInputs_RollbackPrerequisites`
  and `testdata/plugin-package-rollback.json` per design D5
- [x] 4.7 Bump the zod pin to 4.6.6 (plus a matching override) and
  the toml override to 5.0.1 per design D6, sweeping every literal
  site: both manifests, the provenance record, the spec delta,
  the proposal, and the tests
- [x] 4.8 Add the `uf doctor` tracked-lock warning per SC-FR-005,
  with unit tests for the warn, untracked, and unavailable paths
