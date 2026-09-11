## Why

On macOS with `UserModeNetworking: true`, Podman container
IPs (e.g., `10.88.0.x`) are not routable from the host.
The only way to reach services inside the container is via
explicit port publishing (`-p` flags). `buildRunArgs()` and
`buildPersistentRunArgs()` hardcoded `-p 4096:4096` for
`DefaultServerPort` and handled demo ports from config/flags,
but never parsed `.devcontainer/devcontainer.json` for
`forwardPorts`. Users had to manually add `-p` flags or
use `--demo-ports` for every port they needed.

## What Changes

Add a `parseDevcontainerPorts` helper that reads
`forwardPorts` from `.devcontainer/devcontainer.json` and
returns host:container port mappings. Call it from both
`buildPersistentRunArgs` and `buildRunArgs` to publish
ports not already covered by `DefaultServerPort` or demo
ports. Handle JSONC comments and trailing commas per the
devcontainer spec.

## Capabilities

### New Capabilities

- `parseDevcontainerPorts()`: Reads `forwardPorts` from
  devcontainer.json, handles numeric and `"host:container"`
  string entries, strips JSONC comments and trailing commas,
  validates port ranges (1-65535), deduplicates against
  excluded ports and within the forwardPorts array itself.
- `stripJSONComments()`: Removes `//` and `/* */` comments
  from JSONC input, preserving string contents.
- `stripTrailingCommas()`: Removes trailing commas before
  `]` or `}` in JSON input, preserving string contents.
- `parsePortEntry()`: Extracts host:container port pair
  from a single `forwardPorts` entry (number or string).

### Modified Capabilities

- `buildRunArgs()`: Calls `parseDevcontainerPorts()` after
  publishing `DefaultServerPort`.
- `buildPersistentRunArgs()`: Calls `parseDevcontainerPorts()`
  after publishing demo ports, deduplicating against both
  `DefaultServerPort` and demo ports.

### Removed Capabilities

(none)

## Impact

- `internal/sandbox/config.go` — new types and functions
  (~200 lines)
- `internal/sandbox/podman.go` — devcontainer port
  integration in `buildPersistentRunArgs()` (~10 lines)
- `internal/sandbox/sandbox_test.go` — 11 new test
  functions (~410 lines)
- `docs/cli-reference.md` — notes for `uf sandbox init`,
  `uf sandbox create`, `uf sandbox start`
- `docs/configuration.md` — note about automatic
  forwardPorts reading
- `CHANGELOG.md` — Unreleased/Fixed entry

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

Internal plumbing within the sandbox package. No
inter-hero artifact interfaces are affected.

### II. Composability First

**Assessment**: PASS

The devcontainer port parsing is self-contained within
the sandbox package. No new external dependencies.

### III. Observable Quality

**Assessment**: PASS

Port mappings are observable via `podman inspect` and
`podman port`. The `-p` flags appear in the generated
argument list, which is testable.

### IV. Testability

**Assessment**: PASS

All functions use dependency injection (`opts.ReadFile`)
for testability. Eleven unit and integration tests cover
happy path, edge cases, JSONC, trailing commas, port
validation, deduplication, and absent files.
