<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. GoReleaser Supply Chain Security

- [x] 1.1 Add `sboms` section to `.goreleaser.yaml` with
  two entries: `artifacts: archive` and
  `id: source, artifacts: source` (FR-008)
- [x] 1.2 Add `signs` section to `.goreleaser.yaml` with
  Cosign keyless signing of checksum artifacts using
  `sign-blob --yes --bundle` args (FR-009)
- [x] 1.3 Verify GoReleaser config is valid by running
  `goreleaser check` (or equivalent local validation)

## 2. Release Workflow Redesign

- [x] 2.1 Change `.github/workflows/release.yml` trigger
  from `on: push: tags: ['v*']` to
  `on: workflow_dispatch` with required `tag` input of
  type string with description "Release tag (e.g.,
  v0.15.0)" (FR-001, FR-M01)
- [x] 2.2 Add `preflight` job (ubuntu-latest) with steps:
  (a) validate tag format via regex
  `^v[0-9]+\.[0-9]+\.[0-9]+$` (FR-002),
  (b) verify tag does not exist via
  `git ls-remote --tags` (FR-003),
  (c) verify semver ordering — new tag > latest
  existing tag (FR-004),
  (d) verify CI status on HEAD via `gh api` check-runs
  endpoint for four required checks (FR-005),
  (e) verify unreleased commits via `git log` between
  latest tag and HEAD (FR-006),
  (f) create annotated tag and push using default
  GITHUB_TOKEN (FR-007).
  Output `has_signing_secrets` for sign-macos
  conditional.
- [x] 2.3 Update `release` job: add `needs: preflight`,
  install Cosign via `sigstore/cosign-installer`,
  install Syft via `anchore/sbom-action/download-syft`,
  set `GORELEASER_CURRENT_TAG` from workflow input,
  add `id-token: write` permission (FR-009, FR-M03).
  All GitHub Actions MUST be pinned to commit SHAs per
  project CI conventions (not mutable tags).
- [x] 2.4 Update `sign-macos` job: change `needs` from
  `release` (unchanged job name, just dependency
  rewiring). Verify `has_signing_secrets` output is
  passed from preflight through release job.
- [x] 2.5 Remove the `on: push: tags` trigger block
  entirely (FR-R01). Verify no other workflows depend
  on the tag push event.

## 3. Fedora Packaging

- [x] 3.1 [P] Create `.packit.yaml` with upstream metadata
  (`upstream_project_url`, `upstream_tag_template: v{version}`,
  `upstream_package_name`, `downstream_package_name`,
  `specfile_path`), `files_to_sync` list, and four jobs:
  `copr_build` (PR trigger, targets: fedora-rawhide-x86_64,
  fedora-44-x86_64, fedora-43-x86_64),
  `propose_downstream` (release trigger, branches:
  rawhide, f44, f43), `koji_build` (commit trigger,
  same branches), `bodhi_update` (commit trigger,
  f44, f43) (FR-010)
- [x] 3.2 [P] Create `unbound-force.spec` following Fedora
  Packaging Guidelines: Name unbound-force, License
  Apache-2.0, Source from GitHub archive tarball,
  BuildRequires golang and go-rpm-macros, `%build`
  using `go build -buildmode=pie` with ldflags
  (ldflags variable paths MUST match the existing
  `.goreleaser.yaml` configuration), `%install` for
  `/usr/bin/unbound-force` and `/usr/bin/uf` symlink,
  `%check` running `go test -race -count=1 ./...`,
  `%files` with binary, symlink, LICENSE, README.md
  (FR-011)
- [x] 3.3 [P] Create `.fmf/version` file containing `1`
  (FR-013)

## 4. Release Documentation

- [x] 4.1 [P] Create `docs/RELEASE_PROCESS.md` documenting:
  how to trigger a release (workflow_dispatch steps
  with screenshot-friendly description), what happens
  automatically (full pipeline: preflight, GoReleaser,
  SBOM, Cosign, macOS signing, Homebrew tap, Packit),
  release cadence (opportunistic), version scheme
  (semver), Fedora packaging overview (Packit flow,
  dist-git, Koji, Bodhi), Fedora package review
  prerequisite, Packit target maintenance schedule
  (~every 6 months on new Fedora release), rollback
  procedures (how to delete a failed release, remove
  the tag, handle Homebrew tap and Fedora update
  state) (FR-012)
- [x] 4.2 [P] Update `AGENTS.md` Project Structure section
  to add `docs/` directory, `.packit.yaml`,
  `unbound-force.spec`, and `.fmf/` if not present.
  Update CI Workflow Structure table to note release
  workflow trigger change.

## 5. Draft Release Cleanup

- [x] 5.1 Delete orphaned draft release v0.13.0 via
  `gh release delete` (FR-014)
- [x] 5.2 Delete orphaned draft release v0.4.1 via
  `gh release delete` (FR-014)

## 6. Verification

- [x] 6.1 Run `make check` to verify no regressions from
  AGENTS.md or other file changes
- [x] 6.2 Validate `.goreleaser.yaml` syntax (goreleaser
  check or YAML lint)
- [x] 6.3 Validate `.packit.yaml` YAML syntax
- [x] 6.4 Validate `unbound-force.spec` syntax (rpmlint
  if available, otherwise manual review)
- [x] 6.5 Verify constitution alignment: Autonomous
  Collaboration (release artifacts are self-describing),
  Composability First (Packit is additive, core release
  works without it), Observable Quality (SBOM and
  Cosign add machine-parseable provenance), Testability
  (copr_build validates spec on PRs, GoReleaser
  snapshot for dry-run)
- [x] 6.6 Update CHANGELOG.md with release-automation
  change entry
- [x] 6.7 File `unbound-force/website` issue for
  user-facing changes: workflow_dispatch release
  trigger, Fedora packaging availability, supply
  chain security artifacts (SBOM, Cosign)

<!-- spec-review: passed -->
<!-- code-review: passed -->
