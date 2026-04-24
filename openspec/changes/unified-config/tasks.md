## 1. Config Package Foundation

- [x] 1.0 Add `github.com/goccy/go-yaml` to `go.mod` as a direct
  dependency. The `internal/config/` package MUST NOT import
  `gopkg.in/yaml.v3`. Existing files using the archived yaml.v3
  are migrated in a separate follow-up change.
- [x] 1.1 Create `internal/config/config.go` with the `Config`
  struct and all nested structs (`SetupConfig`, `ScaffoldConfig`,
  `EmbeddingConfig`, `SandboxConfig`, `GatewayConfig`,
  `DoctorConfig`, `WorkflowConfig`, `ToolConfig`,
  `ResourcesConfig`, `CheConfig`, `HeroesConfig`). Add YAML
  and JSON struct tags on all fields.
- [x] 1.2 Create `internal/config/defaults.go` with a `defaults()`
  function that returns a `Config` populated with all compiled
  defaults (values currently hardcoded as Go constants across
  `setup.go`, `config.go`, `gateway.go`, `checks.go`).
- [x] 1.3 Create `LoadOptions` struct with injectable fields:
  `ProjectDir string`, `ReadFile func(string) ([]byte, error)`,
  `Getenv func(string) string`,
  `UserConfigDir func() (string, error)`. Add a `defaults()`
  method that populates zero-value fields with production
  implementations (`os.ReadFile`, `os.Getenv`,
  `os.UserConfigDir`).
- [x] 1.4 Implement the `merge(base, overlay Config) Config`
  function. Deep merge: non-zero values from overlay replace
  base. Slice fields (e.g., `Skip`) are replaced, not appended.
  Map fields (e.g., `Tools`) are merged key-by-key.
- [x] 1.5 Implement the `applyEnvOverrides(cfg Config,
  getenv func(string) string) Config` function. Map env vars
  to config fields per design D10 (OLLAMA_MODEL, OLLAMA_HOST,
  UF_SANDBOX_IMAGE, UF_SANDBOX_BACKEND, UF_CHE_URL,
  UF_CHE_TOKEN, UF_PACKAGE_MANAGER, UF_SANDBOX_RUNTIME,
  UF_GATEWAY_PORT, UF_GATEWAY_PROVIDER).
- [x] 1.6 Implement `Load(opts LoadOptions) (*Config, error)`.
  Resolution order: defaults → user config → repo config →
  env overrides. Missing files are not errors (return defaults).
  User config path: `os.UserConfigDir()/uf/config.yaml`.
  Repo config path: `<ProjectDir>/.uf/config.yaml`.
- [x] 1.7 Write tests for `config_test.go`: Load with no files,
  Load with repo only, Load with user only, Load with both
  (repo wins), Load with env overrides, merge deep behavior,
  merge slice replacement, merge map merging, defaults
  correctness. All tests use injected ReadFile/Getenv — no
  real filesystem.

## 2. Config Template and Init Command

- [x] 2.1 Create `internal/config/template.go` with a
  `Template() string` function that returns the full
  commented-out YAML template for the current version.
  All 7 sections, all fields, all with inline comments
  documenting valid values and defaults.
- [x] 2.2 Implement `InitFile(opts InitOptions) (*InitResult,
  error)` in `internal/config/init.go`. When no config file
  exists: write the template. When config file exists: parse
  existing file as `ast.Node` tree (via `goccy/go-yaml`), parse
  template as `ast.Node` tree, walk template to add missing
  sections, remove deprecated sections, preserve uncommented
  user values. Use `YAMLPath` for section-level manipulation.
  Write result atomically (temp file + rename).
- [x] 2.3 Define `InitResult` struct with fields: `Created bool`,
  `Updated bool`, `SectionsAdded []string`,
  `SectionsRemoved []string`, `Path string`.
- [x] 2.4 Write tests for InitFile: create from scratch,
  update adding new section, update removing deprecated section,
  preserve uncommented values, idempotent re-run, atomic write
  (verify temp+rename pattern). Test with injected ReadFile
  and WriteFile.
- [x] 2.5 Create `cmd/unbound-force/config.go` with
  `newConfigCmd() *cobra.Command` returning a command group
  with `init`, `show`, and `validate` subcommands.
- [x] 2.6 Implement `configInitParams` struct and
  `runConfigInit()` function following established params/run
  pattern. Wire `--dir` flag.
- [x] 2.7 Register `newConfigCmd()` in `cmd/unbound-force/main.go`
  via `root.AddCommand(newConfigCmd())`.

## 3. Config Show Command

- [x] 3.1 Implement `configShowParams` struct with `targetDir`,
  `format` (`text`/`json`), `stdout`. Implement
  `runConfigShow()` that calls `config.Load()` and outputs the
  effective config as YAML (text format) or JSON (json format).
- [x] 3.2 Write tests: show with defaults only, show with
  merged config, show JSON format is valid JSON, show YAML
  format is valid YAML.

## 4. Config Validate Command

- [x] 4.1 Create `internal/config/schema.go`. Generate a JSON
  Schema from the `Config` struct using
  `internal/schemas/GenerateSchema()`. Implement
  `Validate(data []byte) error` using
  `internal/schemas/ValidateBytes()`.
- [x] 4.2 Implement a `ValidationReport` struct following the
  doctor `CheckResult`/`CheckGroup` pattern. Implement
  `FormatText()` and `FormatJSON()` formatters.
- [x] 4.3 Implement `configValidateParams` struct and
  `runConfigValidate()` function. Load the config file (raw
  bytes), validate against schema, report results. Missing
  file = pass (exit 0). Invalid file = fail (exit non-zero).
- [x] 4.4 Write tests: valid config passes, invalid field
  value fails, missing file passes, multiple errors all
  reported, JSON and text output formats.

## 5. Setup Integration

- [x] 5.1 Add config-derived fields to `setup.Options`:
  `PackageManager string`, `SkipTools []string`,
  `ToolMethods map[string]ToolConfig`.
- [x] 5.2 Update `cmd/unbound-force/main.go` `runSetup()` to
  load config via `config.Load()` and populate the new Options
  fields from `cfg.Setup`.
- [x] 5.3 Update each `installXxx()` function in `setup.go` to
  check its tool method config: if `method == "skip"` or tool
  is in `SkipTools`, return skipped result. If method is
  specific (`homebrew`, `dnf`, `rpm`, `curl`), use that method.
  If `auto`, use existing auto-detect logic.
- [x] 5.4 Wire `installViaRpm()` into the method dispatch. This
  function exists at `setup.go:711-741` but is currently never
  called. Add it as the handler for `method: rpm` and
  `method: dnf`.
- [x] 5.5 Implement `package_manager: manual` behavior: when
  set, all tools with `method: auto` are skipped.
- [x] 5.6 Replace duplicated `graniteModel` / `graniteEmbedDim`
  constants in `setup.go` with config-derived values from
  `cfg.Embedding`. Pass embedding config through Options.
- [x] 5.7 Write tests: skip tool via config, per-tool method
  override, manual package manager skips auto tools, embedding
  model from config, backward compat (no config = existing
  behavior).

## 6. Scaffold Integration

- [x] 6.1 Remove the `workflowConfigContent` constant and the
  `.uf/config.yaml` creation logic from `initSubTools()` in
  `scaffold.go` (lines 756-814).
- [x] 6.2 Add `Language` config reading: if `cfg.Scaffold.Language`
  is set and `--lang` flag is not provided, use the config
  value instead of auto-detection. Update `runInit()` in
  `main.go` to load config and pass language preference.
- [x] 6.3 Update `printSummary()` to no longer mention
  `.uf/config.yaml` in the sub-tool results.
- [x] 6.4 Write tests: language from config overrides auto-detect,
  `--lang` flag overrides config, no config = existing
  auto-detect behavior, `uf init` no longer creates
  `.uf/config.yaml`.

## 7. Doctor Integration

- [x] 7.1 Add a "Configuration" check group to `doctor.go` that:
  - Reports whether `.uf/config.yaml` exists (info, not
    required).
  - Reports whether `~/.config/uf/config.yaml` exists (info).
  - Validates the config file against JSON Schema if present.
  - Warns if `.uf/sandbox.yaml` exists (deprecated).
- [x] 7.2 Add config-derived fields to `doctor.Options`:
  `SkipChecks []string`,
  `ToolSeverities map[string]string`.
- [x] 7.3 Update `runDoctor()` in `main.go` to load config and
  populate doctor Options from `cfg.Doctor`.
- [x] 7.4 Update tool check functions to respect severity
  overrides from config (`doctor.tools.<name>: optional`
  overrides the hardcoded `recommended`).
- [x] 7.5 Update tool check functions to respect the skip list
  (`doctor.skip: [embedding_capability]` suppresses that check).
- [x] 7.6 Replace `graniteModel` constant in `checks.go` with
  a config-derived value for the embedding model name.
- [x] 7.7 Write tests: config validation check passes/fails,
  sandbox.yaml deprecation warning, skip list suppresses checks,
  severity overrides apply, embedding model from config.

## 8. Sandbox Integration

- [x] 8.1 Update `cmd/unbound-force/sandbox.go` to load unified
  config and populate sandbox Options from `cfg.Sandbox`. Map
  `cfg.Sandbox.Runtime` → `opts.BackendName` (for now, runtime
  and backend are unified; Docker backend is future work).
- [x] 8.2 Update sandbox `LoadConfig()` in `workspace.go`:
  when `sandbox:` section in unified config is non-empty, use
  those values. When empty, fall back to `.uf/sandbox.yaml`.
  Print deprecation warning to stderr when sandbox.yaml found.
- [x] 8.3 Add `runtime` field to sandbox config (alongside
  `backend`). When `runtime: docker` is set but no Docker
  backend exists, return an actionable error message.
- [x] 8.4 Write tests: config from unified file, fallback to
  sandbox.yaml, deprecation warning, runtime field validation,
  backward compat with existing sandbox.yaml.

## 9. Gateway Integration

- [x] 9.1 Update `cmd/unbound-force/gateway.go` to load config
  and populate gateway Options from `cfg.Gateway`. CLI flags
  override config values (only apply config when flag is at
  zero value).
- [x] 9.2 Write tests: port from config, provider from config,
  CLI flag overrides config, no config = existing defaults.

## 10. Agent Model Cleanup [P]

- [x] 10.1 Remove `model: google-vertex-anthropic/claude-opus-4-6@default`
  from all 12 agent files in `.opencode/agents/`:
  `cobalt-crush-dev.md`, `constitution-check.md`,
  `divisor-adversary.md`, `divisor-architect.md`,
  `divisor-curator.md`, `divisor-envoy.md`,
  `divisor-guard.md`, `divisor-herald.md`,
  `divisor-scribe.md`, `divisor-sre.md`,
  `divisor-testing.md`, `mx-f-coach.md`.
- [x] 10.2 Sync all 12 scaffold asset copies in
  `internal/scaffold/assets/opencode/agents/`.
- [x] 10.3 Update `expectedAssetPaths` in scaffold tests if
  any file count changes (unlikely — files remain, only
  frontmatter changes).
- [x] 10.4 Verify agent inheritance works: confirm
  `opencode.json` or `~/.config/opencode/opencode.json`
  has a `model` field, or that OpenCode's default model
  resolution provides a reasonable fallback.

## 11. Documentation [P]

- [x] 11.1 Update AGENTS.md: add `uf config` command group to
  Project Structure, add config file to the structure tree,
  document the layered resolution model, update Recent Changes.
- [x] 11.2 Update `.uf/config.yaml` references throughout
  AGENTS.md (workflow config section is now one of seven).
- [x] 11.3 Add config documentation to the Embedding Model
  Alignment section (config replaces hardcoded constants).

## 12. Build and Verification

- [x] 12.1 Run `make check` (build + test + vet + lint) and
  fix any failures.
- [x] 12.2 Run CI-equivalent checks: read
  `.github/workflows/test.yml` and replicate locally.
- [x] 12.3 Run `uf config validate` on the project's own
  config (if any) to dogfood the validator.
- [x] 12.4 Verify backward compatibility: run `uf doctor`,
  `uf init`, `uf setup --dry-run` without a config file and
  confirm identical behavior to pre-change.

## 13. Constitution Alignment Verification [P]

- [x] 13.1 Verify Autonomous Collaboration: internal packages
  do not depend on config file mechanics (config loading is
  cmd-layer only, values flow through Options structs).
- [x] 13.2 Verify Composability First: all subsystems function
  identically when no config file exists (missing file =
  compiled defaults).
- [x] 13.3 Verify Observable Quality: `uf config show --format
  json` produces valid, machine-parseable JSON. `uf config
  validate` produces structured validation results.
- [x] 13.4 Verify Testability: all config functions are testable
  with injected dependencies (no real filesystem, no real
  env vars). Coverage strategy: unit tests for Load/merge/
  InitFile/Validate, integration-style tests for cmd layer
   wiring. Target ≥80% line coverage for `internal/config/`
   as the initial coverage floor for this new package.
<!-- spec-review: passed -->
<!-- code-review: passed -->
