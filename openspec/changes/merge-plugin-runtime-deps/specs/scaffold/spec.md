# scaffold: merge plugin runtime dependencies

Amends `stop-shipping-plugin-lock` (SC-FR-003, SC-FR-005). `uf init`
merges a target's runtime dependencies into the tool-owned
`.opencode/package.json` instead of replacing them.

Requirement numbering: the SC-FR ids were allocated by
`review-council-multi-model-fanout`; this change allocates no new
ids and modifies SC-FR-003 and SC-FR-005 in place.

## ADDED Requirements

None.

## MODIFIED Requirements

### Requirement: Reproducible Plugin Dependencies

**ID: SC-FR-003** This requirement MUST be satisfied.

Previously: the requirement (as last written by
`stop-shipping-plugin-lock`) required tracking
`.opencode/package.json`; not tracking or shipping
`.opencode/package-lock.json`; exact pins for
`@opencode-ai/plugin` 1.19.0 and Zod 4.6.6; this repository's
Vitest 5.0.3 and `@vitest/coverage-v8` 5.0.3 pins; MIT license and
exact version retention verified through the provenance record;
no scaffolded `devDependencies`; keeping a target's existing
`devDependencies` on refresh; the `toml`, `zod`, and `uuid`
overrides; the `.opencode/.gitignore` lines; and tests validating
the exact pins with `uf doctor` comparing them to the install.
Every one of those clauses is carried unchanged. This change adds
the runtime-dependency merge rule and its scenarios.

The repository and scaffold MUST track `.opencode/package.json`.
They MUST NOT track or ship `.opencode/package-lock.json`. Direct
dependencies MUST pin `@opencode-ai/plugin` 1.19.0 and Zod 4.6.6
exactly. This repository's manifest MUST also pin the test
dependencies Vitest 5.0.3 and `@vitest/coverage-v8` 5.0.3. All
four pinned dependencies MUST retain their MIT license and
exact version. Tests MUST verify the license and versions
through the plugin package provenance record
(`plugin-package-provenance.json`), not through a lock file.
The scaffolded manifest MUST NOT declare `devDependencies`,
because a lockless install resolves them even under
`--omit=dev`. When `uf init` refreshes an existing
`.opencode/package.json`, it MUST keep the `devDependencies`
that file already declares. The `toml`, `zod`, and `uuid` security
overrides MUST remain.

When `uf init` refreshes an existing `.opencode/package.json`, it
MUST merge the file's runtime `dependencies` with the shipped
pins: a shipped package that is missing MUST be added, a shipped
package at another version MUST be set to the shipped pin, and a
package the shipped manifest does not declare MUST be kept at the
version the file declares. The merged `dependencies` MUST be
written in sorted key order, and a file whose merged content
equals its current content MUST be left byte-identical. All other
fields, including `overrides`, MUST be restored from the shipped
manifest. A file that cannot be parsed MUST be replaced with the
shipped manifest.

`.opencode/.gitignore` MUST ignore `node_modules`, `bun.lock`, and
`package-lock.json`. It MUST NOT ignore `package.json`. Tests MUST
validate the exact manifest pins, and `uf doctor` MUST compare them
with the installed versions.

#### Scenario: Lock written by another tool

- **GIVEN** opencode or npm writes `.opencode/package-lock.json`
- **WHEN** the user runs `git status`
- **THEN** the lock is ignored
- **AND** the next `uf init` neither overwrites nor reads it

#### Scenario: User-added runtime dependency survives uf init

- **GIVEN** `.opencode/package.json` declares a runtime dependency
  `left-pad` 1.3.0 that uf does not ship
- **WHEN** `uf init` runs, with or without `--force`
- **THEN** `left-pad` keeps version 1.3.0
- **AND** every shipped runtime pin is declared at its shipped
  version

#### Scenario: Drifted shipped pin is restored

- **GIVEN** `.opencode/package.json` pins zod 4.2.0 beside a
  user-added package
- **WHEN** `uf init` runs
- **THEN** zod is pinned at 4.6.6
- **AND** the manifest is reported as updated

#### Scenario: Unchanged merged manifest is skipped

- **GIVEN** `.opencode/package.json` already holds the merged
  shipped pins and user packages in sorted order
- **WHEN** `uf init` runs
- **THEN** the file is byte-identical and reported as skipped

#### Scenario: Unparsable manifest is restored

- **GIVEN** `.opencode/package.json` is not valid JSON
- **WHEN** `uf init` runs
- **THEN** the file is replaced with the shipped manifest

### Requirement: Plugin Doctor Validation

**ID: SC-FR-005** This requirement MUST be satisfied.

Previously: the requirement (as last written by
`stop-shipping-plugin-lock`) required Node and npm discovery and
version checks, the `manifest-install` comparison, installed
dependency presence, both plugin imports, registration state; no
lock requirement; failure on an installed version that differs
from its pin or a declared package missing from `package.json`;
a warning with `git rm --cached` remediation for a still-tracked
lock; and an npm 10 warning with npm 11 remediation when
`devDependencies` are declared. Every one of those clauses is
carried unchanged. This change scopes the comparison to uf's
shipped runtime packages and adds a scenario for user-added
packages.

`uf doctor` MUST check Node and npm discovery and version support,
agreement between the runtime `package.json` pins and the installed
versions (`manifest-install`), installed dependency presence, both plugin
imports, and the plugin registration state. It MUST NOT require
`.opencode/package-lock.json`. An installed version that differs
from its pin, or a declared package missing from `package.json`,
MUST fail. The `manifest-install` and dependency presence checks
MUST cover uf's shipped runtime packages and MUST NOT report a
runtime package the user added as drift.
`uf doctor` MUST warn when `.opencode/package-lock.json`
is still git-tracked, because the merged `.gitignore` line does not
untrack an already-committed lock, and the warning MUST point at
`git rm --cached` as the remediation.
`uf doctor` MUST warn on npm 10 when `.opencode/package.json`
declares `devDependencies`, and activation and doctor remediation
MUST name npm 11, because `uf init` keeps those `devDependencies`
and the lockless install crashes on npm 10.9.9.

#### Scenario: Installed version drifts from its pin

- **GIVEN** `.opencode/package.json` pins zod 4.6.6
- **AND** `node_modules/zod` reports 4.2.0
- **WHEN** `uf doctor` runs
- **THEN** `manifest-install` fails and names zod

#### Scenario: User-added runtime dependency is not drift

- **GIVEN** `.opencode/package.json` declares `left-pad` beside the
  shipped pins
- **AND** every shipped package is installed at its pin
- **WHEN** `uf doctor` runs
- **THEN** `manifest-install` passes and does not name `left-pad`

#### Scenario: Lock file remains tracked after the migration

- **GIVEN** `.opencode/package-lock.json` is git-tracked and now
  gitignored
- **WHEN** `uf doctor` runs
- **THEN** the tracked-lock check warns and names
  `git rm --cached .opencode/package-lock.json` as the remediation

#### Scenario: npm 10 with target devDependencies

- **GIVEN** `.opencode/package.json` declares `devDependencies`
- **AND** the detected npm is 10.9.9
- **WHEN** `uf doctor` runs
- **THEN** the npm version check warns and names npm 11 as the
  remediation

## REMOVED Requirements

None.
