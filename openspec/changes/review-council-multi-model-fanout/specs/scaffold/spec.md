## ADDED Requirements

### Requirement: Complete Multi-Model Review Assets

**ID: SC-FR-001** This requirement MUST be satisfied.

`uf init` MUST provision the advisor, `invoke-agent` plugin,
`review-dispatch` policy plugin, matrix, reviewer-capabilities
manifest, empty sibling template, package manifest, package lock, and
scaffold-owned review commands. The reviewer manifest MUST validate
against `schemas/reviewer-capabilities/v1.0.0.schema.json`. Canonical
command, skill, plugin, agent, manifest, and configuration assets MUST
remain byte-identical to embedded mirrors where drift applies.

The exact asset manifest, path mapping, directory expectations, and
file-count assertions MUST include every new asset. Existing agent
mirrors MUST remain drift-covered. This change MUST NOT add reviewer
metadata to agent frontmatter. Speckit test review MUST remain
live-only unless Speckit command ownership changes separately.

#### Scenario: Fresh full scaffold

- **GIVEN** an empty target directory
- **WHEN** full `uf init` deploys assets
- **THEN** every managed multi-model asset exists
- **AND** the sibling file has no active organization entries

#### Scenario: Canonical asset drift

- **GIVEN** a managed command differs from its mirror
- **WHEN** drift tests run
- **THEN** the tests fail and name the mismatch

### Requirement: UF Asset Path Mapping

**ID: SC-FR-002** This requirement MUST be satisfied.

The scaffold MUST map embedded `uf/` paths to runtime `.uf/` paths.
`uf/` MUST be known. Unknown prefixes MUST continue to fail the
asset-prefix contract test.

#### Scenario: Review matrix mapping

- **GIVEN** embedded asset `uf/review-matrix.yaml`
- **WHEN** `mapAssetPath` maps it
- **THEN** the result is `.uf/review-matrix.yaml`

### Requirement: Reproducible Plugin Dependencies

**ID: SC-FR-003** This requirement MUST be satisfied.

The repository and scaffold MUST track `.opencode/package.json` and
`.opencode/package-lock.json`. Direct dependencies MUST pin
`@opencode-ai/plugin` 1.4.10 and Zod 4.1.8.
Test dependencies MUST pin Vitest 5.0.3 and
`@vitest/coverage-v8` 5.0.3.

The plugin package MUST provide the supported OpenCode plugin, tool,
and
client ABI. Zod MUST provide the SDK-compatible runtime schemas used by
the tool.
Vitest MUST provide TypeScript ESM mocking and threshold enforcement.
The V8 adapter MUST provide branch and statement instrumentation.
Hand-written session protocols and a custom Node test harness MUST NOT
replace these maintained interfaces in this change.

All four dependencies MUST retain their MIT license and exact version.
The npm lock MUST retain integrity hashes. Implementation verification
MUST fail on license, version, manifest, lock, or integrity mismatch.
Rollback MUST remove both plugin activations, the reviewer manifest,
scripts, and direct dependencies. It MUST restore host-model Task
dispatch, `.opencode/.gitignore`, and the tracked baseline lock whose
SHA256 is
`b628c764236019785a620deadc06373949236ce84fb38bc68a39f619411874e3`.

`.opencode/.gitignore` MUST ignore `node_modules` and `bun.lock`. It
MUST NOT ignore `package.json` or `package-lock.json`. Tests and
`uf doctor` MUST validate manifest and lock consistency.

#### Scenario: Fresh checkout dependencies

- **GIVEN** a fresh checkout has no installed modules
- **WHEN** npm validates the manifest and lock
- **THEN** the locked runtime install has a reproducible input
- **AND** no generated manifest is required

### Requirement: Staged Plugin Activation

**ID: SC-FR-004** This requirement MUST be satisfied.

`uf init` MUST resolve `node` and `npm` and parse their semantic
versions before plugin installation. Supported versions MUST be Node
20 through 24 and npm 10 through 11. A missing executable, malformed
version, or unsupported version MUST prevent plugin activation and
report an actionable Node 22 and npm 10 remediation.

Version output MAY end in one LF or CRLF. After removing it and
trimming ASCII space or tab only at both edges, output MUST contain one
non-empty line. Node MUST match anchored ASCII regex
`^v?(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`. npm MUST
match the same regex without `v`. Components above unsigned 32-bit
range MUST be rejected before conversion. Leading zeroes and Unicode
digits, signs, embedded whitespace, prerelease or build data, extra
lines, and
npm `v` prefixes MUST be rejected. Tests MUST cover Node 19, 20, 24,
and 25; npm 9, 10, 11, and 12; an explicit rejected npm `v10.11.0`
fixture; CRLF; huge values; malformed output; and command failure.

After prerequisite validation, `uf init` MUST deploy non-plugin assets
and run `npm ci --ignore-scripts --omit=dev` in the target `.opencode/`
directory. Repository tests MUST use a full locked development install.
The scaffold MUST stage both plugin sources outside OpenCode's
auto-discovery path and run provider-free import and tool-definition
probes.

Only after install and both probes MAY it atomically move source to
`.opencode/plugins/invoke-agent/index.ts` and
`.opencode/plugins/review-dispatch/index.ts`. Activation MUST register
both plugins in `opencode.json` (the project `plugin` array) rather than
rely on plugin auto-discovery.

On install or probe failure, plugin source MUST be absent from
`.opencode/plugins/`. Other deployed assets MUST remain. The review
plugins subtool result MUST report `failed` with activation `inactive`
and a safe remediation command. Independent subtools MUST continue.
The top-level result MUST set `Status` to `partial`, increment
`FailedSubTools` by one, and retain successful file counts. The CLI
MUST print one failed and inactive review-plugins result plus
remediation. It MUST return nil and exit zero. Fatal core scaffold I/O
MUST still return an error and exit non-zero. Rerunning `uf init`
MUST retry idempotently.

The command boundary MUST receive the target working directory and be
injectable for tests. Lifecycle scripts MUST remain disabled.
Prerequisite and activation tests MUST be network-free.

#### Scenario: Successful staged activation

- **GIVEN** dependencies install and the load probe passes
- **WHEN** activation completes
- **THEN** both plugin sources are atomically placed at runtime paths
- **AND** both plugins are registered in `opencode.json`

#### Scenario: Offline install failure

- **GIVEN** npm cannot access required cached dependencies
- **WHEN** installation fails
- **THEN** no plugin source remains in `.opencode/plugins/`
- **AND** remediation names `npm ci --ignore-scripts --omit=dev`

#### Scenario: Unsupported Node version

- **GIVEN** Node 18 and npm 10 are available
- **WHEN** prerequisite validation runs
- **THEN** dependency installation does not start
- **AND** remediation names a supported Node and npm version

#### Scenario: Plugin activation partially fails

- **GIVEN** core assets deploy but the plugin probe fails
- **WHEN** `uf init` completes
- **THEN** the CLI exits zero with top-level status `partial`
- **AND** `FailedSubTools` increases while file counts remain
- **AND** review-plugins is failed, inactive, and not registered
- **AND** output gives remediation and rerun retries activation

### Requirement: Plugin Doctor Validation

**ID: SC-FR-005** This requirement MUST be satisfied.

`uf doctor` MUST check Node and npm discovery and version support,
manifest and lock agreement, installed dependency presence, both plugin
imports, and the plugin registration state. Runtime checks MUST be
network-free. It MUST fail when an active plugin cannot load. It MUST
distinguish an inactive repairable scaffold from a broken active plugin
and provide remediation.

#### Scenario: Registered plugin cannot load

- **GIVEN** either review plugin is registered in `opencode.json`
- **AND** its dependency import fails
- **WHEN** `uf doctor` runs
- **THEN** that plugin check fails
- **AND** output identifies the repair command

#### Scenario: npm is unavailable

- **GIVEN** the scaffold includes plugin dependency files
- **AND** npm is not available on PATH
- **WHEN** `uf doctor` runs
- **THEN** the prerequisite check fails
- **AND** output provides Node 22 and npm 10 remediation

### Requirement: Persistent Coverage Gates

**ID: SC-FR-006** This requirement MUST be satisfied.

The repository MUST track `coverage-gate.json`. Its closed v1
contract MUST map each changed or new Go helper scope to a package and
optional file or function matcher and a minimum. New Go helper
scopes MUST be at least 80 percent statements. Pure path, mapping,
configuration, artifact, and schema helper scopes MUST be at least
90 percent statements.

A `coverage-gate` Make target MUST run race-enabled uncached Go tests,
preserve the existing global 80 percent and backlog 90 percent checks,
and enforce every manifest entry. Missing packages, empty values,
non-numeric values, and results below threshold MUST fail. Existing
ratchet commands and thresholds MUST NOT be weakened or replaced.

The gate MUST derive changed production Go functions from the CI PR
base SHA or the local `main...HEAD` merge base. It MUST exclude test,
generated, and vendored files. Every changed function MUST match
exactly one manifest scope. An omitted or multiply matched function
MUST fail the gate.

`make check` and Local CI MUST invoke the same target. Final
verification MUST execute integrated `make check` and the Local CI
path. The manifest, parser, completeness check, package selection, and
failure cases MUST have isolated tests.

#### Scenario: Scoped helper coverage regresses

- **GIVEN** a manifest scope requires 90 percent statements
- **AND** its measured result is 89.9 percent
- **WHEN** local or CI coverage gates run
- **THEN** the gate fails and names the scope and threshold

#### Scenario: Coverage value is absent

- **GIVEN** a manifest scope produces no numeric coverage value
- **WHEN** the gate evaluates that scope
- **THEN** the gate fails rather than treating absence as zero or pass

#### Scenario: Changed helper is omitted

- **GIVEN** a changed production function matches no manifest scope
- **WHEN** local or CI coverage gates run
- **THEN** the gate fails and names the omitted function

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
