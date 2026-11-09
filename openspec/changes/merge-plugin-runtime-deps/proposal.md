# Proposal: merge-plugin-runtime-deps

## Why

`uf init` treats `.opencode/package.json` as tool-owned. On every
run it restores the shipped manifest (runtime pins plus
overrides), which drops any runtime dependency a user added for
their own `.opencode/` plugins. The lockless install that follows
(`npm install --omit=dev --no-package-lock`) then prunes those
packages from `node_modules`, so a user plugin breaks on the next
`uf init` (issue #978).

`stop-shipping-plugin-lock` (PR #957) removed the lock from the
fight but left the manifest ownership question open. It already
merges `devDependencies` (`preservePluginDevDependencies`); runtime
dependencies have no equivalent.

## What Changes

- **`uf init` merges runtime dependencies.** When it refreshes an
  existing `.opencode/package.json`, uf's shipped runtime pins are
  added when missing and bumped when drifted, and packages uf does
  not ship keep the target's declared version. The merged block is
  written with sorted keys, so the output is deterministic and an
  unchanged target stays byte-identical (skipped).
- **Everything else stays tool-owned.** `overrides`, `engines`,
  `name`, and other top-level fields are restored from the shipped
  manifest. `devDependencies` keep their existing merge.
- **An unparsable manifest is restored.** A target manifest that is
  not valid JSON (or whose dependency values are not strings) is
  replaced with the shipped manifest, the same silent-passthrough
  contract the `devDependencies` merge already follows.
- **`uf doctor` is unchanged in behavior.** `manifest-install` and
  `dependencies` already iterate only uf's shipped packages, so a
  user-added package is never reported as drift; a missing or
  drifted shipped pin still fails. A regression test pins this.
- **The provenance record is unchanged.** It verifies uf's shipped
  pins; user-added packages are outside uf's supply-chain claim.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `scaffold`: SC-FR-003 gains the runtime-dependency merge rule for
  `uf init` refreshes; SC-FR-005 clarifies that `uf doctor`
  compares only uf's shipped runtime pins.

### Removed Capabilities
- None.

## Impact

- `internal/scaffold/plugin_activation.go` (new
  `mergePluginManifest`, `preservePluginRuntimeDependencies`),
  `internal/scaffold/scaffold.go` (call site).
- Tests in `internal/scaffold/` and `internal/doctor/`.
- Docs: `docs/architecture.md` (file ownership), `AGENTS.md`,
  `CHANGELOG.md`.
- User packages added to `.opencode/package.json` are installed by
  the activation install alongside uf's pins. They are unpinned by
  uf and outside the provenance record; their versions are the
  user's responsibility.
- No gate changes: shipped pins, overrides, coverage thresholds and
  CI flags are untouched.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

The change replaces an ownership fight over one file with an
explicit split: uf owns its pins and overrides, the user owns the
packages they add.

### II. Composability First

**Assessment**: PASS

Projects can extend `.opencode/` with their own plugin
dependencies without forking or patching the scaffold.

### III. Observable Quality

**Assessment**: PASS

`uf doctor` still reports missing and drifted shipped pins, and
the merge result is reported through the existing
created/updated/skipped buckets.

### IV. Testability

**Assessment**: PASS

The merge is a pure function over two byte slices with
table-driven unit tests, and a scenario-named acceptance test
drives it through `Run()`.

### V. Security by Default

**Assessment**: PASS

Security overrides stay tool-owned and are always restored, and
shipped pins always win over a target's version. Lifecycle scripts
stay disabled on the install that now also installs user packages.

### VI. Deterministic Execution

**Assessment**: PASS

The merged block is rendered in sorted key order; re-running the
merge over its own output is a no-op (pinned by test).
