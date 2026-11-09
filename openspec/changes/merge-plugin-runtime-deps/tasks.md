<!--
  [P] marks tasks eligible for parallel execution.
-->

## 1. Scaffold

- [x] 1.1 Add `preservePluginRuntimeDependencies` to
  `internal/scaffold/plugin_activation.go`: shipped pins added or
  bumped, unknown target packages kept, sorted rendering, shipped
  bytes returned when the target adds nothing, passthrough on an
  unparsable target or anchor miss (design D1-D4)
- [x] 1.2 Compose it with `preservePluginDevDependencies` in
  `mergePluginManifest` and call that from the `.opencode/package.json`
  restore path in `internal/scaffold/scaffold.go` (design D5)

## 2. Doctor and provenance

- [x] 2.1 Assess `internal/doctor/` `manifest-install` and
  `dependencies`: no behavior change, document the scope in the
  `checkManifestInstallConsistency` comment (design D6)
- [x] 2.2 Assess `plugin-package-provenance.json`: unchanged
  (design D7)

## 3. Tests

- [x] 3.1 [P] Table-driven unit tests for
  `preservePluginRuntimeDependencies` (passthrough cases, add,
  bump, keep, literal ranges, sorted keys, idempotence, anchor
  miss) and `mergePluginManifest` (runtime plus dev merge,
  overrides tool-owned, repository manifest round-trip) in
  `internal/scaffold/plugin_manifest_merge_test.go`
- [x] 3.2 [P] Scenario-named acceptance test
  `TestSCFR003_UserRuntimeDependencySurvivesInit` driving `Run()`
  through the SC-FR-003 scenarios
- [x] 3.3 [P] Doctor test
  `TestCheckManifestInstallConsistency_UserRuntimeDependencyIsNotDrift`
  for the SC-FR-005 user-added scenario plus drifted and missing
  shipped pins

## 4. Documentation and verification

- [x] 4.1 Document manifest merge ownership in
  `docs/architecture.md` (File Ownership Model), `AGENTS.md`, and
  `CHANGELOG.md`
- [x] 4.2 `make check` passes
- [x] 4.3 File the documentation-gate issue (#1012) for the user-facing
  `uf init` manifest merge before merge
