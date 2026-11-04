# scaffold: stop shipping the plugin lock

Amends `review-council-multi-model-fanout` (SC-FR-001, SC-FR-003,
SC-FR-004, SC-FR-005). The scaffold no longer ships, tracks, or
installs from `.opencode/package-lock.json`.

## ADDED Requirements

None.

## MODIFIED Requirements

### Requirement: Complete Multi-Model Review Assets

**ID: SC-FR-001** This requirement MUST be satisfied.

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

The repository and scaffold MUST track `.opencode/package.json`.
They MUST NOT track or ship `.opencode/package-lock.json`. Direct
dependencies MUST pin `@opencode-ai/plugin` 1.19.0 and Zod 4.1.8
exactly. Test dependencies MUST pin Vitest 5.0.3 and
`@vitest/coverage-v8` 5.0.3. The `toml` and `uuid` security
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
- **AND** remediation names `npm install --ignore-scripts --omit=dev --no-package-lock`

### Requirement: Plugin Doctor Validation

**ID: SC-FR-005** This requirement MUST be satisfied.

`uf doctor` MUST check Node and npm discovery and version support,
agreement between the `package.json` pins and the installed versions
(`manifest-install`), installed dependency presence, both plugin
imports, and the plugin registration state. It MUST NOT require
`.opencode/package-lock.json`. An installed version that differs
from its pin, or a declared package missing from `package.json`,
MUST fail.

#### Scenario: Installed version drifts from its pin

- **GIVEN** `.opencode/package.json` pins zod 4.1.8
- **AND** `node_modules/zod` reports 4.2.0
- **WHEN** `uf doctor` runs
- **THEN** `manifest-install` fails and names zod

## REMOVED Requirements

None.
