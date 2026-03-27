## Why

`uf init` (the Go binary) scaffolds project files but
cannot customize third-party tool files that are owned
by other packages. Specifically, the OpenSpec CLI
(`@fission-ai/openspec`) deploys its own skill files
when installed. These files need project-specific
customizations:

1. **Branch enforcement**: OpenSpec skills need
   `opsx/<name>` branch creation/validation/cleanup
   (matching Speckit's branch discipline)
2. **Dewey context**: OpenSpec skills need instructions
   to query Dewey for related context before drafting
   proposals, implementing changes, or exploring ideas
3. **3-tier degradation**: OpenSpec skills need the
   Dewey fallback pattern (Tier 3 full, Tier 2
   graph-only, Tier 1 file reads) so they work without
   Dewey

These customizations require LLM reasoning to find
the correct insertion points in files whose structure
may change with OpenSpec CLI updates. A Go binary
cannot do this reliably -- but an OpenCode slash
command can.

## What Changes

Create a `/uf-init` slash command
(`.opencode/command/uf-init.md`) that:

1. Checks prerequisites (OpenSpec skill and command
   files exist)
2. Reads each target file and checks if customizations
   are already present (idempotent)
3. Uses LLM reasoning to find the correct insertion
   points and insert customizations
4. Reports what was found, applied, skipped, and any
   errors

The command file is scaffolded by `uf init` (Go binary)
so it's available in every project.

## Capabilities

### New Capabilities
- `/uf-init` slash command: LLM-driven customization
  of third-party tool files
- Branch enforcement in OpenSpec skills and commands
- Dewey context queries in OpenSpec skills and commands
- 3-tier Dewey degradation in OpenSpec skills

### Modified Capabilities
- Scaffold asset count increases by 1

### Removed Capabilities
- None

## Impact

- `.opencode/command/uf-init.md` -- NEW (live copy)
- `internal/scaffold/assets/opencode/command/uf-init.md`
  -- NEW (scaffold asset)
- `internal/scaffold/scaffold_test.go` -- asset count
- `cmd/unbound-force/main_test.go` -- file count

## Constitution Alignment

### I. Autonomous Collaboration
**Assessment**: PASS -- customizations are applied to
Markdown files (artifacts), not runtime coupling.

### II. Composability First
**Assessment**: PASS -- `/uf-init` is optional. All
tools work without it. The customizations enhance but
don't create dependencies.

### III. Observable Quality
**Assessment**: PASS -- the command reports what it
found and did with status indicators.

### IV. Testability
**Assessment**: PASS -- the command is a Markdown file
(no Go code to test). The scaffold asset tests verify
it's deployed. The customizations themselves are
verifiable by reading the target files.
