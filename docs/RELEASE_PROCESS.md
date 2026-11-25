# Release Process for unbound-force

The release process values simplicity and automation in
order to provide better predictability and low cost for
maintainers.

## Quick Start

1. Go to **Actions** → **Release** → **Run workflow**
2. Type the version tag (e.g., `v0.15.0`)
3. Click **Run workflow**
4. Done.

### Optional Skip Controls

The workflow exposes three optional boolean inputs for
debugging or recovering from partial failures:

- **Skip semver check**: Bypass semver ordering verification
  against the latest existing tag
- **Skip CI checks**: Bypass CI check verification on HEAD
- **Skip unreleased check**: Bypass verification that
  unreleased commits exist since the last tag

These should only be used when you understand why the
check is failing and have confirmed the release is safe.

## What Happens Automatically

The release workflow calls
[jflowers/ship](https://github.com/jflowers/ship)'s shared
`go-release.yml`, the same pipeline replicator, dewey and gaze
use. It runs [org-infra](https://github.com/complytime/org-infra)'s
reusable workflows for preflight validation and GoReleaser
execution (ship pins org-infra in one place), then the macOS
signing and Homebrew publishing below.

### Pre-flight Validation

The `reusable_release_preflight` workflow validates:

- **Tag format**: Must match `vMAJOR.MINOR.PATCH` (strict
  semver, no pre-release suffixes)
- **Tag uniqueness**: Must not already exist — with smart
  re-run detection (if the tag already exists at HEAD, it
  is treated as a re-run after partial failure, not an
  error)
- **Semver ordering**: New tag must be greater than the
  latest existing tag (spec-compliant semver comparison)
- **CI status**: Required checks must have passed on HEAD
  (configured via `ci_checks` input:
  `["Build and Test", "Standardized CI / Run linters",
  "OSV-Scanner / Trivy Source Scan", "Plugin Tests"]`)
- **Unreleased commits**: At least one commit must exist
  since the last release

If any validation fails, the workflow aborts with a clear
error message. Tag creation uses the GitHub API to create
annotated tags.

### Build and Release

The `reusable_release_goreleaser` workflow:

1. Checks out the code at the validated tag
2. Sets up Go (version from `go.mod`)
3. Installs **Cosign** for Sigstore keyless signing
4. Installs **Syft** for SBOM generation
5. Runs **GoReleaser** to build cross-platform binaries:
   - darwin (amd64, arm64)
   - linux (amd64, arm64)
6. **RPM packages** via nfpm (attached to GitHub release)
7. **SBOM generation** via Syft (archive + source)
8. **Cosign signing** of checksums (Sigstore keyless)
9. **Changelog** auto-generated from conventional commits
10. **Homebrew cask and formula** generated and uploaded as
    release assets (via `release.extra_files` in
    `.goreleaser.yaml`)

### macOS Signing and Homebrew

ship's `go-release.yml` checks the signing secrets before
anything is tagged: all of them, or none (an unsigned release).
With them:

1. Darwin binaries are signed with `codesign`
2. Signed binaries are notarized via Apple's notarytool
3. Unsigned release assets are replaced with signed ones
4. `checksums.txt` is rewritten for the signed archives

Then, for the Homebrew tap:

1. The cask and formula (GoReleaser release assets) get every
   sha256 from the final `checksums.txt`
2. Both are committed in a clone of the
   [homebrew-tap](https://github.com/unbound-force/homebrew-tap)
3. `brew audit --cask --strict` validates the cask through a
   throwaway tap
4. brew must load the formula (GoReleaser formulas can't pass
   `brew audit`: a redundant `version`, no `test do`)
5. The cask is installed from that clone, and the
   `cask_install_check` in `release.yml` verifies the executable
   and the `uf` alias; the runner is discarded after
6. The cask and formula are pushed to the tap (rebased and
   retried if another release pushed first)

The cask's `postflight_steps` (quarantine removal and the `uf`
symlink) come from the `custom_block` in `.goreleaser.yaml`, so
no transformation runs at release time. Any audit, load or
install failure stops the job before the tap push. For a
transient failure, inspect the failed step and re-run the
failed jobs; do not move an existing release tag.

### Fedora Packaging

After the GitHub release is published:

1. [Packit](https://packit.dev/) detects the release event
2. PRs are created in Fedora dist-git for rawhide, f44, f43
3. After a Fedora maintainer reviews and merges:
   - [Koji builds](https://koji.fedoraproject.org) are
     triggered
   - [Bodhi updates](https://bodhi.fedoraproject.org) are
     submitted for f44 and f43

End result: `dnf install unbound-force`

## Re-running After Partial Failure

If the release job fails after the tag was created (e.g.,
GoReleaser error, network timeout), you can safely re-run
the workflow with the same tag. The preflight detects that
the tag already exists at HEAD and treats it as a re-run,
skipping tag creation and re-validation.

## Cadence

Releases are opportunistic — cut when there is value to
ship. There is no fixed schedule. Project maintainers
discuss and agree on releases.

## Version Scheme

Semantic versioning: `vMAJOR.MINOR.PATCH`

- **MAJOR**: Breaking changes to CLI, schemas, or agent
  contracts
- **MINOR**: New features, commands, or hero capabilities
- **PATCH**: Bug fixes, refactoring, documentation

## Fedora Packages

### Prerequisites

Before Packit can propose to Fedora dist-git, a one-time
[Fedora package review](https://docs.fedoraproject.org/en-US/package-maintainers/Joining_the_Package_Maintainers/)
must be completed:

1. File a Bugzilla review request
2. A Fedora reviewer reviews the spec
3. Package is approved and created in dist-git
4. Packit automation activates

The `copr_build` job works immediately on PRs without
this prerequisite — it validates the RPM spec in a real
build environment.

### Target Maintenance

Packit targets (in `.packit.yaml`) must be updated with
each Fedora release cycle (~every 6 months):

- When a new Fedora version is released (e.g., F45),
  add it to all target lists
- When an old Fedora version reaches EOL, remove it
- `rawhide` is always included

Current targets (as of May 2026):
- rawhide (→ F45)
- f44 (EOL: 2027-06-02)
- f43 (EOL: 2026-12-02)

## Supply Chain Security

Each release includes:

| Artifact | Tool | Purpose |
|----------|------|---------|
| `checksums.txt` | GoReleaser | SHA256 checksums |
| `*.sigstore.json` | Cosign | Keyless signature bundle |
| `*.sbom.json` | Syft | Software Bill of Materials |
| macOS signature | codesign | Apple Developer ID |
| macOS notarization | notarytool | Apple notarization |

To verify a release:

```bash
cosign verify-blob \
  --bundle checksums.txt.sigstore.json \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate-identity-regexp 'github.com/unbound-force/unbound-force' \
  checksums.txt
```

## Rollback Procedures

If a release needs to be retracted:

### Delete the GitHub Release

```bash
gh release delete vX.Y.Z --repo unbound-force/unbound-force
git push --delete upstream vX.Y.Z
```

### Homebrew Tap

The Homebrew tap auto-updates on release. To revert:

```bash
git clone https://github.com/unbound-force/homebrew-tap.git
cd homebrew-tap
git revert HEAD  # reverts the last cask/formula update
git push
```

### Fedora Updates

If Packit has already proposed to dist-git:

- **Before merge**: Close the Packit-created PR in
  [Fedora dist-git](https://src.fedoraproject.org/rpms/unbound-force)
- **After Koji build**: Contact the Fedora package
  maintainer to unpush the update
- **After Bodhi update**: The update can be revoked via
  `bodhi updates edit --request=revoke FEDORA-YYYY-XXXXXXXXXX`

## Troubleshooting

### Pre-flight fails with "CI not passed"

Ensure the latest commit on `main` has all CI checks
green. Push a fix if needed and wait for CI to complete
before retrying. Alternatively, use the **Skip CI checks**
input if you have verified the checks manually.

### Pre-flight fails with "tag already exists"

If the tag exists at a different commit, the version has
already been released. Choose a higher version number.

If the tag exists at HEAD, the workflow detects this as a
re-run and continues automatically.

### GoReleaser fails

Check the GoReleaser output in the workflow logs. Common
causes:
- Invalid `.goreleaser.yaml` syntax
- Go build failures (should be caught by CI)
- GitHub API rate limiting (retry)

### macOS signing fails

If no signing secrets are configured, ship's signing job
is skipped and darwin binaries are unsigned but functional.
If only some are, the release fails before anything is
tagged, naming the missing ones.

### Reusable workflow fails

The release pipeline calls jflowers/ship's `go-release.yml`,
pinned by commit SHA, which pins org-infra. If it fails:

- Check [jflowers/ship](https://github.com/jflowers/ship) and the
  [org-infra repository](https://github.com/complytime/org-infra)
  for known issues
- A failure before preflight tags nothing, so the same tag
  can be dispatched again
- Individual failed jobs can be re-run via the GitHub
  Actions UI without re-running the entire pipeline

### Packit propose_downstream fails

Ensure the Fedora package review has been completed and
the package exists in dist-git. Check the
[Packit dashboard](https://dashboard.packit.dev/) for
detailed error messages.

## Maintaining the ship Dependency

`release.yml` pins jflowers/ship's `go-release.yml` by commit
SHA; ship pins org-infra. To upgrade:

1. Check [ship's tags](https://github.com/jflowers/ship/tags)
   and its README for input changes
2. Update the SHA in `release.yml` and its version comment
   (e.g., `# v0.8.0`)
3. Test with a pre-release tag if possible
