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

The config file has six top-level sections. Each section is
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

> **Note:** The Podman backend also reads `forwardPorts` from
> `.devcontainer/devcontainer.json` and publishes those ports
> automatically. Ports already listed in `demo_ports` or the
> default server port (4096) are deduplicated.

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

## Review Council Configuration

The review council is configured through four separate `.uf/` files,
not through `.uf/config.yaml`. Each is a closed, schema-validated
contract and is optional -- absent files fall back to compiled
defaults.

| File | Purpose |
|------|---------|
| `.uf/review-matrix.yaml` | Named model profiles and ordered per-mode, per-agent run arrays, plus `limits` (persona/run/timeout/cost). Project-owned: seeded once on first deploy, survives plain `uf init` re-runs, reset to canonical only by `--force`. |
| `.uf/review-matrix.override.yaml` | Per-developer model/variant overrides for specific profile tiers. Only `profiles.<tier>.[model, variant]` paths are permitted (whitelist). Gitignored — each developer maintains their own; never seeded or reset by `uf init`. |
| `.uf/reviewer-capabilities.yaml` | Classifies each Divisor persona as review or content and declares its ordered review scopes. Project-owned: seeded once on first deploy, survives plain `uf init` re-runs, reset to canonical only by `--force`. |
| `.uf/sibling-repos.yaml` | Declares the sibling repositories eligible for bounded evidence acquisition. |

`.uf/review-matrix.yaml` and `.uf/reviewer-capabilities.yaml` are the
review-config files and share identical scaffold ownership semantics:
each is a project-owned extension point, seeded once on first deploy
from the embedded canonical asset, skipped on plain `uf init` re-runs
(locally added `reviewers:` entries and matrix customizations survive
byte-for-byte), and reset to the canonical asset only by
`uf init --force` -- a destructive reset for both files that discards
local entries. `.uf/review-matrix.override.yaml` is deliberately
different: it is gitignored, per-developer, and never seeded or reset
by `uf init`.

Each file has a corresponding JSON Schema under `schemas/` (and a
scaffold mirror under `internal/scaffold/assets/uf/schemas/`), so invalid
values fail closed at load time rather than silently misrouting review
fan-out.

---

## Delegation Model Matrix

Content-agent delegation -- the `uf_delegate_command` tool behind
`/uf.unleash` and `/uf.run-pipeline-plan` -- selects each child
session's model through a capability-driven matrix. The pattern
mirrors the review council: a tracked file carries structure only,
and a gitignored per-developer override carries model identities.

| File | Purpose |
|------|---------|
| `.uf/model-matrix.yaml` | Capability-named profiles, the `packs` mapping (UF convention pack to capability), task-kind `defaults`, per-agent `agents` pins, and `classifier` structure. Model strings are optional everywhere -- a profile without a model falls back to the host model. |
| `.uf/model-matrix.override.yaml` | Per-developer model identities and classifier LLM-engine values. Only `profiles.*.model`, `profiles.*.variant`, and `classifier.llm.*` paths are permitted (fail-closed whitelist). Gitignored -- each developer maintains their own. |

Both files are optional. With no matrix present, delegation behaves
exactly as before: the child session runs on the host model with no
errors. A JSON Schema lives at `schemas/model-matrix/v1.schema.json`
(with mirrors under `.uf/schemas/model-matrix/` and the scaffold
assets), so invalid values fail closed at load time rather than
silently misrouting delegation.

### Tracked matrix

```yaml
version: 1
profiles:            # capability-named model slots; model optional
  spec-authoring: {} #   -- host fallback until the override fills it
  code-writing: {}
  content-writing: {}
  mechanical: {}     # explicit-profile-only in the starter set
packs:               # convention pack name -> capability
  meta-prompt: spec-authoring
  go: code-writing
  typescript: code-writing
  python: code-writing
  content: content-writing
defaults:            # closed task-kind key set
  plan: spec-authoring
  tasks: spec-authoring
  implement: code-writing
agents: {}           # optional per-agent capability pins
classifier: {}       # structure only -- LLM values live in the override
```

Rules enforced at load time:

- `profiles` may declare 1--32 capabilities; each entry MAY carry
  `model` (a `provider/model` identity) and `variant`.
- Every `packs` key MUST match a pack file under
  `.opencode/uf/packs/` (pack-existence check).
- Every `packs`, `defaults`, and `agents` value MUST name a declared
  capability (cross-reference check).
- `agents` pins reject `divisor-*` names -- reviewer-agent selection
  is governed by the review matrix, not this one.
- The tracked file MUST NOT populate `classifier.llm` values
  (endpoint, model, apiKey, request timeout); populated values fail
  the load with the override file named as the remediation.

### Selection precedence

`uf_delegate_command` accepts optional `model` (+`variant`),
`profile`, and `pack` parameters. The child model resolves through
a fixed chain, first match wins:

1. explicit `model` parameter
2. explicit `profile` (capability) parameter
3. explicit `pack` parameter (validated: the pack file exists AND
   has a matrix entry; the pack also renders a tool-generated
   directive instructing the child to load the pack before writing)
4. `agents:` pin for the delegated agent
5. static `defaults:` by task kind (`plan`/`tasks` map to
   spec-authoring)
6. classifier cascade for `implement` and unknown task kinds
   (agent engine, then LLM engine, then the static default)
7. host model (the resolved profile carries no model, or no matrix
   is configured)

An unknown explicit profile, an invalid explicit pack, or an
explicit profile/pack with no matrix configured fails closed with a
non-retryable error naming the remediation. When any matrix profile
model resolves, host-model introspection is skipped entirely --
delegation no longer depends on `HOST_MODEL_UNAVAILABLE`.

### Classifier

For `implement` and unknown task kinds, selection runs a two-engine
cascade:

- **Agent engine** -- the hidden `model-matrix-classifier` subagent
  investigates the feature directory's spec/plan/tasks and relevant
  code, then responds with only a JSON object:
  `{"pack": "<name>", "confidence": <0.0-1.0>}`. Bounded by
  `classifier.agent_timeout_ms` (default 45,000).
- **LLM engine** -- a chat-completions POST (`temperature: 0`, JSON
  response format) against an endpoint configured only in the
  gitignored override. The API key is interpolated from `{env:VAR}`
  -- never a literal key. Bounded by
  `classifier.llm.request_timeout_ms` (default 10,000).

The first valid result whose confidence meets the threshold
(`classifier.confidence_threshold`, default 0.6) and whose pack is
matrix-valid wins. Below-threshold results, timeouts, malformed
responses, and unconfigured engines all fall back to the static
default. Classifier failure never blocks delegation. Results are
cached in-process, keyed by the SHA-256 of the feature directory,
prompt, and agent.

### Override file

Per-developer identities arrive via
`.uf/model-matrix.override.yaml` (gitignored; `uf init` scaffolds
the gitignore entry):

```yaml
profiles:
  spec-authoring:
    model: provider/model-id
  code-writing:
    model: provider/other-model-id
    variant: fast
classifier:
  llm:
    endpoint: https://llm.example.com/v1/chat/completions
    model: classifier-model-id
    apiKey: "{env:MY_API_KEY}"
```

The whitelist is fail-closed: any key outside `profiles.*.model`,
`profiles.*.variant`, and `classifier.llm.*` -- including structural
keys such as `packs`, `defaults`, `agents`, or `classifier.enabled`
-- fails validation with an error naming the whitelist. The
override merges into the tracked matrix by nested path: absent keys
fall through, only already-declared capabilities may be overridden,
and the merged result re-validates against the identical schema.

### Audit trail

Every delegation records `selection_source` (one of
`explicit-model`, `explicit-profile`, `explicit-pack`, `agent-pin`,
`default-kind`, `classifier-agent`, `classifier-llm`, `host`), the
selected pack, the resolved capability and model, and classifier
confidence when applicable -- in the success envelope and the
checkpoint data, so model selection is auditable after the fact.

---

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
