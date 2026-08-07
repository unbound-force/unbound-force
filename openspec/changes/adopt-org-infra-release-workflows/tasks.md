## Tasks

### 1. Replace inline preflight with reusable workflow

- [x] T1.1: Remove inline preflight steps (validate
  tag, check uniqueness, verify semver, verify CI,
  verify unreleased, create tag)
- [x] T1.2: Add `reusable_release_preflight` caller
  with SHA pin and version comment
- [x] T1.3: Wire `ci_checks` input with build, lint,
  and security scan check names
- [x] T1.4: Wire `skip_*` boolean inputs through
  workflow_dispatch
- [x] T1.5: Add `ci_checks` source comment documenting
  which workflow files define the check names

### 2. Replace inline GoReleaser with reusable workflow

- [x] T2.1: Remove inline GoReleaser steps (checkout,
  setup-go, cosign, syft, goreleaser-action, artifact
  upload)
- [x] T2.2: Add `reusable_release_goreleaser` caller
  with SHA pin and version comment
- [x] T2.3: Add `release.extra_files` to
  `.goreleaser.yaml` for atomic Homebrew artifact upload
- [x] T2.4: Add `name_template` to formula entry to
  avoid filename collision with cask

### 3. Extract and expand signing secrets check

- [x] T3.1: Extract `check-signing-secrets` into
  separate job with `permissions: {}`
- [x] T3.2: Expand to verify all 6 required secrets
- [x] T3.3: Add `::notice::` for missing secrets
- [x] T3.4: Update `sign-macos` `needs` and `if`
  condition

### 4. Restore operational safeguards

- [x] T4.1: Add workflow-level concurrency group
- [x] T4.2: Add `name:` field to `sign-macos` job

### 5. Update documentation

- [x] T5.1: Update `docs/RELEASE_PROCESS.md` preflight
  section
- [x] T5.2: Update build and release section
- [x] T5.3: Add skip controls section
- [x] T5.4: Add re-running after partial failure section
- [x] T5.5: Add reusable workflow troubleshooting
  section
- [x] T5.6: Add org-infra upgrade path section
- [x] T5.7: Update troubleshooting entries
- [x] T5.8: Add CHANGELOG entry

### 6. Create retroactive spec (this file)

- [x] T6.1: Create `openspec/changes/` directory
- [x] T6.2: Write proposal with rationale and
  acceptance criteria
- [x] T6.3: Write task breakdown
