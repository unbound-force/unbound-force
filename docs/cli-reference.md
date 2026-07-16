# CLI Reference

Unbound Force ships two binaries: `uf` (the framework toolkit)
and `mutimind` (the product owner backend). This document covers
every command, flag, and subcommand for both.

---

## uf

`uf` (alias for `unbound-force`) is the CLI tool that scaffolds,
configures, and operates the Unbound Force specification framework.

### uf init

Scaffold the Unbound Force specification framework into the
current directory. Creates Speckit templates, scripts, OpenCode
commands and agents, Divisor review personas, convention packs,
and OpenSpec schema files.

User-owned files (templates, scripts, agents, config) are skipped
if they already exist. Tool-owned files (speckit commands, OpenSpec
schema, convention packs) are updated if their content has changed.

```
uf init [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--force` | bool | `false` | Overwrite all existing files |
| `--divisor` | bool | `false` | Deploy only Divisor review agents and convention packs |
| `--lang` | string | `""` | Project language for convention pack (auto-detected from go.mod, package.json, etc. if omitted) |

> **Note:** Dewey indexing during init uses `--no-embeddings` for
> fast metadata-only indexing. Run `dewey index` separately to
> generate embeddings for semantic search. To skip Dewey entirely,
> set `setup.tools.dewey.method: skip` in `.uf/config.yaml`.

**Example**

```bash
# Scaffold into the current project (auto-detect language)
uf init

# Deploy only The Divisor review agents
uf init --divisor

# Force-overwrite all files for a Go project
uf init --force --lang go
```

### uf version

Print the unbound-force version, commit hash, and build date.

```
uf version
```

No flags. Output format:

```
unbound-force v0.9.0 (commit abc1234, built 2026-05-01T12:00:00Z)
```

### uf doctor

Diagnose the Unbound Force development environment. Checks for
required tools, version managers, scaffolded files, hero
availability, Replicator status, MCP server configuration,
and agent/skill integrity.

Exit code 0 when all checks pass or only warnings exist.
Exit code 1 when any check fails.

```
uf doctor [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format` | string | `"text"` | Output format: `text` or `json` |
| `--dir` | string | `"."` | Target directory to check |

**Example**

```bash
# Terminal-friendly colored report
uf doctor

# JSON output for CI pipelines
uf doctor --format json

# Check a specific project directory
uf doctor --dir /path/to/project
```

### uf setup

Install and configure the Unbound Force development tool chain.
Detects existing version and package managers, installs missing
tools, and scaffolds project files. Idempotent -- safe to run
multiple times.

```
uf setup [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--dry-run` | bool | `false` | Print actions without executing |
| `--yes` | bool | `false` | Skip confirmation prompts |
| `--dir` | string | `"."` | Target directory for setup |

**Example**

```bash
# Interactive setup
uf setup

# Preview what would be installed
uf setup --dry-run

# Non-interactive setup (CI-friendly)
uf setup --yes
```

### uf sandbox

Manage containerized OpenCode development sessions. Supports
Podman (local) and DevPod backends.

```
uf sandbox <subcommand> [flags]
```

#### uf sandbox init

Scaffold a `.devcontainer/devcontainer.json` configuration for use
with DevPod or other devcontainer-compatible tools. The template
includes the OpenCode dev image, gateway proxy environment
variables, and port forwarding for the OpenCode server.

The generated config is **OS-specific**: macOS uses
`--userns=keep-id:uid=1000,gid=1000` for Podman VM UID mapping,
Linux uses plain `--userns=keep-id` to avoid subuid range
conflicts on container restart. `.devcontainer/` is gitignored
because each developer must generate their platform-appropriate
config by running `uf sandbox init`.

```
uf sandbox init [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--image` | string | `""` | Container image (default `"quay.io/unbound-force/opencode-dev:latest"`) |
| `--demo-ports` | int slice | `nil` | Additional ports to forward (comma-separated, e.g., `3000,8080`) |
| `--force` | bool | `false` | Overwrite existing `.devcontainer/devcontainer.json` |

**Example**

```bash
# Scaffold devcontainer with defaults
uf sandbox init

# Override image and expose demo ports
uf sandbox init --image myregistry/dev:latest --demo-ports 3000,8080

# Force-overwrite existing devcontainer.json
uf sandbox init --force
```

#### uf sandbox create

Provision a persistent sandbox workspace for the current project.
Uses DevPod when configured, Podman with named volumes otherwise.
The workspace persists across stop/start cycles.

```
uf sandbox create [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--backend` | string | `"auto"` | Backend: `auto`, `podman`, or `devpod` |
| `--image` | string | `""` | Container image (Podman only; default from `UF_SANDBOX_IMAGE` or `quay.io/unbound-force/opencode-dev:latest`) |
| `--memory` | string | `""` | Memory limit (default `"8g"`) |
| `--cpus` | string | `""` | CPU limit (default `"4"`) |
| `--name` | string | `""` | Workspace name override (default `"uf-sandbox-<project-name>"`) |
| `--detach` | bool | `false` | Start without attaching TUI |
| `--uidmap` | bool | `false` | Use explicit UID/GID mapping (for macOS when Podman machine virtiofs does not support `--userns=keep-id`) |
| `--demo-ports` | int slice | `nil` | Additional ports to expose for demos (comma-separated, e.g., `3000,8080`) |
| `--ide` | string | `"none"` | IDE for DevPod to open: `none`, `vscode`, `openvscode`, `fleet`, `jupyternotebook`, `cursor` |

**Example**

```bash
# Create with defaults (auto-detect backend)
uf sandbox create

# Create with custom resources and detached
uf sandbox create --memory 16g --cpus 8 --detach

# Create with demo ports exposed
uf sandbox create --demo-ports 3000,8080
```

#### uf sandbox destroy

Permanently delete the sandbox workspace and all associated
state (named volumes, DevPod workspace).

```
uf sandbox destroy [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--yes` | bool | `false` | Skip confirmation prompt |
| `--force` | bool | `false` | Force destroy even if workspace is running |

**Example**

```bash
# Destroy with confirmation prompt
uf sandbox destroy

# Non-interactive destroy
uf sandbox destroy --yes
```

#### uf sandbox start

Start a containerized OpenCode session. If a persistent workspace
exists (from `uf sandbox create`), resumes it. Otherwise, starts
an ephemeral container with the current project directory mounted.

```
uf sandbox start [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--mode` | string | `"isolated"` | Mount mode: `isolated` (read-only) or `direct` (read-write) |
| `--detach` | bool | `false` | Start container without attaching TUI |
| `--image` | string | `""` | Container image (default from `UF_SANDBOX_IMAGE` or `quay.io/unbound-force/opencode-dev:latest`) |
| `--memory` | string | `""` | Container memory limit (default `"8g"`) |
| `--cpus` | string | `""` | Container CPU limit (default `"4"`) |
| `--backend` | string | `"auto"` | Backend: `auto`, `podman`, or `devpod` |
| `--no-parent` | bool | `false` | Mount only the project directory (disable parent directory mount) |
| `--uidmap` | bool | `false` | Use explicit UID/GID mapping (for macOS when Podman machine virtiofs does not support `--userns=keep-id`) |
| `--ide` | string | `"none"` | IDE for DevPod to open: `none`, `vscode`, `openvscode`, `fleet`, `jupyternotebook`, `cursor` |

**Example**

```bash
# Start (or resume) in isolated mode
uf sandbox start

# Start with direct read-write mount
uf sandbox start --mode direct

# Start detached with custom backend
uf sandbox start --detach --backend podman
```

#### uf sandbox stop

Stop the running sandbox. For persistent workspaces (created via
`uf sandbox create`), the workspace state is preserved. For
ephemeral containers, the container is removed.

```
uf sandbox stop
```

No flags.

#### uf sandbox attach

Attach the terminal to the running sandbox's OpenCode server via
`opencode attach`. Requires the sandbox to be running and OpenCode
to be installed.

```
uf sandbox attach
```

No flags.

#### uf sandbox extract

Generate a patch from the container's git history, present it for
review, and apply it to the host repo on confirmation. Uses
`git format-patch` / `git am` for commit-preserving round-trip
extraction.

```
uf sandbox extract [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--yes` | bool | `false` | Skip confirmation prompt |

**Example**

```bash
# Extract with review prompt
uf sandbox extract

# Extract without confirmation
uf sandbox extract --yes
```

#### uf sandbox status

Display the current state of the sandbox workspace including
workspace name, backend, image, state, project directory, server
URL, demo endpoints, and uptime.

```
uf sandbox status
```

No flags.

### uf gateway

Start and manage a local LLM reverse proxy that serves the
Anthropic Messages API. The gateway auto-detects the cloud
provider from environment variables and injects host-side
credentials into upstream requests.

Supported providers:
- Anthropic (`ANTHROPIC_API_KEY`)
- Vertex AI (`CLAUDE_CODE_USE_VERTEX=1` + `ANTHROPIC_VERTEX_PROJECT_ID`)
- AWS Bedrock (`CLAUDE_CODE_USE_BEDROCK=1`)

#### uf gateway (start)

Start the gateway. When run without a subcommand, `uf gateway`
starts the proxy.

```
uf gateway [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--port` | int | `53147` | Port to listen on |
| `--provider` | string | `""` | Provider override: `anthropic`, `vertex`, or `bedrock` (auto-detected if omitted) |
| `--detach` | bool | `false` | Run gateway in the background |

**Example**

```bash
# Start with auto-detected provider
uf gateway

# Start on a custom port, detached
uf gateway --port 8080 --detach

# Force a specific provider
uf gateway --provider vertex
```

#### uf gateway stop

Terminate a running background gateway and remove its PID file.
Prints "No gateway running." if no gateway is found.

```
uf gateway stop
```

No flags.

#### uf gateway status

Display the running gateway's provider, port, PID, and uptime.
Prints "No gateway running." if no gateway is found.

```
uf gateway status
```

No flags.

### uf config

Manage the unified `.uf/config.yaml` configuration file.

```
uf config <subcommand> [flags]
```

#### uf config init

Create or update the `.uf/config.yaml` file. All values are
commented out by default -- uncomment what you want to change.
If the file already exists, new sections are added and deprecated
sections are removed (a backup is saved to `.uf/config.yaml.bak`).

```
uf config init [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--dir` | string | `"."` | Target directory |

**Example**

```bash
# Create config in the current project
uf config init
```

#### uf config show

Display the effective configuration after all layers merge
(compiled defaults, config file overrides, environment variables).

```
uf config show [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--dir` | string | `"."` | Target directory |
| `--format` | string | `"text"` | Output format: `text` (YAML) or `json` |

**Example**

```bash
# Show effective config as YAML
uf config show

# Show effective config as JSON
uf config show --format json
```

#### uf config validate

Validate the `.uf/config.yaml` file against known field values.
If no config file exists, reports that compiled defaults are used
(which is valid).

```
uf config validate [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--dir` | string | `"."` | Target directory |
| `--format` | string | `"text"` | Output format: `text` or `json` |

**Example**

```bash
# Validate config
uf config validate

# Validate with JSON output (CI-friendly)
uf config validate --format json
```

---

## mutimind

`mutimind` is the Muti-Mind product owner backend CLI for backlog
management and GitHub issue synchronization. It stores backlog
items as YAML files under `.uf/muti-mind/backlog/` and produces
JSON artifacts under `.uf/muti-mind/artifacts/`.

### Global Flags

These flags apply to all `mutimind` subcommands:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format` | string | `"text"` | Output format: `text` or `json` |
| `--backlog-dir` | string | `".uf/muti-mind/backlog"` | Backlog directory |
| `--artifacts-dir` | string | `".uf/muti-mind/artifacts"` | Artifacts directory |

### mutimind init

Initialize the Muti-Mind environment. Creates the backlog and
artifacts directories and writes a default `config.yaml` if one
does not exist.

```
mutimind init
```

No command-specific flags.

**Example**

```bash
mutimind init
```

### mutimind add

Add a new backlog item. Assigns the next available ID
automatically.

```
mutimind add [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--type` | string | `""` | Item type: `epic`, `story`, `task`, or `bug` (required) |
| `--title` | string | `""` | Item title (required) |
| `--priority` | string | `"P3"` | Priority: `P1` through `P5` |
| `--description` | string | `""` | Item description |

**Example**

```bash
# Add a high-priority story
mutimind add --type story --title "User authentication" --priority P1

# Add a bug with description
mutimind add --type bug --title "Login timeout" \
  --description "Session expires after 30 seconds"
```

### mutimind list

List backlog items with optional filtering.

```
mutimind list [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--status` | string | `""` | Filter by status |
| `--sprint` | string | `""` | Filter by sprint |

**Example**

```bash
# List all items
mutimind list

# List only in-progress items
mutimind list --status in-progress

# List items in a specific sprint, as JSON
mutimind list --sprint sprint-3 --format json
```

### mutimind update

Update fields on an existing backlog item.

```
mutimind update <id> [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--priority` | string | `""` | New priority |
| `--status` | string | `""` | New status |
| `--sprint` | string | `""` | New sprint |

**Example**

```bash
# Move an item to in-progress
mutimind update BLI-003 --status in-progress

# Reprioritize and assign to a sprint
mutimind update BLI-003 --priority P1 --sprint sprint-2
```

### mutimind show

Show full details of a backlog item.

```
mutimind show <id>
```

No command-specific flags (respects the global `--format` flag).

**Example**

```bash
# Human-readable output
mutimind show BLI-003

# JSON output
mutimind show BLI-003 --format json
```

### mutimind sync-push

Push local backlog items to GitHub Issues. If an item ID is
provided, pushes only that item. Otherwise, pushes all items
that need syncing.

```
mutimind sync-push [id]
```

No command-specific flags.

**Example**

```bash
# Push all items
mutimind sync-push

# Push a specific item
mutimind sync-push BLI-003
```

### mutimind sync-pull

Pull GitHub Issues into the local backlog. Creates local backlog
items for issues that do not yet have a corresponding local entry.

```
mutimind sync-pull
```

No command-specific flags.

### mutimind sync-status

Report on the synchronization state between local backlog items
and GitHub Issues.

```
mutimind sync-status
```

No command-specific flags.

### mutimind sync

Perform a bidirectional sync including conflict detection. Pushes
local changes and pulls remote changes, reporting any conflicts.

```
mutimind sync
```

No command-specific flags.

### mutimind sync-project

Sync backlog items with GitHub Projects.

```
mutimind sync-project
```

No command-specific flags.

### mutimind generate-artifact

Generate a JSON artifact envelope for a backlog item. The artifact
is written to the artifacts directory.

```
mutimind generate-artifact <item_id>
```

No command-specific flags.

**Example**

```bash
mutimind generate-artifact BLI-003
```

### mutimind decide

Generate an acceptance-decision artifact for a backlog item.
Records whether the item is accepted, rejected, or conditionally
accepted, along with the rationale and criteria evaluation.

```
mutimind decide [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--item` | string | `""` | Backlog item ID (required) |
| `--decision` | string | `""` | Decision: `accept`, `reject`, or `conditional` (required) |
| `--rationale` | string | `""` | Rationale for the decision |
| `--report-ref` | string | `""` | Gaze report reference |
| `--met` | string slice | `nil` | Acceptance criteria met (comma-separated) |
| `--failed` | string slice | `nil` | Acceptance criteria failed (comma-separated) |

**Example**

```bash
# Accept an item with rationale
mutimind decide --item BLI-003 --decision accept \
  --rationale "All acceptance criteria verified by Gaze" \
  --report-ref ".uf/artifacts/quality-report/BLI-003.json" \
  --met "AC-1,AC-2,AC-3"

# Conditionally accept with failed criteria
mutimind decide --item BLI-005 --decision conditional \
  --rationale "Coverage below threshold" \
  --met "AC-1,AC-2" --failed "AC-3"
```
