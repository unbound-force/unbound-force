## Context

The devcontainer spec defines `forwardPorts` as
`Array<number | string>` in `devcontainer.json`, using
JSONC (JSON with Comments) as the canonical format.
Podman sandbox commands (`uf sandbox create`,
`uf sandbox start`) need to translate these entries into
`-p host:container` flags for `podman run`.

On macOS with `UserModeNetworking: true`, container IPs
are not routable from the host. Explicit port publishing
is the only way to reach services inside the container.

## Goals / Non-Goals

### Goals

- Parse `forwardPorts` from `.devcontainer/devcontainer.json`
  in both ephemeral and persistent sandbox paths
- Support numeric ports (8080) and string host:container
  mappings ("8080:3000")
- Handle JSONC comments and trailing commas
- Validate port ranges (1-65535)
- Deduplicate against `DefaultServerPort`, demo ports, and
  duplicate entries within the forwardPorts array itself

### Non-Goals

- Full devcontainer spec compliance beyond `forwardPorts`
- Changing DevPod backend behavior — DevPod manages its
  own port forwarding
- Runtime port forwarding after container creation

## Decisions

### D1: JSONC handling via stripping

Strip comments and trailing commas before standard
`json.Unmarshal` rather than importing a JSONC-aware
parser. This avoids a new dependency and is sufficient
for the subset of JSONC used in devcontainer.json.

### D2: First-wins deduplication

When duplicate host ports appear in `forwardPorts`
(e.g., `[8080, "8080:3000"]`), the first entry wins.
A `seen` map tracks emitted host ports. This prevents
podman "port already bound" errors.

### D3: Graceful degradation

If `devcontainer.json` is absent, unreadable, or
contains no `forwardPorts`, the helper returns nil
silently. No error is surfaced to the user. Invalid
individual entries are skipped without affecting valid
ones.

### D4: ReadFile injection

The file is read via `opts.ReadFile` (dependency
injection) rather than direct `os.ReadFile`, maintaining
testability without filesystem access.

## Risks / Trade-offs

- **Risk**: JSONC stripping is a best-effort parser, not
  a full JSONC implementation. Deeply nested comments
  inside strings with escape sequences could theoretically
  confuse it.
  **Mitigation**: The string-literal detection preserves
  backslash escapes correctly. Devcontainer.json files
  in practice use simple comment patterns.

- **Trade-off**: First-wins deduplication means if a user
  writes `[8080, "8080:3000"]`, the plain port (8080:8080)
  is used, not the host:container mapping (8080:3000).
  This is deterministic and documented; reordering entries
  changes semantics, which matches the array-ordered
  semantics of the devcontainer spec.
