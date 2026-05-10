## Context

The `uf sandbox` command has a `Backend` interface with
two implementations: `PodmanBackend` (ephemeral and
persistent) and `CheBackend` (Eclipse Che). Research
revealed that the CheBackend is architecturally broken
— modern Che requires a full Kubernetes cluster, and
kubedock (the intended local bridge) cannot host Che.

DevPod (loft-sh/devpod) is a client-only tool that runs
`devcontainer.json` workspaces directly in Docker/Podman
without Kubernetes. It uses the subprocess model (same as
`podman`, `gcloud`, `opencode`) and has no Go library
imports needed.

The gateway proxy (`uf gateway`) handles Vertex AI
credential injection but is only wired into the
ephemeral `Start()` path. Persistent workspaces
(`Create()` and persistent `Start()`) skip it entirely.

## Goals / Non-Goals

### Goals
- Replace CheBackend with DevPodBackend
- Wire gateway auto-start into `Create()` and
  persistent `Start()` dispatch functions
- Scaffold `.devcontainer/devcontainer.json` via
  `uf sandbox init`
- Add DevPod doctor checks (context-sensitive)
- Remove all Che-specific code and config fields

### Non-Goals
- Importing DevPod Go packages (subprocess only)
- Supporting DevPod providers other than Podman
  (Docker, K8s, cloud VMs are future work)
- Multi-workspace management (list, switch)
- Replacing the existing PodmanBackend (both coexist)
- Modifying the `uf gateway` command itself

## Decisions

### D1: DevPod as subprocess, not library

DevPod is invoked via `devpod` CLI subprocess calls
through the existing `ExecCmd` / `ExecInteractive`
injection pattern. No DevPod Go packages are imported.

Rationale: (1) License isolation — MPL-2.0 subprocess
use has zero copyleft impact on Apache-2.0. (2) Follows
the established pattern used for podman, opencode,
gcloud, and chectl. (3) DevPod's CLI is its primary
interface; the Go packages are internal.

### D2: devcontainer.json over devfile

The scaffolded config is `.devcontainer/devcontainer.json`
(the devcontainer spec), not `devfile.yaml`. Rationale:
(1) DevPod consumes `devcontainer.json`, not devfiles.
(2) The devcontainer spec is the industry standard
(VS Code, GitHub Codespaces, JetBrains, DevPod).
(3) Devfiles are specific to Eclipse Che / DevWorkspace
Operator, which we are removing.

### D3: Gateway wiring in dispatch functions

Wire `autoStartGateway()` in the top-level `Create()`
and `Start()` dispatch functions in `sandbox.go`, not
inside individual backend implementations. This ensures
all backends (Podman persistent and DevPod) get gateway
integration automatically, and any future backend also
inherits it.

Gateway port and active flag are passed to backends
via two new fields on `Options`: `GatewayPort int` and
`GatewayActive bool`. Set by dispatch functions after
`autoStartGateway()` returns, before delegating to the
backend.

### D4: DevPod Podman provider auto-configuration

When `DevPodBackend.Create()` runs, it calls
`devpod up --provider podman` to ensure DevPod uses
Podman. If the Podman provider is not configured in
DevPod, `devpod up` auto-configures it on first use.

Gateway env vars are passed to DevPod via the
`--workspace-env KEY=VALUE` flag, which sets environment
variables inside the workspace container. This flag is
stable since DevPod 0.5.x. The `--dotfiles-env` flag
is NOT suitable — it only applies during dotfile
installation and would not make vars available to
OpenCode at runtime.

`DevPodBackend.Create()` MUST also pre-flight check
for `podman` in PATH before calling `devpod up`. If
Podman is not installed, return an actionable error
rather than letting DevPod's error propagate (which
may be less clear).

### D5: DevPod workspace naming convention

DevPod workspaces are named `uf-sandbox-<project-name>`
to match the existing Podman persistent workspace
convention (`containerNameForProject()`,
`volumeNameForProject()`). This is passed via
`devpod up --id uf-sandbox-<project-name>`.

Rationale: consistent naming across backends makes the
mental model simpler. Users see `uf-sandbox-myproject`
regardless of whether the backend is Podman or DevPod.

### D5a: DevPod persistent workspace detection

`isPersistentWorkspace()` currently checks for a Podman
named volume. For DevPod workspaces, it MUST also call
`devpod status uf-sandbox-<project> --output json` to
detect existing DevPod workspaces. If the status
command returns a workspace (regardless of state), the
workspace is considered persistent.

This check is guarded by `LookPath("devpod")` — if
DevPod is not installed, the check is skipped and only
the Podman volume check applies.

### D5b: Minimum DevPod version

DevPod >= 0.5.0 is required. This is the minimum
version that supports `--provider`, `--workspace-env`,
`--id`, and `--output json` flags.

A `parseDevPodVersion()` function MUST be implemented
following the `parsePodmanVersion()` pattern: call
`devpod version`, parse the semver output, and return
an error if the version is below 0.5.0.

### D6: Remove Che code completely

All Che-specific code is deleted — `che.go`,
`BackendChe` constant, `CheURL` on Options,
`resolveCheURL()`, `addCheAuth()`, Che config fields
in `SandboxConfig`, and Che-related test functions.

Rationale: The CheBackend is non-functional for local
use (requires K8s infrastructure). Keeping dead code
violates the zero-waste mandate. If remote Che support
is needed in the future, it can be re-implemented as a
new backend behind a separate spec.

### D7: Devcontainer template as embedded asset

The devcontainer template lives at
`internal/scaffold/assets/devcontainer/devcontainer.json`
and is embedded via `embed.FS`. The template is deployed
to `.devcontainer/devcontainer.json` in the project root
by both `uf init` (standard scaffold) and
`uf sandbox init` (standalone with custom flags).

Deployed by `uf init` alongside agents, commands, and
packs. The devcontainer is a standard spec recognized
by VS Code, GitHub Codespaces, JetBrains, and DevPod —
having it present is harmless for non-DevPod users and
saves an extra step for DevPod users.

`uf sandbox init` remains as a standalone command for
users who want to regenerate the devcontainer with
custom flags (`--image`, `--demo-ports`, `--force`)
without running full `uf init`.

Because this asset IS deployed by `uf init`, it MUST
be added to `expectedAssetPaths` in `scaffold_test.go`
(not `knownNonEmbeddedFiles`). The scaffold engine's
idempotency ensures `uf init` skips the file if it
already exists.

The expected devcontainer.json output:
```json
{
  "image": "quay.io/unbound-force/opencode-dev:latest",
  "forwardPorts": [4096],
  "containerEnv": {
    "ANTHROPIC_BASE_URL":
      "http://host.containers.internal:53147",
    "ANTHROPIC_API_KEY": "gateway"
  },
  "postStartCommand": "nohup opencode serve --port 4096 > /tmp/opencode-server.log 2>&1 & printf 'protocol=https\\nhost=github.com\\n\\n' | git credential fill 2>/dev/null | grep '^password=' | cut -d= -f2 | gh auth login --with-token 2>/dev/null || true",
  "remoteUser": "dev"
}
```

Note: `\\n` in JSON becomes `\n` in shell. The
`postStartCommand` value is not directly paste-able
from the JSON source — the escaping is handled by
the JSON parser at runtime.

Note: `ANTHROPIC_API_KEY=gateway` is a sentinel value
indicating the gateway proxy handles authentication.
The devcontainer template MUST include a JSON comment
(via `//` key or separate doc) explaining this value.
When the gateway is not running, OpenCode will use its
own configured provider — the sentinel value is
harmless in that case.

### D8: Doctor checks context-sensitive

DevPod doctor checks appear only when DevPod is
detected (`devpod` in PATH) or configured
(`sandbox.backend == "devpod"` in config). This avoids
cluttering output for Podman-only users.

Checks: (1) `devpod` binary presence with install hint,
(2) `.devcontainer/devcontainer.json` existence with
`uf sandbox init` hint.

### D9: gh CLI auth relay via postStartCommand

DevPod proxies git credentials into containers via its
own credential helper, but this does NOT relay `gh` CLI
authentication. The `gh` CLI uses a separate token store
(`~/.config/gh/hosts.yml`). Without explicit relay,
`gh` commands (PR reviews, issue management, etc.) fail
inside DevPod sessions.

The `postStartCommand` includes a pipeline that extracts
the GitHub token from DevPod's git credential proxy and
pipes it to `gh auth login --with-token`:

```sh
printf 'protocol=https\nhost=github.com\n\n' \
  | git credential fill 2>/dev/null \
  | grep '^password=' | cut -d= -f2 \
  | gh auth login --with-token 2>/dev/null || true
```

Design choices:
- **`postStartCommand` over `postCreateCommand`**: runs
  on every container start (not just first creation),
  so rotated or refreshed tokens are picked up
  automatically.
- **`|| true` for graceful degradation**: if git
  credentials are not available (e.g., running outside
  DevPod, or `gh` is not installed), the command fails
  silently and the container starts normally.
- **JSON escaping**: `\\n` in JSON becomes `\n` in
  shell. The command is not directly paste-able from
  the JSON source file.
- **Injection safety**: the pipeline uses only hardcoded
  literals and stdin piping. No user-controlled input is
  interpolated into shell commands. The token flows
  through pipes (stdin), never as a command-line
  argument, preventing `/proc/cmdline` exposure.
- **Token persistence**: `gh auth login --with-token`
  overwrites any existing token in
  `~/.config/gh/hosts.yml` (0600 permissions). The
  relayed token inherits whatever scope DevPod's
  credential helper provides (typically a GitHub PAT
  with repo scope). Running on every start ensures
  rotated tokens take effect.
- **Log isolation**: `/tmp/opencode-server.log` captures
  only the backgrounded `opencode serve` process. The
  credential pipeline runs in the foreground with its
  own `2>/dev/null` redirections. Tokens cannot leak
  into the log file.

## Risks / Trade-offs

### R1: DevPod availability

DevPod must be installed separately. Mitigation:
`uf doctor` checks for it, `uf setup` can install it
via Homebrew in a future change, and the Podman backend
continues to work without DevPod.

### R2: DevPod Podman provider networking

DevPod's Podman provider may use different networking
defaults than raw `podman run`. The gateway URL
(`http://host.containers.internal:53147`) must be
reachable from DevPod-managed containers. Mitigation:
DevPod uses Podman directly, so
`host.containers.internal` should resolve the same way.

### R3: DevPod CLI interface stability

`devpod` CLI flags may change between versions.
Mitigation: pin to stable DevPod releases, test against
the CLI in CI, and use `devpod version` for
compatibility checks.

### R4: Che removal is breaking

Users who configured `--backend che` or
`UF_SANDBOX_BACKEND=che` will get an error. Mitigation:
the error message will suggest `--backend devpod` as
the replacement with a migration note. The Che backend
was non-functional for local use, so this breakage
affects zero working setups.

## Test Strategy

All tests are **unit tests** using the established
`ExecCmd` / `LookPath` / `HTTPGet` injection pattern.
No real DevPod binary is required in CI.

- **DevPodBackend methods**: Unit tests with injected
  `ExecCmd` that returns canned `devpod` output. Tests
  verify correct CLI arguments, gateway env var
  injection, error propagation, and status parsing.
- **Gateway wiring (Create/Start)**: Unit tests with
  injected `HTTPGet` (gateway health) and `ExecCmd`
  (gateway start). Verify gateway is called before
  backend delegation.
- **Devcontainer scaffolding**: Unit tests using
  `t.TempDir()`. Verify JSON output, idempotency,
  force overwrite, and custom demo ports.
- **Doctor checks**: Unit tests with injected
  `LookPath`. Verify context-sensitive display and
  install hint content.
- **testOpts() update**: The shared test helper MUST
  be updated so `LookPath` returns `not found` for
  `devpod` by default (prevents auto-detection in
  ephemeral-mode tests). Tests that need DevPod
  override this explicitly.
- **Coverage target**: Maintain existing package
  coverage level (no regression). No integration or
  e2e tests — DevPod binary not available in CI.
<!-- scaffolded by uf vdev -->
