# Contract: Scaffold Engine Changes

**Branch**: `027-externalize-tool-init` | **Date**: 2026-04-11

## Embedded Asset Removal

### Files Removed from `internal/scaffold/assets/specify/`

Remove the entire `specify/` directory (12 files):

```text
specify/
├── config.yaml                          # REMOVE
├── scripts/
│   └── bash/
│       ├── check-prerequisites.sh       # REMOVE
│       ├── common.sh                    # REMOVE
│       ├── create-new-feature.sh        # REMOVE
│       ├── setup-plan.sh                # REMOVE
│       └── update-agent-context.sh      # REMOVE
└── templates/
    ├── agent-file-template.md           # REMOVE
    ├── checklist-template.md            # REMOVE
    ├── constitution-template.md         # REMOVE
    ├── plan-template.md                 # REMOVE
    ├── spec-template.md                 # REMOVE
    └── tasks-template.md               # REMOVE
```

### Files Removed from `internal/scaffold/assets/openspec/`

Remove 1 file, keep 5:

```text
openspec/
├── config.yaml                          # REMOVE
└── schemas/
    └── unbound-force/
        ├── schema.yaml                  # KEEP (project-specific)
        └── templates/
            ├── proposal.md              # KEEP
            ├── spec.md                  # KEEP
            ├── design.md               # KEEP
            └── tasks.md                # KEEP
```

### Post-Removal Asset Count

Before: 55 embedded assets
Removed: 13 (12 specify + 1 openspec/config.yaml)
After: 42 embedded assets

## Function Changes

### initSubTools(opts *Options) []subToolResult

Add three new delegation blocks after the existing Replicator
delegation and before `configureOpencodeJSON()`. Each follows
the established Dewey/Replicator pattern.

#### 1. Specify CLI Delegation

```go
// Specify: init if binary available and .specify/ absent.
if _, err := opts.LookPath("specify"); err == nil {
    specifyDir := filepath.Join(opts.TargetDir, ".specify")
    if _, statErr := os.Stat(specifyDir); os.IsNotExist(statErr) {
        _, _ = fmt.Fprintf(opts.Stdout, "  Initializing Speckit framework...\n")
        if _, initErr := opts.ExecCmd("specify", "init"); initErr != nil {
            results = append(results, subToolResult{
                name: ".specify/", action: "failed",
                detail: "specify init failed"})
        } else {
            results = append(results, subToolResult{
                name: ".specify/", action: "initialized"})
        }
    }
}
```

**Gate**: `LookPath("specify")` + `os.Stat(".specify/")`
**Command**: `specify init`
**Result name**: `.specify/`

#### 2. OpenSpec CLI Delegation

```go
// OpenSpec: init if binary available and openspec/config.yaml absent.
// Gate on config.yaml (not openspec/ directory) because the
// embedded custom schema creates openspec/schemas/ before
// initSubTools() runs.
if _, err := opts.LookPath("openspec"); err == nil {
    openspecConfig := filepath.Join(opts.TargetDir, "openspec", "config.yaml")
    if _, statErr := os.Stat(openspecConfig); os.IsNotExist(statErr) {
        _, _ = fmt.Fprintf(opts.Stdout, "  Initializing OpenSpec framework...\n")
        if _, initErr := opts.ExecCmd("openspec", "init", "--tools", "opencode"); initErr != nil {
            results = append(results, subToolResult{
                name: "openspec/", action: "failed",
                detail: "openspec init failed"})
        } else {
            results = append(results, subToolResult{
                name: "openspec/", action: "initialized"})
        }
    }
}
```

**Gate**: `LookPath("openspec")` + `os.Stat("openspec/config.yaml")`
**Command**: `openspec init --tools opencode`
**Result name**: `openspec/`

Design decision: Gate on `openspec/config.yaml` rather than
`openspec/` directory because the embedded asset walk creates
`openspec/schemas/unbound-force/` before `initSubTools()` runs.
If we gated on the directory, the delegation would be skipped
even when `openspec init` hasn't run yet.

#### 3. Gaze CLI Delegation

```go
// Gaze: init if binary available and gaze agent file absent.
if _, err := opts.LookPath("gaze"); err == nil {
    gazeAgent := filepath.Join(opts.TargetDir, ".opencode", "agents", "gaze-reporter.md")
    if _, statErr := os.Stat(gazeAgent); os.IsNotExist(statErr) {
        _, _ = fmt.Fprintf(opts.Stdout, "  Initializing Gaze integration...\n")
        if _, initErr := opts.ExecCmd("gaze", "init"); initErr != nil {
            results = append(results, subToolResult{
                name: "gaze", action: "failed",
                detail: "gaze init failed"})
        } else {
            results = append(results, subToolResult{
                name: "gaze", action: "initialized"})
        }
    }
}
```

**Gate**: `LookPath("gaze")` + `os.Stat(".opencode/agents/gaze-reporter.md")`
**Command**: `gaze init`
**Result name**: `gaze`

Design decision: Gate on the agent file rather than a directory
because Gaze creates files inside `.opencode/` (which already
exists from the asset walk), not a separate workspace directory.

### Delegation Order in initSubTools()

```text
1. .uf/config.yaml creation (existing)
2. Dewey init + index (existing)
3. Replicator init (existing)
4. Specify init (NEW)
5. OpenSpec init (NEW)
6. Gaze init (NEW)
7. configureOpencodeJSON() (existing)
```

New delegations are placed after Replicator (infrastructure
tools first, then development workflow tools) and before
`configureOpencodeJSON()` (which must run last to detect all
installed tools).

### Empty Directory Creation

The existing empty directory creation code in `Run()` (lines
200–209) creates `openspec/specs/` and `openspec/changes/`.
This code is KEPT — it's idempotent and useful when `openspec`
is not installed.

## Test Changes

### expectedAssetPaths

Remove 13 entries:

```go
// REMOVE: Speckit templates (6)
"specify/templates/agent-file-template.md",
"specify/templates/checklist-template.md",
"specify/templates/constitution-template.md",
"specify/templates/plan-template.md",
"specify/templates/spec-template.md",
"specify/templates/tasks-template.md",
// REMOVE: Speckit config (1)
"specify/config.yaml",
// REMOVE: Speckit scripts (5)
"specify/scripts/bash/check-prerequisites.sh",
"specify/scripts/bash/common.sh",
"specify/scripts/bash/create-new-feature.sh",
"specify/scripts/bash/setup-plan.sh",
"specify/scripts/bash/update-agent-context.sh",
// REMOVE: OpenSpec config (1)
"openspec/config.yaml",
```

### knownNonEmbeddedFiles

Add entries for files created by `specify init`:

```go
// Speckit files — created by specify init, not scaffolded by uf init
".specify/config.yaml":                              true,
".specify/templates/agent-file-template.md":          true,
".specify/templates/checklist-template.md":           true,
".specify/templates/constitution-template.md":        true,
".specify/templates/plan-template.md":                true,
".specify/templates/spec-template.md":                true,
".specify/templates/tasks-template.md":               true,
".specify/scripts/bash/check-prerequisites.sh":       true,
".specify/scripts/bash/common.sh":                    true,
".specify/scripts/bash/create-new-feature.sh":        true,
".specify/scripts/bash/setup-plan.sh":                true,
".specify/scripts/bash/update-agent-context.sh":      true,
// OpenSpec config — created by openspec init
"openspec/config.yaml":                               true,
```

### TestCanonicalSources_AreEmbedded

Update `canonicalDirs` to remove `.specify/` from the walk
(or add all `.specify/` files to `knownNonEmbeddedFiles`).
The standalone config file check for `.specify/config.yaml`
and `openspec/config.yaml` must also be updated.

### New Test Functions

Add tests for each new delegation (success, skip, and
failure paths):

1. `TestInitSubTools_SpecifyInit` — verify `specify init` is
   called when binary available and `.specify/` absent
2. `TestInitSubTools_SpecifySkipped` — verify skip when
   `.specify/` exists
3. `TestInitSubTools_SpecifyNotInstalled` — verify skip when
   binary not in PATH
4. `TestInitSubTools_SpecifyFailed` — verify `subToolResult`
   with action `"failed"` when `specify init` returns error
5. `TestInitSubTools_OpenSpecInit` — verify `openspec init
   --tools opencode` is called
6. `TestInitSubTools_OpenSpecSkipped` — verify skip when
   `openspec/config.yaml` exists
7. `TestInitSubTools_OpenSpecFailed` — verify `subToolResult`
   with action `"failed"` when `openspec init` returns error
8. `TestInitSubTools_GazeInit` — verify `gaze init` is called
9. `TestInitSubTools_GazeSkipped` — verify skip when agent
   file exists
10. `TestInitSubTools_GazeFailed` — verify `subToolResult`
    with action `"failed"` when `gaze init` returns error

### mapAssetPath and knownAssetPrefixes

The `knownAssetPrefixes` list includes `"specify/"`. After
removing all specify assets, no files use this prefix. Two
options:
1. Remove `"specify/"` from `knownAssetPrefixes` — cleaner
2. Keep it — harmless, documents historical mapping

**Decision**: Remove `"specify/"` from `knownAssetPrefixes`.
The mapping code in `mapAssetPath()` can also remove the
`specify/` case, but keeping it is harmless (dead code that
documents the historical mapping). For cleanliness, remove
both.
