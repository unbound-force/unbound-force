# scaffold: stop shipping the plugin lock

Amends `review-council-multi-model-fanout` (SC-FR-001, SC-FR-003,
SC-FR-004, SC-FR-005). The scaffold no longer ships, tracks, or
installs from `.opencode/package-lock.json`.

Requirement numbering: the SC-FR ids were allocated by
`review-council-multi-model-fanout`, which created the scaffold
capability; unarchived siblings amend them in place
(`clarification-marker-superset` amends SC-FR-003,
`init-override-stubs-gitignore-sync` carries SC-FR-015). This
change allocates no new ids and creates no reserved range; it
modifies SC-FR-001, SC-FR-003, SC-FR-004, and SC-FR-005 in
place.

## ADDED Requirements

None.

## MODIFIED Requirements

### Requirement: Complete Multi-Model Review Assets

**ID: SC-FR-001** This requirement MUST be satisfied.

Previously: the requirement also required `uf init` to provision
`.opencode/package-lock.json` as one of the shipped plugin
package inputs. This change drops the lock from the provisioned
set; every other provisioned asset is unchanged.

`uf init` MUST provision the advisor, `invoke-agent` plugin,
`review-dispatch` policy plugin, matrix, reviewer-capabilities
manifest, empty sibling template, package manifest, and
scaffold-owned review commands. It MUST NOT provision
`.opencode/package-lock.json`. Canonical command, skill, plugin,
agent, manifest, and configuration assets MUST remain byte-identical
to embedded mirrors where drift applies.

#### Scenario: Fresh full scaffold

- **GIVEN** an empty target directory
- **WHEN** full `uf init` deploys assets
- **THEN** every managed multi-model asset exists
- **AND** `.opencode/package-lock.json` is not created

### Requirement: Reproducible Plugin Dependencies

**ID: SC-FR-003** This requirement MUST be satisfied.

Previously: the requirement mandated tracking both
`.opencode/package.json` and `.opencode/package-lock.json`,
required the npm lock to retain integrity hashes with
verification failing on license, version, manifest, lock, or
integrity mismatch, required `.opencode/.gitignore` to not
ignore `package-lock.json`, and required tests and `uf doctor`
to validate manifest and lock consistency. It also carried the
maintained-interface contract (plugin ABI, Zod runtime schemas,
Vitest ESM mocking and thresholds, V8 instrumentation) and a
rollback clause restoring a tracked baseline lock (SHA256
`5ebee4af68d7a6125f5d0f52ecca7c419bbfbd9beff13503f3d5d6f36a09f965`).
This change stops tracking the lock and drops the lock-integrity
and rollback-lock clauses (accepted risk recorded in the
proposal's Impact; history in design decision D5), leaves the
maintained-interface contract to the dependency contracts
themselves, keeps the MIT license and exact-version clause
verified through the plugin package provenance record instead
of the lock, flips `.opencode/.gitignore` to ignore
`package-lock.json`, and replaces manifest-and-lock consistency
with the `manifest-install` comparison.

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

`.opencode/.gitignore` MUST ignore `node_modules`, `bun.lock`, and
`package-lock.json`. It MUST NOT ignore `package.json`. Tests MUST
validate the exact manifest pins, and `uf doctor` MUST compare them
with the installed versions.

#### Scenario: Lock written by another tool

- **GIVEN** opencode or npm writes `.opencode/package-lock.json`
- **WHEN** the user runs `git status`
- **THEN** the lock is ignored
- **AND** the next `uf init` neither overwrites nor reads it

### Requirement: Staged Plugin Activation

**ID: SC-FR-004** This requirement MUST be satisfied.

Previously: the requirement ran `npm ci` from the tracked lock
in the target `.opencode/` directory with lifecycle scripts
disabled. This change replaces `npm ci` with a lockless `npm
install --ignore-scripts --omit=dev --no-package-lock
--no-audit --no-fund` that neither reads nor writes a lock; the
staging, probe, activation, and failure semantics are unchanged.

After prerequisite validation, `uf init` MUST deploy non-plugin
assets and run `npm install --ignore-scripts --omit=dev
--no-package-lock --no-audit --no-fund` in the target `.opencode/`
directory. The install MUST NOT read or write a lock file. Lifecycle
scripts MUST remain disabled. Staging, provider-free probes, atomic
activation, `opencode.json` registration, the failure semantics, and
the Node and npm version rules are unchanged from the prior
requirement.

#### Scenario: Offline install failure

- **GIVEN** npm cannot access required dependencies
- **WHEN** installation fails
- **THEN** no plugin source remains in `.opencode/plugins/`
- **AND** remediation names
  `npm install --ignore-scripts --omit=dev --no-package-lock`
- **AND** remediation offers a pinned npm 11 runner through `npx`
  (the `PLUGIN_TEST_NPM` twin) for targets whose npm 10 crashes
  resolving devDependencies

### Requirement: Plugin Doctor Validation

**ID: SC-FR-005** This requirement MUST be satisfied.

Previously: the requirement had `uf doctor` validate manifest
and lock consistency and required
`.opencode/package-lock.json`. This change drops the lock
requirement and replaces lock consistency with the
`manifest-install` comparison between `package.json` pins and
installed versions.

`uf doctor` MUST check Node and npm discovery and version support,
agreement between the runtime `package.json` pins and the installed
versions (`manifest-install`), installed dependency presence, both plugin
imports, and the plugin registration state. It MUST NOT require
`.opencode/package-lock.json`. An installed version that differs
from its pin, or a declared package missing from `package.json`,
MUST fail. `uf doctor` MUST warn when `.opencode/package-lock.json`
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
