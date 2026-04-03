# Research: Doctor & Setup Dewey Alignment

**Branch**: `023-doctor-setup-dewey` | **Date**: 2026-04-03

## R1: Dewey Embedding Capability Verification

**Question**: How can `uf doctor` verify that Dewey can
actually generate embeddings end-to-end?

**Findings**:

1. **Dewey's own doctor**: Dewey exposes `dewey doctor`
   which checks the full embedding pipeline: Ollama
   running, model loaded, embeddings in database. Output
   format is human-readable text with emoji indicators.

2. **Ollama HTTP API**: The Ollama embedding endpoint
   (`POST http://localhost:11434/api/embed`) accepts a
   model name and input text, returning an embedding
   vector. A successful response with a non-empty
   `embeddings` array confirms the full pipeline works.

3. **`ollama list` check**: The existing `checkDewey()`
   function already runs `ollama list` to verify the
   model is pulled. This is necessary but not sufficient
   — the model can be pulled but Ollama not serving.

**Decision**: Use the Ollama HTTP API directly via
`http://localhost:11434/api/embed` with a minimal test
input. This is the most reliable end-to-end check:
- It verifies Ollama is running (connection succeeds)
- It verifies the model is loaded (model name accepted)
- It verifies embeddings work (response contains vectors)

The alternative of shelling out to `dewey doctor` was
rejected because:
- It produces human-readable output that would need
  parsing (fragile)
- It checks many things beyond embeddings (vault,
  workspace, MCP server) that are out of scope for this
  check
- It requires Dewey to be installed, but we want to
  check embedding capability even when Dewey is absent
  (the Ollama+model check is still valuable standalone)

**Implementation approach**: Add a new `checkEmbeddingCapability()`
function in `checks.go` that:
1. Checks if Ollama is reachable at `localhost:11434`
2. Sends a minimal embedding request for `granite-embedding:30m`
3. Verifies the response contains a non-empty embeddings array
4. Reports PASS/FAIL with actionable hints

The Ollama endpoint URL should be injectable via
`Options` for testability (Constitution Principle IV).

## R2: Ollama Check Demotion Semantics

**Question**: How should the Ollama serving check be
demoted to informational status?

**Findings**:

1. **Current behavior**: The `ollama` binary is checked
   in `coreToolSpecs` as an optional tool (neither
   `required` nor `recommended`). When found, it gets
   `Pass` severity. When not found, it gets `Pass`
   severity with "not found" message. The post-check
   `checkOllamaModel()` enriches the result with model
   status.

2. **Existing severity model**: The doctor uses three
   severities: `Pass`, `Warn`, `Fail`. There is no
   explicit `Info` severity. However, `Pass` with an
   `InstallHint` renders as `[INFO]` in plain text and
   `⊘` (gray) in color mode (see `formatIndicator()`).

3. **Dewey manages Ollama**: Per Spec 021 FR-021 and
   the spec's US-3, Dewey manages Ollama's lifecycle.
   The direct Ollama serving check can produce false
   negatives (Ollama not running because Dewey hasn't
   started it yet).

**Decision**: Demote the Ollama serving check by:
1. Keeping the `ollama` binary check in `coreToolSpecs`
   as-is (it's already optional/informational)
2. In the Dewey health check group, when the new
   embedding capability check is present, annotate the
   existing embedding model check with a note that
   "Dewey manages Ollama lifecycle"
3. The embedding capability check subsumes the need for
   a direct Ollama serving check — if embeddings work,
   Ollama is necessarily serving

This avoids adding a new severity type and keeps the
change minimal. The key insight is that the embedding
capability check (R1) already covers the Ollama serving
case — it's a superset check.

## R3: Swarm Plugin Fork Source

**Question**: What is the correct npm package name and
install source for the forked Swarm plugin?

**Findings**:

1. **Current install commands**: `installSwarmPlugin()`
   uses `opencode-swarm-plugin@latest` as the package
   name, installed via `bun add -g` or `npm install -g`.

2. **Fork repository**: The fork is at
   `unbound-force/swarm-tools` on GitHub. The npm
   package name for GitHub-hosted packages can be
   installed via:
   - `npm install -g github:unbound-force/swarm-tools`
   - `bun add -g github:unbound-force/swarm-tools`

3. **Package name assumption**: The spec states "The
   fork uses the same package name structure as the
   upstream." This means the npm package name is still
   `opencode-swarm-plugin` but installed from the
   GitHub repo URL rather than the npm registry.

**Decision**: Change the install source from the npm
registry package to the GitHub repository:
- `bun add -g github:unbound-force/swarm-tools`
- `npm install -g github:unbound-force/swarm-tools`

This installs from the fork's GitHub repo directly,
ensuring the forked version is used. The `swarm` binary
name remains the same (the fork doesn't rename it).

For the replacement case (US-2 acceptance scenario 2),
the existing `installSwarmPlugin()` already checks
`LookPath("swarm")` first. If the upstream is installed,
it returns "already installed." To handle replacement,
we need to:
1. Check if swarm is installed
2. If installed, check if it's from the fork (by
   checking the package source or version)
3. If from upstream, uninstall and reinstall from fork

However, detecting the source of an installed npm global
package is complex and fragile. A simpler approach: always
install from the fork URL. npm/bun will replace the
existing package if the name matches. The `--force` flag
or simply re-running the install command will update the
source.

**Revised decision**: Remove the early-return "already
installed" check for the swarm binary. Instead:
1. Always attempt to install from the fork
2. npm/bun `install -g` is idempotent — it updates if
   already installed
3. This ensures the fork version is always current

This is simpler and more reliable than source detection.

## R4: Ollama API Endpoint Configurability

**Question**: Should the Ollama API endpoint be
configurable or hardcoded to `localhost:11434`?

**Findings**:

1. **Ollama default**: Ollama listens on
   `localhost:11434` by default. The `OLLAMA_HOST`
   environment variable can override this.

2. **Dewey configuration**: Dewey's `config.yaml`
   contains the Ollama endpoint URL. However, reading
   Dewey's config adds complexity and a dependency on
   Dewey's config format.

3. **Existing pattern**: The current doctor code uses
   `opts.ExecCmd("ollama", "list")` which relies on the
   `ollama` CLI finding its own server. It doesn't
   configure the endpoint.

**Decision**: Use `OLLAMA_HOST` environment variable
(via `opts.Getenv`) with fallback to `localhost:11434`.
This follows Ollama's own convention and doesn't require
reading Dewey's config. The endpoint is injectable via
`Options.Getenv` for testability.

## R5: HTTP Client for Embedding Check

**Question**: How should the embedding check make HTTP
requests to the Ollama API?

**Findings**:

1. **Standard library**: Go's `net/http` package is
   already available (no new dependency needed).

2. **Testability**: The HTTP call must be injectable for
   unit testing. Options:
   a. Inject an `HTTPClient` interface on `Options`
   b. Inject a function `func(url, body) (response, error)`
   c. Use `opts.ExecCmd("curl", ...)` (fragile, slow)

3. **Existing patterns**: The `Options` struct uses
   function injection (`LookPath`, `ExecCmd`,
   `ReadFile`). An `HTTPGet` or `HTTPPost` function
   field follows this pattern.

**Decision**: Add an `EmbedCheck` function field to
`Options` with signature:
```go
EmbedCheck func(model string) error
```

The production implementation sends a POST to the Ollama
`/api/embed` endpoint. Tests inject a mock. This keeps
the `Options` API clean (callers don't need to know about
HTTP) and follows the existing injection pattern.

The function returns `nil` on success (embeddings work)
or an error describing what failed (connection refused,
model not found, empty response).

## R6: Test Strategy

**Question**: What is the testing approach for these
changes?

**Findings**:

1. **Existing test patterns**: Both `doctor_test.go` and
   `setup_test.go` use injected `Options` with mock
   functions. Tests create temp directories, inject
   `LookPath`/`ExecCmd` mocks, and verify `CheckResult`
   fields.

2. **Coverage targets**: The changes touch two files
   with well-established test patterns. New tests should
   follow the same patterns.

**Decision**: Testing strategy:

- **Unit tests (doctor)**:
  - `TestCheckDewey_EmbeddingCapability_Pass`: Dewey
    installed, embedding check returns nil → PASS
  - `TestCheckDewey_EmbeddingCapability_Fail`: Dewey
    installed, embedding check returns error → FAIL
    with actionable hint
  - `TestCheckDewey_EmbeddingCapability_Skip`: Dewey
    not installed → embedding check skipped
  - `TestCheckDewey_OllamaDemotion`: Verify Ollama
    check message includes "managed by Dewey" note

- **Unit tests (setup)**:
  - `TestInstallSwarmPlugin_ForkSource`: Verify install
    command uses `github:unbound-force/swarm-tools`
  - `TestInstallSwarmPlugin_ReplacesUpstream`: Verify
    fork install runs even when swarm binary exists
  - Update existing tests that assert on the old
    install command strings

- **Coverage strategy**: Unit tests only. No integration
  tests needed — the changes are to injectable function
  calls, not to actual HTTP or subprocess behavior.

## R7: JSON Output Consistency (FR-007)

**Question**: How do the new checks appear in
`--format=json` output?

**Findings**:

1. **Automatic**: The `CheckResult` struct serializes
   to JSON via `json.MarshalIndent`. New fields in
   `CheckResult` are automatically included.

2. **Severity values**: `Pass`, `Warn`, `Fail` serialize
   as `"pass"`, `"warn"`, `"fail"` via `MarshalJSON`.

3. **No new fields needed**: The embedding capability
   check uses the existing `CheckResult` struct with
   `Name`, `Severity`, `Message`, `InstallHint`. No
   schema changes required.

**Decision**: No special JSON handling needed. The
existing serialization handles everything. FR-007 is
satisfied automatically by using the standard
`CheckResult` struct.
