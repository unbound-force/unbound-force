<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file --
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Create Pack Content

- [x] 1.1 [P] Create `.opencode/uf/packs/ci.md` (canonical source) with frontmatter (`pack_id: ci`, `language: Any`, `version: 1.0.0`) and all CI rules: Action Pinning & Supply Chain (CI-001, CI-002, CI-003), Workflow Structure (CI-010, CI-011, CI-012), Permissions & Secrets (CI-020, CI-021, CI-022), Reusable Workflow Design (CI-030, CI-031, CI-032), Custom Rules (empty).
- [x] 1.2 [P] Create `.opencode/uf/packs/ci-custom.md` (canonical source) with frontmatter (`pack_id: ci-custom`, `language: Any`, `version: 1.0.0`), `## Custom Rules` section header, and sentinel comment `<!-- Add project-specific rules below this line -->`.
- [x] 1.3 Copy `.opencode/uf/packs/ci.md` to `internal/scaffold/assets/opencode/uf/packs/ci.md` (byte-identical embedded copy for drift detection).
- [x] 1.4 Copy `.opencode/uf/packs/ci-custom.md` to `internal/scaffold/assets/opencode/uf/packs/ci-custom.md` (byte-identical embedded copy for drift detection).

## 2. Scaffold Engine Updates

- [x] 2.1 Update `shouldDeployPack()` in `internal/scaffold/scaffold.go`: add `"ci"` and `"ci-custom"` to the always-deploy condition (line ~365).
- [x] 2.2 Update `collectDeployedPacks()` in `internal/scaffold/scaffold.go`: add `"ci.md"` and `"ci-custom.md"` to the candidates slice (line ~1299).

## 3. Test Updates

- [x] 3.1 Update `expectedAssetPaths` in `internal/scaffold/scaffold_test.go`: add `"opencode/uf/packs/ci-custom.md"` and `"opencode/uf/packs/ci.md"` in alphabetical order. Update the comment count from 11 to 13.
- [x] 3.2 [P] Add `ci.md` to the skip list in `TestValidateConventionPack_AllPacksValid` in `internal/schemas/packvalidator_test.go`, following the `severity.md` and `content.md` pattern.
- [x] 3.3 Update `TestCollectDeployedPacks_Go` in `internal/scaffold/scaffold_test.go`: add `"ci.md": true` and `"ci-custom.md": true` to the expected map; update expected count.
- [x] 3.4 Update `TestCollectDeployedPacks_Default` in `internal/scaffold/scaffold_test.go`: add `ci.md` and `ci-custom.md` to the allowed list; update expected count from 5 to 7.
- [x] 3.5 Update `TestCollectDeployedPacks_WithRoot_AllEmpty` in `internal/scaffold/scaffold_test.go`: add `"ci-custom.md"` to the stubs list and add `"ci.md"` to the `required` non-custom pack assertions.
- [x] 3.6 Update `TestCollectDeployedPacks_WithRoot_OnePopulated`: add `ci-custom.md` as an empty stub alongside `default-custom.md` and `content-custom.md`; verify it is excluded from the result. Update `TestCollectDeployedPacks_WithRoot_EmptyRootFallback`: add `ci-custom.md` to the expected pack list in the `for _, name := range` assertion.
- [x] 3.7 Add test cases to `TestIsToolOwned` in `internal/scaffold/scaffold_test.go`: `"opencode/uf/packs/ci.md"` (true) and `"opencode/uf/packs/ci-custom.md"` (false).
- [x] 3.8 Add test cases to `TestShouldDeployPack` in `internal/scaffold/scaffold_test.go`: `ci.md` and `ci-custom.md` return true regardless of language.
- [x] 3.9 Update `TestRun_DivisorSubset` in `internal/scaffold/scaffold_test.go`: add `"ci"` to the allowed pack prefixes in the DivisorOnly pack filter (alongside default, severity, content).

## 4. User Config Trim

- [x] 4.1 Replace the "YAML / GitHub Actions Workflows" section in `~/.config/opencode/coding-standards.md` (lines ~143-168) with a condensed inline summary of essential CI rules (pin to SHA, least-privilege, naming conventions) plus a reference to the `ci.md` pack for expanded enforceable rules. Keep the "Containers" section unchanged.

## 5. Documentation Gate

- [x] 5.1 Run `uf init` to update AGENTS.md and CLAUDE.md convention pack lists (deferred to `chore/uf-init-sync` branch per behavioral rules).
- [x] 5.2 Add CHANGELOG.md entry under the appropriate version header: `feat: add ci.md convention pack for CI workflow authoring`.
- [x] 5.3 File a documentation issue (unbound-force/unbound-force#271) against this repo for the new ci.md convention pack.

## 6. Verification

- [x] 6.1 Run `make test` to verify all tests pass (drift detection, pack deployment, validator exemption, tool ownership, expected asset manifest, DivisorSubset).
- [x] 6.2 Run `make lint` to verify no lint issues.
- [x] 6.3 Run `make build` to verify the binary builds with the new embedded assets.
- [x] 6.4 Verify constitution alignment: Principle II (ci.md works standalone, no mandatory dependencies), Principle IV (all changes covered by existing test patterns), Principle V (CI-002 advances supply chain integrity).
<!-- spec-review: passed -->
<!-- code-review: passed -->
