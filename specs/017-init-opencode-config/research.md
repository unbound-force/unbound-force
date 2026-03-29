# Research: Init OpenCode Config

## R1: opencode.json MCP Entry Format

**Decision**: Use the `"mcp"` key with nested server
objects. Each server has `type`, `command` (array),
and `enabled` fields.

**Rationale**: This matches the actual `opencode.json`
in the repo root, which OpenCode reads successfully.
The `"mcpServers"` format with string `command` + array
`args` was an older convention.

**Alternatives considered**:
- `"mcpServers"` key with string command + args array:
  rejected because the current file uses `"mcp"` and
  works.
- Supporting both in init: rejected because init should
  write the canonical format only. Doctor reads both
  for backward compatibility.

## R2: Swarm Plugin Detection

**Decision**: Use `.hive/` directory existence as proxy
for swarm plugin availability.

**Rationale**: The `.hive/` directory is created by
`swarm init` which runs in `initSubTools()`. If it
exists, the swarm plugin is configured and ready.
The actual npm package `opencode-swarm-plugin` may not
be resolvable via `which` because it's an npm package,
not a standalone binary.

**Alternatives considered**:
- `which opencode-swarm-plugin`: rejected because npm
  packages don't always have standalone binaries in
  PATH.
- Always add the plugin entry: rejected because it
  would add a non-functional entry in repos without
  swarm.

## R3: Existing configureOpencodeJSON() Pattern

**Decision**: Port the read-modify-write pattern from
`setup.go` to `scaffold.go`, expanding it to handle
both MCP server entries and the plugin array.

**Rationale**: The existing `configureOpencodeJSON()`
in setup.go uses `map[string]json.RawMessage` to
preserve unknown keys, checks for existing entries
before adding, and uses `json.MarshalIndent` for
readable output. This pattern is proven and testable.

**Alternatives considered**:
- Writing a new JSON manipulation approach: rejected
  because the existing pattern works and is tested.
- Using a typed struct instead of RawMessage: rejected
  because it would not preserve unknown user keys.

## R4: Doctor checkMCPConfig() Fix

**Decision**: Check both `"mcp"` and `"mcpServers"`
keys (canonical first, legacy fallback). Parse the
command field as either string or array (first element
= binary name).

**Rationale**: The current check only looks for
`"mcpServers"` (line 666 in checks.go), which means
it never finds the Dewey server entry in the canonical
`"mcp"` format. The command field struct uses
`Command string` + `Args []string`, but the actual
format is `"command": ["dewey", "serve", "--vault", "."]`.
Both issues must be fixed together.

**Alternatives considered**:
- Only fix the key and not the command format:
  rejected because both are broken and fixing one
  without the other still produces incorrect results.

## R5: Setup Step Removal

**Decision**: Remove the opencode.json step entirely
from setup.go and renumber from 16 to 15 total steps.

**Rationale**: Setup's final step runs `uf init`,
which now handles opencode.json. Keeping a placeholder
step adds noise. The step count change affects progress
messages (`[N/16]` → `[N/15]`) and any tests that
assert on step count.

**Alternatives considered**:
- Keep a "managed by uf init" placeholder step:
  rejected to reduce output noise.

## R6: Injectable File I/O for scaffold.Options

**Decision**: Add `ReadFile` and `WriteFile` function
fields to `scaffold.Options`, defaulting to
`os.ReadFile` and `os.WriteFile` in `Run()`.

**Rationale**: `initSubTools()` currently uses
`os.Stat`/`os.WriteFile` directly. Adding
`configureOpencodeJSON()` requires reading and writing
JSON files, which must be testable without real
filesystem side effects beyond `t.TempDir()`. The
setup package already uses this pattern
(`opts.ReadFile`, `opts.WriteFile`).

**Alternatives considered**:
- Using `os.ReadFile`/`os.WriteFile` directly in tests
  with `t.TempDir()`: viable but less isolated than
  injection. The injection approach is already the
  project convention in setup.go.
