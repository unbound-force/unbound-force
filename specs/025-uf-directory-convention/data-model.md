# Data Model: Unified .uf/ Directory Convention

**Branch**: `025-uf-directory-convention` | **Date**: 2026-04-06

## Overview

This spec does not introduce new data types, schemas, or
persistent state. It is a pure path renaming operation.
All existing data structures (WorkflowInstance, WorkflowConfig,
MetricsSnapshot, backlog items) remain unchanged — only the
filesystem paths where they are stored change.

## Path Mapping Table

### Per-Repo Tool Directories

| Old Path | New Path | Owner |
|----------|----------|-------|
| `.dewey/` | `.uf/dewey/` | Dewey |
| `.hive/` | `.uf/replicator/` | Replicator |
| `.unbound-force/` | `.uf/` | Unbound Force (root) |
| `.unbound-force/workflows/` | `.uf/workflows/` | Orchestration engine |
| `.unbound-force/artifacts/` | `.uf/artifacts/` | Artifact envelope I/O |
| `.unbound-force/config.yaml` | `.uf/config.yaml` | Workflow config |
| `.muti-mind/` | `.uf/muti-mind/` | Muti-Mind |
| `.muti-mind/backlog/` | `.uf/muti-mind/backlog/` | Muti-Mind |
| `.muti-mind/artifacts/` | `.uf/muti-mind/artifacts/` | Muti-Mind |
| `.muti-mind/config.yaml` | `.uf/muti-mind/config.yaml` | Muti-Mind |
| `.mx-f/` | `.uf/mx-f/` | Mx F |
| `.mx-f/data/` | `.uf/mx-f/data/` | Mx F metrics |
| `.mx-f/impediments/` | `.uf/mx-f/impediments/` | Mx F impediments |

### Convention Pack Directory

| Old Path | New Path |
|----------|----------|
| `.opencode/unbound/packs/` | `.opencode/uf/packs/` |

### Scaffold Asset Directory

| Old Path | New Path |
|----------|----------|
| `internal/scaffold/assets/opencode/unbound/packs/` | `internal/scaffold/assets/opencode/uf/packs/` |

### Hero Manifest

| Old Path | New Path |
|----------|----------|
| `.unbound-force/hero.json` | `.uf/hero.json` |

## Directory Structure After Migration

```text
.uf/                          # Root per-repo directory
├── config.yaml               # Workflow configuration
├── workflows/                # Orchestration workflow state
│   └── wf-*.json
├── artifacts/                # Hero artifact envelopes
│   ├── quality-report/
│   ├── review-verdict/
│   └── ...
├── hero.json                 # Hero manifest (in hero repos)
├── dewey/                    # Dewey workspace (created by dewey init)
│   ├── graph.db
│   └── sources.yaml
├── replicator/               # Replicator data (created by replicator init)
│   ├── cells.db
│   └── ...
├── muti-mind/                # Muti-Mind data
│   ├── backlog/
│   ├── artifacts/
│   └── config.yaml
└── mx-f/                     # Mx F data
    ├── data/
    ├── impediments/
    └── retros/

.opencode/
├── agents/                   # Agent files (unchanged)
├── command/                  # Command files (unchanged)
├── skill/                    # Skill packages (unchanged)
├── skills/                   # OpenSpec skills (unchanged)
└── uf/                       # Convention packs (renamed from unbound/)
    └── packs/
        ├── go.md
        ├── go-custom.md
        ├── typescript.md
        ├── typescript-custom.md
        ├── default.md
        ├── default-custom.md
        ├── severity.md
        ├── content.md
        └── content-custom.md
```

## .gitignore Changes

```gitignore
# Old (remove):
.unbound-force/
.dewey/

# New (add):
.uf/
```

The single `.uf/` entry covers all tool subdirectories
(dewey, replicator, muti-mind, mx-f, workflows, artifacts).

## Schema Changes

### hero-manifest/v1.0.0.schema.json

The `description` field references `.unbound-force/hero.json`.
Update to `.uf/hero.json`.

No structural schema changes. The schema itself is unchanged
— only the description text that documents the canonical
file location.

### acceptance-decision sample

The sample file references
`.unbound-force/artifacts/quality-report/...`. Update to
`.uf/artifacts/quality-report/...`.

## Constants and Config Values

### scaffold.go: workflowConfigContent

```go
// Old:
const workflowConfigContent = `# .unbound-force/config.yaml
...`

// New:
const workflowConfigContent = `# .uf/config.yaml
...`
```

### scaffold.go: initSubTools paths

```go
// Old:
ufDir := filepath.Join(opts.TargetDir, ".unbound-force")
deweyDir := filepath.Join(opts.TargetDir, ".dewey")
hiveDir := filepath.Join(opts.TargetDir, ".hive")

// New:
ufDir := filepath.Join(opts.TargetDir, ".uf")
deweyDir := filepath.Join(opts.TargetDir, ".uf", "dewey")
replicatorDir := filepath.Join(opts.TargetDir, ".uf", "replicator")
```

### scaffold.go: isConventionPack

```go
// Old:
func isConventionPack(relPath string) bool {
    return strings.HasPrefix(relPath, "opencode/unbound/packs/")
}

// New:
func isConventionPack(relPath string) bool {
    return strings.HasPrefix(relPath, "opencode/uf/packs/")
}
```

### orchestration/engine.go: Orchestrator fields

```go
// Old:
WorkflowDir string  // .unbound-force/workflows/
ArtifactDir string  // .unbound-force/artifacts/

// New:
WorkflowDir string  // .uf/workflows/
ArtifactDir string  // .uf/artifacts/
```

### cmd/mutimind/main.go: Default flag values

```go
// Old:
rootCmd.PersistentFlags().StringVar(&params.BacklogDir,
    "backlog-dir", ".muti-mind/backlog", "Backlog directory")
rootCmd.PersistentFlags().StringVar(&params.ArtifactsDir,
    "artifacts-dir", ".muti-mind/artifacts", "Artifacts directory")

// New:
rootCmd.PersistentFlags().StringVar(&params.BacklogDir,
    "backlog-dir", ".uf/muti-mind/backlog", "Backlog directory")
rootCmd.PersistentFlags().StringVar(&params.ArtifactsDir,
    "artifacts-dir", ".uf/muti-mind/artifacts", "Artifacts directory")
```

### doctor/checks.go: checkScaffoldedFiles

```go
// Old:
packsDir := filepath.Join(opts.TargetDir,
    ".opencode", "unbound", "packs")
// display name: ".opencode/unbound/packs/"

// New:
packsDir := filepath.Join(opts.TargetDir,
    ".opencode", "uf", "packs")
// display name: ".opencode/uf/packs/"
```

### doctor/checks.go: checkDewey

```go
// Old:
deweyDir := filepath.Join(opts.TargetDir, ".dewey")

// New:
deweyDir := filepath.Join(opts.TargetDir, ".uf", "dewey")
```

### doctor/checks.go: checkReplicator

```go
// Old:
hivePath := filepath.Join(opts.TargetDir, ".hive")
// display name: ".hive/"

// New:
replicatorPath := filepath.Join(opts.TargetDir,
    ".uf", "replicator")
// display name: ".uf/replicator/"
```
