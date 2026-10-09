## Context

The release workflow generates the `unbound-force` Homebrew Cask with
GoReleaser, patches signed macOS checksums, and then copies the Cask to
`unbound-force/homebrew-tap`. The configured GoReleaser hook emits
Homebrew's deprecated `postflight` Ruby block. GoReleaser does not
offer a configuration field that emits `postflight_steps`, and its
arbitrary-Ruby hook body cannot safely be renamed into the restricted
declarative DSL.

The proposal establishes that the repository must own a narrowly
scoped translation at the generated-artifact boundary. It aligns with
Observable Quality and Testability by making that transformation
deterministic, independently testable, and verified before publication.

## Goals / Non-Goals

### Goals
- Convert only the known generated `unbound-force` post-install hook
  into valid `postflight_steps` syntax.
- Preserve quarantine removal and creation of the `uf` symlink alias.
- Fail before tap publication if GoReleaser output no longer matches
  the expected input structure.
- Provide isolated regression tests without network access or Homebrew.
- Retain all existing checksum, signing, and release quality gates.

### Non-Goals
- Change the pinned reusable workflow, GoReleaser configuration, or
  GoReleaser version-selection policy.
- Modify the Replicator Cask or another repository's release pipeline.
- Generalize arbitrary GoReleaser Ruby hooks into Homebrew install-step
  syntax.
- Replace Homebrew audit or macOS installation smoke testing.

## Decisions

### Use a version-controlled Go transformer

Add a small Go command backed by an `internal` package to transform a
specified generated Cask file. The release workflow invokes the
command after checksum patching and before the Cask is copied to the
tap. Keeping the parsing and rewrite logic in Go allows standard
library unit tests to validate exact input and output without requiring
Homebrew, Ruby, network access, or a release token.

The command accepts explicit input and output paths. It reads the
input, identifies exactly one literal legacy block defined below, writes
the literal replacement only after validation, and returns a contextual
error for absent, duplicate, or mismatched blocks. It MUST NOT run shell
content from the Cask or broaden the match beyond the expected generated
hook.

### Emit declarative replacement steps

The canonical input and replacement are intentionally literal so fixture
tests can prove exact, fail-closed behavior:

```ruby
postflight do
  if OS.mac?
    system_command "/usr/bin/xattr", args: ["-dr", "com.apple.quarantine", "#{staged_path}/unbound-force"]
    # Create uf symlink alias for daily-use convenience (FR-002/FR-009).
    system_command "/bin/ln", args: ["-sf", "#{staged_path}/unbound-force", "#{HOMEBREW_PREFIX}/bin/uf"]
  end
end
```

```ruby
postflight_steps do
  on_macos do
    run "/usr/bin/xattr",
      args: ["-dr", "com.apple.quarantine", "{{staged_path}}/unbound-force"]
    symlink "unbound-force", "{{HOMEBREW_PREFIX}}/bin/uf", overwrite: true
  end
end
```

`run` performs quarantine removal, while `symlink` with `overwrite:
true` preserves the legacy `ln -sf` behavior. The `{{...}}` values are
Homebrew deferred tokens; they replace the arbitrary Ruby interpolation
used by the legacy hook. `on_macos` preserves the original OS guard.

### Validate semantically before tap publication

The existing `sign-macos` job already runs on `macos-latest`, making it
the authoritative validation boundary. After signed-checksum patching
and transformation, it checks for one supported stanza, no legacy
`postflight do` stanza, and both required command paths. It then runs
Homebrew static validation and installs the transformed Cask from an
isolated staged copy of the cloned tap.

The smoke test asserts the installed executable and `uf` alias exist,
then uninstalls and cleans up the alias. It executes before the Cask is
copied to the tap's publishable path or `git push` runs. Any validation
or cleanup failure stops publication. Existing signed-checksum patching
and release gates remain unchanged.

### Test observable transformations

Unit tests use fixture Cask text to assert successful replacement,
unchanged surrounding Cask content, missing-block rejection,
duplicate-block rejection, and mismatched-block rejection. A workflow
regression test verifies the required ordering: checksum patching,
transformation, Homebrew validation, staged install smoke test, tap
copy, and `git push`. It also verifies transformer or validation
failure leaves publication unreachable.

The transformer coverage target is 100% of its five defined behavioral
contracts, rather than a new source-line percentage threshold: exact
single-hook replacement, absent-hook rejection, duplicate-hook
rejection, near-match rejection, and surrounding-content preservation.
The success fixture compares exact expected output. Each rejection
fixture asserts a contextual error and no output file creation or
mutation. The existing CI coverage gates remain unchanged; the
verification task reports package coverage as supplemental evidence.

## Risks / Trade-offs

- GoReleaser may change its Cask template. Exact matching deliberately
  fails the release rather than publishing an unverified Cask; the
  generated template must then be reviewed and the transformer updated.
- `postflight_steps` supports only declarative operations. This design
  covers the two known commands and intentionally does not attempt to
  translate arbitrary future Ruby hooks.
- A new Go command increases repository code surface, but it confines
  transformation behavior to testable code and avoids fragile inline
  workflow scripting.
- The mandatory staged install uses release artifacts and Homebrew
  dependencies, so it is network-dependent and may lengthen releases.
  It is limited to the existing isolated macOS release runner and
  provides validation that unit tests cannot.
- The release workflow changes, but the design does not alter action
  pins, permissions, checksum verification, signing, or CI flags.
