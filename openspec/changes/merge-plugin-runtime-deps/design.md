# Design: merge-plugin-runtime-deps

## Context

`stop-shipping-plugin-lock` made `.opencode/package.json` the only
tracked plugin package input and kept it tool-owned (SC-FR-003).
Its D2a merge re-inserts a target's existing `devDependencies`
before `overrides`, but runtime `dependencies` are still replaced
wholesale. A user who adds a package for their own `.opencode/`
plugin loses it on the next `uf init`, and the lockless install
prunes it from `node_modules` (issue #978).

## Cross-Change Reconciliation

This change amends SC-FR-003 and SC-FR-005 as last written by the
still-unarchived `stop-shipping-plugin-lock` delta (which in turn
amends `review-council-multi-model-fanout` and
`clarification-marker-superset`). It archives after
`stop-shipping-plugin-lock` so its MODIFIED text matches the
canonical text that change leaves behind. Every MUST of the
predecessor SC-FR-003 and SC-FR-005 text is carried forward; the
delta's `Previously` paragraphs list them.

## Decisions

### D1. Shipped pins win; unknown packages are kept

`preservePluginRuntimeDependencies(shipped, existing)` computes the
merged runtime map as the target's `dependencies` overlaid by the
shipped `dependencies`:

- a shipped package missing from the target is added;
- a shipped package at a different version is bumped to the
  shipped pin (uf's pins are the reproducibility and security
  contract, so they are not user-overridable);
- a package uf does not ship keeps the target's version string
  verbatim (any npm range, not normalized).

Removing a package uf used to ship (if a future release drops a
pin) leaves a stale entry the merge cannot tell apart from a user
package. That is accepted: it is rare, harmless to install, and a
release that drops a pin can document the cleanup.

### D2. Deterministic, byte-stable rendering

The merged block replaces the shipped `dependencies` body with
sorted keys in the manifest's two-space layout, encoding names and
versions through `jsonString` (no HTML escaping). When the target
adds no package of its own, the shipped bytes are returned
unchanged, so a target that only carries uf's pins (or drifted
versions of them) is restored exactly and an unchanged merged
target is byte-identical and skipped. Re-merging the output is a
no-op.

### D3. Everything else stays tool-owned

Only `dependencies` and `devDependencies` are merged.
`overrides` (the `toml`/`uuid`/`zod` security overrides),
`engines`, `name`, `private`, and `type` are restored from the
shipped manifest; a user override is dropped. Overrides are a
security control and stay uf's.

### D4. Unparsable manifests are restored, not errors

A target that is not valid JSON, or whose `dependencies` values are
not strings, passes the shipped manifest through, so `uf init`
repairs it. This follows the existing D2a contract:
`internal/scaffold` has no warnings channel, and failing `uf init`
on a broken tool-owned file would block the repair path. A shipped
manifest without a populated `dependencies` block (an anchor miss)
also passes through; the round-trip test against this repository's
manifest and the anchor-miss test cases contain that silence at
origin.

### D5. Composition and `--force`

`mergePluginManifest` runs the runtime merge, then the
`devDependencies` merge; the two splice at independent anchors
(`dependencies` body, `overrides` key). The merge runs before the
tool-owned compare, so it applies on plain re-runs and on
`--force`, matching the `devDependencies` behavior.

### D6. Doctor: no behavior change

`uf doctor` cannot know the "exact installed set" any more, and it
does not need to. `manifest-install` and `dependencies` iterate
`reviewPluginPackages` (uf's shipped runtime packages) only:

- a user-added package is never compared, so it is not drift;
- a shipped package missing from `package.json` still fails;
- a shipped package whose installed version differs from the
  declared pin still fails.

`uf doctor` does not compare declared pins against uf's embedded
pins (the doctor package does not import the scaffold assets); a
manifest whose shipped pin was hand-edited is repaired by the next
`uf init` merge, which bumps it. A table-driven doctor test pins
the three outcomes above with an extra user package present.

### D7. Provenance record: unchanged

`internal/scaffold/testdata/plugin-package-provenance.json` records
license, maintainer, and version provenance for uf's shipped
runtime pins and this repository's test-tool pins.
`TestPluginPackageInputs_ApprovedMaintenanceProvenance` reads the
embedded scaffold manifest, which this change does not alter, so
the record and its test stay valid. User-added packages are outside
uf's supply-chain claim and are not recorded.

## Risks / Trade-offs

- **User packages are unpinned by uf.** They install alongside uf's
  pins with lifecycle scripts disabled and under uf's overrides,
  but their versions and provenance are the user's responsibility.
- **Stale former pins.** See D1.
- **Peer conflicts.** A user package whose peer range excludes a
  shipped pin makes the lockless install fail; activation reports
  the failure through the existing remediation. Shipped pins are
  not relaxed to accommodate it.
