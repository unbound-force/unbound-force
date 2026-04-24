## Context

The Unbound Force CLI has 6 separate configuration surfaces and
40+ hardcoded defaults spread across Go constants in
`internal/setup/`, `internal/doctor/`, `internal/sandbox/`, and
`internal/gateway/`. Users cannot customize package managers,
container runtimes, embedding models, sandbox resources, or
doctor check severities without code changes. The existing
`.uf/sandbox.yaml` is the only subsystem that loads a config
file; all others rely exclusively on CLI flags and environment
variables.

The proposal establishes a unified `.uf/config.yaml` with layered
resolution, a new `internal/config/` package, and a `uf config`
command group. This design documents the technical architecture.

## Goals / Non-Goals

### Goals

- Define the Config struct and layered Load() resolution
- Define the `uf config init` update algorithm
  (comment-preserving merge)
- Define the env var overlay mapping
- Define the JSON Schema validation strategy
- Define the sandbox.yaml migration path
- Define how config flows into each subsystem's Options struct
- Maintain full backward compatibility (missing file = today's
  behavior)

### Non-Goals

- LLM model configuration: handled by OpenCode's own config
  hierarchy. Agent `model:` frontmatter removal is a parallel
  task with no config package dependency.
- New package manager backends (apt, winget): wiring
  `installViaRpm()` is in scope; writing new `installViaApt()`
  is future work. The config schema supports `apt` as a value
  but the implementation can return "not yet supported."
- Docker container runtime backend: the config schema supports
  `sandbox.runtime: docker` but implementing a `DockerBackend`
  is future work. The config enables it without requiring it.
- Windows support: `setup.go` blocks Windows at line 135. The
  config cannot override this platform guard. Windows support
  is a separate effort.
- Remote embedding providers: `embedding.provider` supports
  only `ollama` today. The field exists for future extensibility.

## Decisions

### D1: Config file location — `.uf/config.yaml`

Reuse the existing path rather than introducing a new file name.
The scaffold already creates this file (for workflow config).
By expanding it with sibling sections, we avoid adding yet another
config surface. The file is user-owned and never overwritten by
`uf init`.

**Rationale**: Reduces config surface count from 6 to 5
(absorbing sandbox.yaml). Using the existing path avoids a
rename migration.

### D2: No config by default — `uf config init` creates it

`uf init` will no longer create `.uf/config.yaml`. The file is
only created by `uf config init`. This means first-time users
never encounter config complexity unless they need it.

**Rationale**: Convention over configuration
(constitution II — Composability First). The system works
standalone with zero config files. Config is opt-in.

### D3: Layered resolution — 5 layers

```text
Priority (highest wins):
  1. CLI flags           --port 8080
  2. Environment vars    UF_SANDBOX_IMAGE
  3. Repo config         .uf/config.yaml
  4. User config         ~/.config/uf/config.yaml
  5. Compiled defaults   Go constants
```

Repo config overrides user config. This matches the team-decides
model: if the project says `sandbox.runtime: podman`, a
developer's personal `sandbox.runtime: docker` does not override
it. CLI flags always win for one-off overrides.

**Rationale**: Project decisions (committed to repo) should be
authoritative. Personal preferences fill gaps. This matches
the sandbox `ResolveBackend()` pattern (flag > env > config >
auto-detect) already established in `backend.go:69-123`.

### D4: User config path — `os.UserConfigDir()`

Use Go's `os.UserConfigDir()` to resolve the user config path:
- Linux: `~/.config/uf/config.yaml`
- macOS: `~/Library/Application Support/uf/config.yaml`

This follows XDG conventions on Linux and Apple conventions on
macOS. The `UF_CONFIG_HOME` environment variable MAY override
this path for non-standard setups.

**Rationale**: Go stdlib handles platform differences. No manual
path construction needed.

### D5: Deep merge with zero-value semantics

The merge function overlays non-zero values from higher-priority
configs onto lower-priority ones. A commented-out field (absent
from YAML) produces a zero value and does not override. An
explicitly set field (uncommented) overrides.

```text
User config:       { setup: { package_manager: "dnf" } }
Repo config:       { sandbox: { runtime: "podman" } }
Merged result:     { setup: { package_manager: "dnf" },
                     sandbox: { runtime: "podman" } }
```

For slice fields (like `setup.skip`), repo config replaces
(not appends to) user config. This prevents confusing additive
behavior where a skip list grows unexpectedly.

**Rationale**: Slice-replace is simpler to reason about and
matches how YAML deserialization naturally works. Users who need
additive skip lists can use `uf setup --skip tool1 --skip tool2`
as CLI flag overrides.

### D6: `uf config init` update algorithm

When re-running `uf config init` on an existing file:

1. Parse the existing file into an `ast.Node` tree (preserves
   comments and formatting).
2. Generate the current-version template as an `ast.Node` tree.
3. Walk the template tree:
   - For each section present in the template but absent from
     the existing file: append the section (commented out).
   - For each section present in the existing file but absent
     from the template: remove it (deprecated section).
   - For each section present in both: preserve the existing
     content (user values win).
4. Write the merged tree back.

**Rationale**: `github.com/goccy/go-yaml` provides `ast.Node`
and `CommentMap` for full comment-preserving round-trip editing.
The `YAMLPath` API simplifies section-level manipulation
(add/remove/replace by path). This enables surgical updates
without destroying user formatting. The algorithm ensures the
file is always schema-valid after the command.

Note: the new `internal/config/` package uses `goccy/go-yaml`
exclusively. Existing files importing the archived
`gopkg.in/yaml.v3` are migrated in a separate follow-up change.

### D7: JSON Schema for validation

Define a JSON Schema (draft 2020-12) for the config file. The
schema is generated from the Go `Config` struct using the
existing `internal/schemas/GenerateSchema()` function (which
uses `invopop/jsonschema`). Validation uses the existing
`internal/schemas/ValidateBytes()` function.

This dogfoods the project's own schema infrastructure.

**Rationale**: Constitution III (Observable Quality) — the config
format is machine-parseable and self-validating. Reusing existing
infrastructure avoids new dependencies.

### D8: Config loading at the cmd layer

Config is loaded once in the cmd layer (in each `runXxx()`
function) and distributed to subsystem Options structs. Internal
packages never load config themselves — they receive values
through their Options fields.

```text
cmd layer:     cfg := config.Load(...)
               opts.PackageManager = cfg.Setup.PackageManager

internal pkg:  func Run(opts Options) { /* uses opts fields */ }
```

New configurable fields are added to each package's Options
struct. The injectable function pattern (`ReadFile`, `Getenv`,
`LookPath`) is preserved — config loading uses the same
injection points.

**Rationale**: Constitution I (Autonomous Collaboration) —
internal packages do not depend on config file mechanics. They
receive plain values through their interfaces. This also
preserves the testability pattern: tests construct Options
directly without needing config files.

### D9: Sandbox.yaml backward compatibility

Phase 1 (this change): `config.Load()` reads the `sandbox:`
section from `.uf/config.yaml`. If the section is empty (all
zero values), it falls back to reading `.uf/sandbox.yaml` via
the existing `LoadConfig()` function. If sandbox.yaml exists,
a deprecation warning is printed to stderr.

Phase 2 (future `uf config init` enhancement): `uf config init`
detects `.uf/sandbox.yaml`, migrates its values into
`.uf/config.yaml`, and renames sandbox.yaml to
`.uf/sandbox.yaml.bak`.

Phase 3 (future major version): sandbox.yaml fallback removed.

**Rationale**: Backward compatibility first. Users with existing
sandbox.yaml files keep working. The deprecation warning guides
them toward migration.

### D10: Environment variable mapping

Existing env vars keep their names for backward compatibility.
New env vars follow the `UF_` prefix convention. The mapping:

| Config path               | Env var (existing or new)    |
|---------------------------|------------------------------|
| `setup.package_manager`   | `UF_PACKAGE_MANAGER` (new)   |
| `embedding.model`         | `OLLAMA_MODEL` (existing)    |
| `embedding.dimensions`    | `OLLAMA_EMBED_DIM` (existing)|
| `embedding.host`          | `OLLAMA_HOST` (existing)     |
| `sandbox.backend`         | `UF_SANDBOX_BACKEND` (exist) |
| `sandbox.image`           | `UF_SANDBOX_IMAGE` (existing)|
| `sandbox.runtime`         | `UF_SANDBOX_RUNTIME` (new)   |
| `sandbox.che.url`         | `UF_CHE_URL` (existing)      |
| `sandbox.che.token`       | `UF_CHE_TOKEN` (existing)    |
| `gateway.port`            | `UF_GATEWAY_PORT` (new)      |
| `gateway.provider`        | `UF_GATEWAY_PROVIDER` (new)  |

Not all config fields need env var overrides. Fields like
`scaffold.language`, `doctor.skip`, and `setup.tools.*` are
complex enough that env var representation would be awkward.
These are config-file-only settings.

**Sensitive credentials** (e.g., `UF_CHE_TOKEN`) SHOULD be
provided exclusively via environment variables and MUST NOT
be persisted in config files. `uf config show` MUST redact
any field marked as sensitive.

### D11: Agent model frontmatter cleanup

Remove `model: google-vertex-anthropic/claude-opus-4-6@default`
from all 12 agent `.md` files. OpenCode's native config hierarchy
handles model selection:
- Subagents inherit from their invoking primary agent.
- Primary agents inherit from `opencode.json`'s `model` field.
- Users override at `~/.config/opencode/opencode.json`.

This is parallel work with no dependency on the config package.

**Rationale**: The model choice is an OpenCode domain concern,
not a UF infrastructure concern. Centralizing it in OpenCode's
config avoids duplication.

## Risks / Trade-offs

### R1: YAML comment-preserving merge is complex

The `uf config init` update algorithm (D6) requires AST-level
YAML manipulation. `goccy/go-yaml`'s `ast.Node` API supports this
but the code will be non-trivial (~150-200 lines). Incorrect
implementation could corrupt user config files.

**Mitigation**: Thorough test coverage with edge cases
(empty file, fully commented file, file with custom comments,
deprecated sections, new sections). Write-to-temp-then-rename
for atomic file updates.

### R2: Config scope creep

The config file could grow to absorb every possible setting,
becoming an unwieldy configuration language. Each new feature
might add sections.

**Mitigation**: Strict criteria for what enters the config: only
settings that are genuinely opinionated (user might want
something different) and currently require code changes to
override. Settings already configurable via CLI flags or env vars
don't need config entries unless persistence across runs is
valuable.

### R3: Merge precedence confusion

Five layers of config resolution can make it hard to understand
"why is this value X?" when debugging.

**Mitigation**: `uf config show` displays the effective config.
Future enhancement: `uf config show --verbose` could annotate
each value with its source (which layer it came from).

### R4: Backward compatibility with existing `.uf/config.yaml`

Some users may have manually created `.uf/config.yaml` with the
workflow section (the template that `uf init` used to scaffold).
The new expanded schema must not break these files.

**Mitigation**: The workflow section schema is unchanged. The
expanded file simply adds sibling sections. Existing files with
only a `workflow:` section remain valid. `uf config validate`
would pass them without errors.
