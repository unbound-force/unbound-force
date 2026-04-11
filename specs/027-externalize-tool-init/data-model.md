# Data Model: Externalize Tool Initialization

**Branch**: `027-externalize-tool-init` | **Date**: 2026-04-11

## Overview

This change does not introduce new data structures. It modifies
the behavior of existing functions and removes embedded assets.
The "data model" for this feature is the set of files that move
from embedded assets to external tool ownership.

## Asset Ownership Transfer

### Before: Embedded in `uf` Binary

```text
internal/scaffold/assets/
├── specify/                    # 12 files — deployed by uf init
│   ├── config.yaml
│   ├── scripts/bash/ (5 files)
│   └── templates/ (6 files)
├── openspec/                   # 6 files — deployed by uf init
│   ├── config.yaml
│   └── schemas/unbound-force/ (5 files)
└── opencode/                   # 37 files — deployed by uf init
    ├── agents/ (12 files)
    ├── command/ (15 files)
    ├── skill/ (1 file)
    └── uf/packs/ (9 files)

Total embedded: 55 files
```

### After: Mixed Ownership

```text
internal/scaffold/assets/
├── openspec/                   # 5 files — deployed by uf init
│   └── schemas/unbound-force/ (5 files)
└── opencode/                   # 37 files — deployed by uf init
    ├── agents/ (12 files)
    ├── command/ (15 files)
    ├── skill/ (1 file)
    └── uf/packs/ (9 files)

Total embedded: 42 files

External tool ownership:
├── .specify/                   # Created by: specify init
│   ├── config.yaml
│   ├── scripts/bash/ (5 files)
│   ├── templates/ (6 files)
│   └── memory/ (empty)
├── openspec/config.yaml        # Created by: openspec init
├── openspec/specs/             # Created by: openspec init
├── openspec/changes/           # Created by: openspec init
├── .opencode/agents/gaze-*.md  # Created by: gaze init
└── .opencode/command/gaze*.md  # Created by: gaze init
```

## subToolResult Values

New result entries produced by `initSubTools()`:

| Name | Action | Detail | Condition |
|------|--------|--------|-----------|
| `.specify/` | `initialized` | — | specify in PATH, .specify/ absent |
| `.specify/` | `failed` | `specify init failed` | specify init returns error |
| `openspec/` | `initialized` | — | openspec in PATH, config.yaml absent |
| `openspec/` | `failed` | `openspec init failed` | openspec init returns error |
| `gaze` | `initialized` | — | gaze in PATH, agent file absent |
| `gaze` | `failed` | `gaze init failed` | gaze init returns error |

These follow the same pattern as existing entries:

| Name | Action | Detail | Condition |
|------|--------|--------|-----------|
| `.uf/dewey/` | `initialized` | — | dewey in PATH, .uf/dewey absent |
| `.uf/replicator/` | `initialized` | — | replicator in PATH, .uf/replicator absent |

## stepResult Values (Setup)

New result entries produced by `Run()` in setup:

| Name | Action | Detail | Condition |
|------|--------|--------|-----------|
| `uv` | `already installed` | — | uv in PATH |
| `uv` | `installed` | `via Homebrew` | brew install succeeds |
| `uv` | `installed` | `via curl` | curl install succeeds |
| `uv` | `skipped` | `curl\|bash install requires --yes...` | no Homebrew, no TTY |
| `uv` | `failed` | `brew install failed` | brew install fails |
| `uv` | `failed` | `curl install failed` | curl install fails |
| `Specify CLI` | `already installed` | — | specify in PATH |
| `Specify CLI` | `installed` | `via uv` | uv tool install succeeds |
| `Specify CLI` | `skipped` | `no uv` | uv not available |
| `Specify CLI` | `skipped` | `uv not available...` | uv LookPath fails |
| `Specify CLI` | `failed` | `uv tool install failed...` | uv tool install fails |

## Embedded Asset Manifest Delta

### Removed from `expectedAssetPaths` (13 entries)

```
specify/templates/agent-file-template.md
specify/templates/checklist-template.md
specify/templates/constitution-template.md
specify/templates/plan-template.md
specify/templates/spec-template.md
specify/templates/tasks-template.md
specify/config.yaml
specify/scripts/bash/check-prerequisites.sh
specify/scripts/bash/common.sh
specify/scripts/bash/create-new-feature.sh
specify/scripts/bash/setup-plan.sh
specify/scripts/bash/update-agent-context.sh
openspec/config.yaml
```

### Added to `knownNonEmbeddedFiles` (13 entries)

```
.specify/config.yaml
.specify/templates/agent-file-template.md
.specify/templates/checklist-template.md
.specify/templates/constitution-template.md
.specify/templates/plan-template.md
.specify/templates/spec-template.md
.specify/templates/tasks-template.md
.specify/scripts/bash/check-prerequisites.sh
.specify/scripts/bash/common.sh
.specify/scripts/bash/create-new-feature.sh
.specify/scripts/bash/setup-plan.sh
.specify/scripts/bash/update-agent-context.sh
openspec/config.yaml
```

## Configuration Changes

### opencode.json

No changes. The MCP server configuration is unaffected.

### .uf/config.yaml

No changes. Workflow configuration is unaffected.

### .gitignore

No changes. The standard UF ignore block is unaffected.
