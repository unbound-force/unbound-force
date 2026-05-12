# Configuration

## Overview

Unbound Force uses a unified configuration system based on the
file `.uf/config.yaml`. This file is **optional** -- when it is
absent, compiled defaults are used with no error. Configuration
is layered: multiple sources are merged together, with
higher-priority sources overriding lower-priority ones.

## Configuration Hierarchy

Values are resolved in the following order, from highest to
lowest priority:

```
CLI flags (highest priority)
  > Environment variables
    > .uf/config.yaml (repo-level)
      > ~/.config/uf/config.yaml (user-level)
        > Compiled defaults (lowest priority)
```

Each layer only needs to specify the values it wants to
override. Unspecified fields fall through to the next layer.
Scalar fields are replaced by higher-priority layers. Slice
fields (like `skip` lists) are replaced entirely, not appended.
Map fields (like `tools`) are merged key-by-key.

## Getting Started

Three subcommands manage the configuration file:

```bash
# Create .uf/config.yaml with a commented-out template
uf config init

# Display the effective merged config (all layers resolved)
uf config show

# Validate config file values against known constraints
uf config validate
```

`uf config init` is idempotent. If a config file already exists,
it detects added and removed sections, updates the file, and
saves a backup to `.uf/config.yaml.bak`. If the file is already
current, no changes are made.

`uf config show` accepts a `--format` flag (`text` or `json`)
to control output format. The default is `text` (YAML).

`uf config validate` checks the file for invalid field values
(unknown enum values, out-of-range ports). A missing file is
considered valid because compiled defaults are used.

## Configuration Sections

The config file has seven top-level sections. Each section is
independent and can be included or omitted as needed.

---

### scaffold

Controls what `uf init` deploys when scaffolding a new project.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `language` | string | `auto` | Target language for scaffold templates. Auto-detected from the project if omitted. |

Valid values for `language`: `auto`, `go`, `typescript`,
`python`, `rust`.

```yaml
scaffold:
  language: go
```

---

### doctor

Controls `uf doctor` check behavior. When no fields are set,
all checks run with their built-in severity levels.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `skip` | list of strings | `[]` | Check names to skip entirely. |
| `tools` | map of string to string | `{}` | Override the severity of individual tool checks. |

Valid values for tool severity: `required`, `recommended`,
`optional`.

```yaml
doctor:
  skip:
    - ollama
  tools:
    gaze: optional
    dewey: recommended
```

---

### setup

Controls how `uf setup` installs tools.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `package_manager` | string | `auto` | System package manager to use. |
| `skip` | list of strings | `[]` | Tool names to skip during installation. |
| `tools` | map of string to `ToolConfig` | `{}` | Per-tool install method overrides. |

**ToolConfig fields:**

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `method` | string | (none) | Install method for this tool. |
| `version` | string | (none) | Target version when installing. |

Valid values for `package_manager`: `auto`, `homebrew`, `dnf`,
`apt`, `manual`.

Valid values for `method`: `auto`, `homebrew`, `dnf`, `rpm`,
`apt`, `curl`, `skip`, `nvm`, `fnm`, `mise`.

**Environment variable override:**
`UF_PACKAGE_MANAGER` overrides `setup.package_manager`.

```yaml
setup:
  package_manager: dnf
  skip:
    - ollama
    - dewey
  tools:
    opencode:
      method: curl
    node:
      method: fnm
      version: "22"
    gaze:
      method: homebrew
```

---

### sandbox

Controls `uf sandbox` behavior for containerized development
sessions.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `runtime` | string | `auto` | Container runtime. |
| `backend` | string | `auto` | Sandbox backend. |
| `image` | string | `quay.io/unbound-force/opencode-dev:latest` | Container image to use. |
| `resources.memory` | string | `8g` | Memory limit for the container. |
| `resources.cpus` | string | `4` | CPU limit for the container. |
| `mode` | string | `isolated` | Sandbox isolation mode. |
| `demo_ports` | list of ints | `[]` | Ports to expose from the container. |
| `uid_map` | bool | `false` | Enable UID mapping for rootless containers. |
| `ide` | string | `none` | IDE for DevPod to open after workspace creation. |

Valid values for `runtime`: `auto`, `podman`, `docker`.

Valid values for `backend`: `auto`, `podman`, `devpod`.

Valid values for `mode`: `isolated`, `direct`.

Valid values for `ide`: `none`, `vscode`, `openvscode`,
`fleet`, `jupyternotebook`, `cursor`.

**Environment variable overrides:**

| Environment Variable | Config Field |
|---------------------|--------------|
| `UF_SANDBOX_IMAGE` | `sandbox.image` |
| `UF_SANDBOX_BACKEND` | `sandbox.backend` |
| `UF_SANDBOX_RUNTIME` | `sandbox.runtime` |
| `UF_SANDBOX_UIDMAP` | `sandbox.uid_map` (set to `1` or `true`) |
| `UF_SANDBOX_IDE` | `sandbox.ide` |

```yaml
sandbox:
  runtime: podman
  backend: podman
  image: quay.io/unbound-force/opencode-dev:latest
  resources:
    memory: 16g
    cpus: "8"
  mode: isolated
  ide: none
  demo_ports:
    - 8080
    - 3000
  uid_map: true
```

---

### gateway

Controls `uf gateway` behavior for the LLM reverse proxy.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `port` | int | `53147` | Port the gateway listens on. Must be 0--65535. |
| `provider` | string | `auto` | LLM provider backend. |

Valid values for `provider`: `auto`, `anthropic`, `vertex`,
`bedrock`.

**Environment variable overrides:**

| Environment Variable | Config Field |
|---------------------|--------------|
| `UF_GATEWAY_PORT` | `gateway.port` |
| `UF_GATEWAY_PROVIDER` | `gateway.provider` |

```yaml
gateway:
  port: 53147
  provider: vertex
```

---

### embedding

Controls the embedding model used by Dewey for semantic search.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `model` | string | `granite-embedding:30m` | Ollama model name for embeddings. |
| `dimensions` | int | `256` | Embedding vector dimensions. Must be non-negative. |
| `provider` | string | `ollama` | Embedding provider. |
| `host` | string | `http://localhost:11434` | Ollama server URL. |

Valid values for `provider`: `ollama`.

**Environment variable overrides:**

| Environment Variable | Config Field |
|---------------------|--------------|
| `OLLAMA_MODEL` | `embedding.model` |
| `OLLAMA_EMBED_DIM` | `embedding.dimensions` |
| `OLLAMA_HOST` | `embedding.host` |

```yaml
embedding:
  model: granite-embedding:30m
  dimensions: 256
  provider: ollama
  host: http://localhost:11434
```

---

### workflow

Controls the hero lifecycle workflow, defining which pipeline
phases are executed by humans and which by the AI swarm.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `execution_modes` | map of string to string | (see below) | Per-phase execution mode. |
| `spec_review` | bool | `false` | Enable spec review before implementation. |

Default `execution_modes`:

| Phase | Default Mode |
|-------|-------------|
| `define` | `human` |
| `implement` | `swarm` |
| `validate` | `swarm` |
| `review` | `swarm` |
| `accept` | `human` |
| `reflect` | `swarm` |

Valid values for each mode: `human`, `swarm`.

```yaml
workflow:
  execution_modes:
    define: human
    implement: swarm
    validate: swarm
    review: swarm
    accept: human
    reflect: swarm
  spec_review: true
```

## Common Scenarios

### Using Vertex AI as LLM provider

Configure the gateway to use Google Vertex AI as the backend
provider:

```yaml
gateway:
  provider: vertex
```

Set the required environment variables for Vertex AI
authentication:

```bash
export VERTEX_LOCATION="us-central1"
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account.json"
```

Alternatively, set the provider via environment variable:

```bash
export UF_GATEWAY_PROVIDER=vertex
```

### Skipping tools during setup

To skip installing specific tools (for example, if you manage
them separately or do not need them):

```yaml
setup:
  skip:
    - ollama
    - dewey
    - replicator
```

Individual tools can also be skipped via the `tools` map by
setting their method to `skip`:

```yaml
setup:
  tools:
    ollama:
      method: skip
    dewey:
      method: skip
```

### Custom sandbox resources

Allocate more memory and CPUs to the containerized sandbox
for resource-intensive workloads:

```yaml
sandbox:
  resources:
    memory: 16g
    cpus: "8"
```

### Custom embedding model

Replace the default `granite-embedding:30m` model with a
different Ollama-compatible embedding model:

```yaml
embedding:
  model: mxbai-embed-large
  dimensions: 1024
```

Or override via environment variables without modifying the
config file:

```bash
export OLLAMA_MODEL=mxbai-embed-large
export OLLAMA_EMBED_DIM=1024
```

### Using DevPod as sandbox backend

DevPod provides persistent cloud development environments
with `.devcontainer/devcontainer.json` support.

```yaml
sandbox:
  backend: devpod
  ide: vscode  # optional: none, vscode, openvscode, fleet, jupyternotebook, cursor
```

Initialize a DevPod workspace definition:

```bash
# Scaffold .devcontainer/devcontainer.json (OS-specific)
uf sandbox init

# Create and enter the workspace
uf sandbox create --backend devpod

# Create with VS Code IDE
uf sandbox create --backend devpod --ide vscode
```

`.devcontainer/` is gitignored because the generated config
is OS-specific (macOS and Linux use different Podman user
namespace flags). Each developer runs `uf sandbox init` to
generate their platform-appropriate config.

Auto-detection prefers DevPod when both the `devpod`
binary and `.devcontainer/devcontainer.json` exist in
the project. Falls back to Podman otherwise.

### User-level defaults

Place shared defaults in the user-level config file so they
apply to all projects. Repo-level settings override these:

```bash
mkdir -p ~/.config/uf
cat > ~/.config/uf/config.yaml << 'EOF'
setup:
  package_manager: dnf
sandbox:
  runtime: podman
EOF
```

Individual projects can then override specific fields in their
`.uf/config.yaml` without repeating the shared defaults.
