## ADDED Requirements

### Requirement: Config Package

The system MUST provide an `internal/config/` package that
defines a `Config` struct covering seven sections: `setup`,
`scaffold`, `embedding`, `sandbox`, `gateway`, `doctor`, and
`workflow`.

The package MUST provide a `Load(opts LoadOptions) (*Config,
error)` function that implements 4-layer resolution: compiled
defaults, user config, repo config, and environment variable
overrides. CLI flag overrides are applied by the caller at
the cmd layer, making the full system resolution 5 layers
(per design D3).

The package MUST provide injectable function fields on the
`LoadOptions` struct (`ReadFile`, `Getenv`, `UserConfigDir`)
following the established pattern in `internal/setup/`,
`internal/sandbox/`, etc.

#### Scenario: Load with no config files

- **GIVEN** neither `.uf/config.yaml` nor
  `~/.config/uf/config.yaml` exist
- **WHEN** `config.Load()` is called
- **THEN** the returned Config contains all compiled defaults
  and no error is returned

#### Scenario: Load with repo config only

- **GIVEN** `.uf/config.yaml` exists with
  `sandbox: { runtime: podman }`
- **AND** no user config exists
- **WHEN** `config.Load()` is called
- **THEN** `cfg.Sandbox.Runtime` equals `"podman"`
- **AND** all other fields equal compiled defaults

#### Scenario: Load with both configs — repo wins

- **GIVEN** user config has `sandbox: { runtime: docker }`
- **AND** repo config has `sandbox: { runtime: podman }`
- **WHEN** `config.Load()` is called
- **THEN** `cfg.Sandbox.Runtime` equals `"podman"`
  (repo overrides user)

#### Scenario: Load with env var override

- **GIVEN** repo config has `sandbox: { image: "myimage:v1" }`
- **AND** `UF_SANDBOX_IMAGE=override:v2` is set
- **WHEN** `config.Load()` is called
- **THEN** `cfg.Sandbox.Image` equals `"override:v2"`
  (env var overrides config file)

#### Scenario: Merge preserves non-overlapping fields

- **GIVEN** user config has
  `setup: { package_manager: dnf }`
- **AND** repo config has
  `sandbox: { runtime: podman }`
- **WHEN** `config.Load()` is called
- **THEN** `cfg.Setup.PackageManager` equals `"dnf"`
- **AND** `cfg.Sandbox.Runtime` equals `"podman"`
  (both values preserved — deep merge)

---

### Requirement: Config Init Command

The system MUST provide a `uf config init` command that
creates `.uf/config.yaml` with all sections and fields
present as YAML comments.

When a `.uf/config.yaml` already exists, the command SHOULD
create a backup at `.uf/config.yaml.bak` before modifying
the file.

When a `.uf/config.yaml` already exists, the command MUST:
- Preserve all uncommented (user-set) values
- Add sections present in the current-version template but
  absent from the existing file
- Remove sections present in the existing file but absent
  from the current-version template (deprecated sections)
- Produce a file that passes `uf config validate`

The command MUST NOT overwrite uncommented values without
explicit user confirmation.

Files created by `uf config init` MUST be written with 0o644
permissions. Sensitive credentials (tokens, API keys) SHOULD
be provided via environment variables rather than persisted in
config files.

#### Scenario: First-time config init

- **GIVEN** `.uf/config.yaml` does not exist
- **WHEN** `uf config init` is run
- **THEN** `.uf/config.yaml` is created with all 7 sections
- **AND** every field is commented out
- **AND** the file passes `uf config validate`

#### Scenario: Re-init adds new section

- **GIVEN** `.uf/config.yaml` exists with only `workflow:`
  section and `sandbox: { runtime: docker }` uncommented
- **AND** the current UF version has a new `gateway:` section
- **WHEN** `uf config init` is run
- **THEN** `sandbox.runtime` remains `"docker"` (preserved)
- **AND** the `gateway:` section appears as commented-out YAML
- **AND** the file passes `uf config validate`

#### Scenario: Re-init removes deprecated section

- **GIVEN** `.uf/config.yaml` has a `deprecated_section:` that
  does not exist in the current template
- **WHEN** `uf config init` is run
- **THEN** `deprecated_section:` is removed from the file
- **AND** a message is printed listing removed sections

---

### Requirement: Config Show Command

The system MUST provide a `uf config show` command that
displays the effective configuration after all layers are
merged.

The command MUST support `--format text` (default, YAML
output) and `--format json` (JSON output).

#### Scenario: Show effective config

- **GIVEN** user config sets `setup.package_manager: dnf`
- **AND** repo config sets `sandbox.runtime: podman`
- **AND** `UF_GATEWAY_PORT=9999` is set
- **WHEN** `uf config show` is run
- **THEN** output shows `package_manager: dnf`,
  `runtime: podman`, `port: 9999`
- **AND** all other fields show compiled defaults

#### Scenario: Show with JSON format

- **GIVEN** any configuration state
- **WHEN** `uf config show --format json` is run
- **THEN** output is valid JSON matching the Config struct

---

### Requirement: Config Validate Command

The system MUST provide a `uf config validate` command that
validates `.uf/config.yaml` against a JSON Schema.

The command MUST report all validation errors, not just the
first one.

The command MUST produce output following the established
doctor CheckResult/CheckGroup pattern (structured results
with severity, message, and detail).

The command MUST support `--format text` (default) and
`--format json`.

#### Scenario: Valid config

- **GIVEN** `.uf/config.yaml` exists and is schema-valid
- **WHEN** `uf config validate` is run
- **THEN** all checks show pass severity
- **AND** exit code is 0

#### Scenario: Invalid field value

- **GIVEN** `.uf/config.yaml` has
  `sandbox: { resources: { memory: "lots" } }`
- **WHEN** `uf config validate` is run
- **THEN** a fail-severity result reports the invalid value
- **AND** exit code is non-zero

#### Scenario: No config file

- **GIVEN** `.uf/config.yaml` does not exist
- **WHEN** `uf config validate` is run
- **THEN** a message reports no config file found
- **AND** exit code is 0 (missing file is valid —
  defaults are used)

---

### Requirement: Setup Config Integration

`uf setup` MUST read the `setup` section from the unified
config and respect the following fields:

- `package_manager`: Determines the default install method
  when a tool's `method` is `auto`. Valid values: `auto`,
  `homebrew`, `dnf`, `apt`, `manual`.
- `skip`: A list of tool names to skip during setup. Skipped
  tools produce a `"skipped"` result with detail
  `"excluded by config"`.
- `tools.<name>.method`: Per-tool install method override.
  Valid values: `auto`, `homebrew`, `dnf`, `rpm`, `apt`,
  `curl`, `skip`. Overrides the global `package_manager`
  for this specific tool.
- `tools.<name>.version`: Target version for tools that
  support version selection (e.g., `node`).

When `package_manager` is `manual`, all tools with
`method: auto` MUST be skipped (the CI pattern — tools
are pre-installed in the environment).

#### Scenario: Skip tool via config

- **GIVEN** config has `setup: { skip: [ollama] }`
- **WHEN** `uf setup` runs
- **THEN** Ollama installation is skipped
- **AND** result shows `"skipped"` with detail
  `"excluded by config"`

#### Scenario: Per-tool method override

- **GIVEN** config has
  `setup: { tools: { gaze: { method: rpm } } }`
- **WHEN** `uf setup` runs and Gaze is not installed
- **THEN** Gaze is installed via RPM (using the existing
  `installViaRpm()` function)

#### Scenario: Manual package manager skips auto tools

- **GIVEN** config has
  `setup: { package_manager: manual }`
- **AND** no per-tool method overrides
- **WHEN** `uf setup` runs
- **THEN** all tools with method `auto` are skipped
- **AND** tools with explicit methods (e.g., `method: curl`)
  still execute

---

### Requirement: Scaffold Config Integration

`uf init` MUST read the `scaffold` section from the unified
config.

- `scaffold.language`: Overrides language auto-detection. If
  set, the scaffold engine uses this value instead of probing
  for `go.mod`, `tsconfig.json`, etc. The `--lang` CLI flag
  overrides the config value.

`uf init` MUST NOT create `.uf/config.yaml`. The file is
created exclusively by `uf config init`.

#### Scenario: Language from config

- **GIVEN** config has `scaffold: { language: typescript }`
- **AND** no `--lang` flag is provided
- **WHEN** `uf init` runs
- **THEN** TypeScript convention packs are deployed
- **AND** Go packs are not deployed

---

### Requirement: Embedding Config Integration

`uf setup` and `uf doctor` MUST read the `embedding` section
from the unified config for the embedding model name, dimension,
and Ollama host, instead of using hardcoded Go constants.

The duplicated `graniteModel` constants in `setup.go` and
`checks.go` MUST be replaced by config-derived values.

#### Scenario: Custom embedding model

- **GIVEN** config has
  `embedding: { model: "mxbai-embed-large", dimensions: 1024 }`
- **WHEN** `uf setup` runs
- **THEN** `ollama pull mxbai-embed-large` is executed
- **AND** `OLLAMA_MODEL` is set to `"mxbai-embed-large"`
- **AND** `OLLAMA_EMBED_DIM` is set to `"1024"`

---

### Requirement: Sandbox Config Absorption

The `sandbox` section of the unified config MUST absorb all
fields from the existing `.uf/sandbox.yaml`: `che` (url,
token), `ollama` (host), `backend`, and `demo_ports`.

When the `sandbox` section in `.uf/config.yaml` is empty
(all zero values), the system MUST fall back to reading
`.uf/sandbox.yaml` for backward compatibility.

When `.uf/sandbox.yaml` exists, the system MUST print a
deprecation warning to stderr on each sandbox operation.

#### Scenario: Sandbox config from unified file

- **GIVEN** `.uf/config.yaml` has
  `sandbox: { backend: che, che: { url: "https://che.example" } }`
- **AND** `.uf/sandbox.yaml` does not exist
- **WHEN** `uf sandbox start` runs
- **THEN** the Che backend is used with the configured URL

#### Scenario: Fallback to sandbox.yaml

- **GIVEN** `.uf/config.yaml` has no `sandbox:` section
- **AND** `.uf/sandbox.yaml` exists with `backend: che`
- **WHEN** `uf sandbox start` runs
- **THEN** the Che backend is used (fallback)
- **AND** a deprecation warning is printed to stderr

---

### Requirement: Gateway Config Integration

`uf gateway` MUST read the `gateway` section from the unified
config for default port and provider.

CLI flags (`--port`, `--provider`) MUST override config values.

#### Scenario: Gateway port from config

- **GIVEN** config has `gateway: { port: 9999 }`
- **AND** no `--port` flag is provided
- **WHEN** `uf gateway start` runs
- **THEN** the gateway listens on port 9999

---

### Requirement: Doctor Config Integration

`uf doctor` MUST read the `doctor` section from the unified
config.

- `doctor.skip`: A list of check names to skip. Skipped checks
  MUST NOT appear in the output.
- `doctor.tools.<name>`: Override the severity of a tool check.
  Valid values: `required`, `recommended`, `optional`.

The system MUST add a "Configuration" check group to the doctor
report that validates the config file (if present) against the
JSON Schema.

#### Scenario: Skip doctor check via config

- **GIVEN** config has
  `doctor: { skip: [embedding_capability] }`
- **WHEN** `uf doctor` runs
- **THEN** the embedding capability check is not executed
- **AND** it does not appear in the report

#### Scenario: Override tool severity

- **GIVEN** config has `doctor: { tools: { gaze: optional } }`
- **WHEN** `uf doctor` runs and Gaze is not installed
- **THEN** the Gaze check shows as info/optional
  (not warn/recommended)

---

### Requirement: Agent Model Cleanup

The `model:` field MUST be removed from all agent frontmatter
files where it is set to the same value as the project default.

This affects 12 files in `.opencode/agents/` and their
corresponding scaffold copies in `internal/scaffold/assets/`.

The scaffold engine MUST NOT inject a `model:` field into
agent files during deployment.

#### Scenario: Agent inherits model from OpenCode config

- **GIVEN** `model:` is absent from `divisor-guard.md` frontmatter
- **AND** `opencode.json` has `"model": "anthropic/claude-sonnet-4-5"`
- **WHEN** the divisor-guard agent is invoked
- **THEN** it uses `anthropic/claude-sonnet-4-5`
  (inherited from OpenCode config)

---

## MODIFIED Requirements

### Requirement: uf init no longer creates config file

Previously: `uf init` created `.uf/config.yaml` with a
commented-out workflow section template via the
`workflowConfigContent` constant in `scaffold.go`.

Now: `uf init` MUST NOT create `.uf/config.yaml`. The
`workflowConfigContent` constant MUST be removed from
`scaffold.go`. The config file is created exclusively by
`uf config init`.

Existing `.uf/config.yaml` files created by prior versions
of `uf init` MUST remain valid and MUST continue to be
read by the workflow engine.

---

## REMOVED Requirements

### Requirement: `.uf/sandbox.yaml` as primary config

`.uf/sandbox.yaml` is deprecated as the primary sandbox
configuration source. It is absorbed into the `sandbox:`
section of `.uf/config.yaml`.

The file continues to be read as a fallback (see backward
compatibility scenario above) but users SHOULD migrate to
the unified config via `uf config init`.
