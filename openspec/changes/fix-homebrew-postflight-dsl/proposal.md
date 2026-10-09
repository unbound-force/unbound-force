## Why

Homebrew warns that the generated `unbound-force` Cask uses its
deprecated `postflight` DSL. GoReleaser's `homebrew_casks.hooks`
configuration only accepts arbitrary Ruby hook bodies and currently
renders the deprecated wrapper; it does not expose a
`postflight_steps` setting. The release pipeline must translate this
known generated hook into Homebrew's supported declarative DSL before
the Cask is published to the tap.

## What Changes

- Add a fail-closed transformation to the macOS release publishing
  path for the generated `unbound-force` Cask.
- Replace the specific generated `postflight` hook with
  `postflight_steps` that retain quarantine removal and the `uf`
  symlink alias.
- Verify the transformed Cask contains no deprecated `postflight`
  stanza and retains the expected supported install steps before it is
  copied to `unbound-force/homebrew-tap`.
- Add regression coverage for the transformation and its failure
  conditions without changing any release, security, or quality gate.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `homebrew-cask-publishing`: Publish a Homebrew-compatible Cask with
  declarative post-install steps while preserving existing install
  behavior.

### Removed Capabilities

- None.

## Impact

- Affects `.github/workflows/release.yml` and the release-pipeline
  test coverage for generated Homebrew Casks.
- Does not change the GoReleaser version, its configuration schema, or
  any pinned dependency.
- Removes installation-time Homebrew deprecation warnings while
  retaining signed-artifact checksum patching and tap publication.
- This CI-only change has no website documentation impact.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

The release pipeline continues to publish a self-contained Cask. This
change introduces no runtime dependency or inter-hero communication.

### II. Composability First

**Assessment**: PASS

The Cask remains an independent installation path for the
`unbound-force` CLI. The change adds no dependency on another hero or
repository.

### III. Observable Quality

**Assessment**: PASS

The pipeline will verify observable generated-Cask content before
publication. Release artifacts and existing provenance remain
unchanged.

### IV. Testability

**Assessment**: PASS

The transformation will be isolated from network publishing and tested
against representative generated Cask content, including a missing or
ambiguous source hook failure path.

### V. Security by Default

**Assessment**: PASS

The change preserves the existing signed-checksum flow and immutable
action pins. It uses a narrowly scoped, fail-closed transformation of
the known generated artifact and does not weaken release gates.
