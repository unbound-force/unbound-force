## Context

The DevPod backend passes `--ide none` to `devpod up`
in both `Create()` (devpod.go:70) and `Start()`
(devpod.go:104, implicitly — no `--ide` is passed on
resume). DevPod supports `none`, `vscode`, `openvscode`,
`fleet`, `jupyternotebook`, and `cursor` as IDE values.

The OpenCode TUI runs as a server on port 4096 inside
the container and is independent of the IDE choice. Both
can run simultaneously.

## Goals / Non-Goals

### Goals

- Add `IDE` field to `Options` struct
- Pass `--ide <value>` to `devpod up` in Create and Start
- Support configuration via CLI flag, env var, and config
- Default to `"none"` for backward compatibility
- Validate the IDE value against DevPod's supported list

### Non-Goals

- Installing VS Code or any IDE automatically
- Managing IDE extensions inside the container
- Modifying the ephemeral Podman sandbox path (IDE flag
  is DevPod-only)
- Adding IDE support to `uf sandbox init` devcontainer
  template (VS Code extensions are configured via
  devcontainer.json, not by this change)

## Decisions

### D1: IDE value on Options struct

Add `IDE string` to the `Options` struct alongside the
existing `BackendName`, `Detach`, and `Image` fields.
The field follows the same resolution chain as other
config values: CLI flag > env var > config file >
default.

### D2: Default value is "none"

When `IDE` is empty after resolution, it defaults to
`"none"`. This preserves backward compatibility — the
OpenCode TUI is the primary interface, and IDE
integration is opt-in.

### D3: Validation against supported IDE list

Validate the IDE value against a constant list of
supported values (`none`, `vscode`, `openvscode`,
`fleet`, `jupyternotebook`, `cursor`). Invalid values
produce an error before calling `devpod up`. This
prevents confusing DevPod error messages.

### D4: IDE flag on both create and start

The `--ide` flag is added to both `uf sandbox create`
and `uf sandbox start` cobra commands. On `create`, it
sets the initial IDE. On `start` (resume), it allows
changing the IDE for the resumed session — DevPod
supports this natively.

### D5: Config file support

The `.uf/sandbox.yaml` config gains an `ide` field:

```yaml
sandbox:
  ide: vscode  # default IDE for DevPod workspaces
```

This allows teams to set a project-wide default without
requiring the flag on every invocation.

### D6: Environment variable

`UF_SANDBOX_IDE` env var provides per-session override
without modifying config files. Resolution order:
`--ide` flag > `UF_SANDBOX_IDE` > `.uf/sandbox.yaml`
ide field > `"none"`.

### D7: Attach detects persistent workspaces

`Attach()` now checks `isPersistentWorkspace()` before
the ephemeral container check, matching the pattern
already used by `Stop()`. This fixes a bug where
`uf sandbox attach` could not find running DevPod
workspaces because it only checked for the ephemeral
Podman container name.

### D8: DevPod Start health check wait

`DevPodBackend.Start()` now calls `waitForHealth()`
after `devpod up` returns and before attempting TUI
attach. DevPod containers may take a moment to start
the OpenCode server (via `postStartCommand`). If the
health check times out, the command prints a warning
and returns gracefully — the IDE may still be
connected even if the TUI cannot attach.

### D9: devcontainer postStartCommand

The devcontainer.json template includes a
`postStartCommand` that starts the OpenCode server
in the background. DevPod overrides the container
entrypoint with its own agent process, so the
server does not auto-start from the image entrypoint.
The `postStartCommand` runs after the container is
ready, starting the server on port 4096.

### D10: Destroy ephemeral mode fix

`Destroy()` now checks `isPersistentWorkspace()`
before calling `ResolveBackend()`. Without this check,
`ResolveBackend()` auto-detects DevPod (when `devpod`
is in PATH and `devcontainer.json` exists) even for
ephemeral containers, causing `devpod delete` to fail
with "workspace not found". For ephemeral containers,
`Destroy()` now cleans up directly via
`podman stop` + `podman rm`.

### D11: DevPod Create health check parity

`DevPodBackend.Create()` calls `waitForHealth()` after
`devpod up` returns, matching the existing pattern in
`Start()`. Without this, the OpenCode server inside the
container may not be ready when the user tries to
attach immediately after creation.

### D12: DevPod stderr suppression for tunnel errors

DevPod v0.6.x has a known Bun runtime bug where the
VS Code tunnel crashes after successful workspace
creation, producing a noisy `fetch()` stack trace on
stderr. When `devpod up` returns non-zero, `Create()`
checks `devpod status <ws> --output json` to determine
if the workspace was actually created. If the workspace
is `Running`, the error is treated as a non-fatal
tunnel failure: the raw stderr is suppressed and a
friendly warning is printed. If the workspace is NOT
running, the full error is reported.

This approach avoids string-matching on DevPod's
internal error messages (which are unstable across
versions) and instead uses the workspace status as the
source of truth.

### D13: Start SSH fallback for server start

DevPod snapshots the `devcontainer.json` at workspace
creation time. Workspaces created before the
`postStartCommand` was added will never have the
OpenCode server auto-start. When `waitForHealth()`
times out in both `Create()` and `Start()`, the command
attempts to start the server via SSH.

**Injection safety**: The workspace name MUST be passed
as a separate `exec.Command` argument, never
interpolated into a shell string. The SSH command
after `--` is a hardcoded literal with no
user-controlled interpolation:

```go
opts.ExecCmd("devpod", "ssh", wsName, "--",
    "sh", "-c",
    "nohup opencode serve --port 4096 "+
        "> /tmp/opencode-server.log 2>&1 &")
```

The workspace name (`wsName`) is already sanitized by
`projectName()` which strips all characters except
`[a-z0-9-]`. The command after `--` contains only
hardcoded literals — no user input is interpolated.

A second `waitForHealth()` call follows with a shorter
timeout to verify the server started. If both attempts
fail, the command prints a warning and returns without
error.

**Concurrent server start**: If the server is already
running (e.g., `postStartCommand` succeeded but the
health check timed out due to network latency), the
SSH fallback will fail with "address already in use"
on port 4096. This is non-fatal — the second
`waitForHealth()` will succeed because the server is
already responding.

### D14: Destroy confirmation — replace `fmt.Fscanln`

The confirmation prompt in `runSandboxDestroy()` uses
`fmt.Fscanln` which is fundamentally broken for
interactive prompts: it uses whitespace-delimited token
scanning, not line-oriented reading. On macOS iTerm2,
pressing Enter can send a bare `\r` (0x0D), causing
`fmt.Fscanln` to block indefinitely.

Replace `fmt.Fscanln(p.stdin, &response)` with
`bufio.NewScanner(p.stdin)` + `scanner.Scan()` +
`scanner.Text()`. This correctly handles `\n`, `\r\n`,
and bare `\r` line endings. It also handles EOF from
piped input — `scanner.Scan()` returns `false` and
`scanner.Text()` returns `""`, which is treated as
cancellation.

Behavior: any input that is not `"y"` or `"yes"`
(case-insensitive) prints "Cancelled." and returns
nil. This covers empty Enter, "n", "no", EOF, and
bare `\r`. The `--yes` flag bypasses the prompt
entirely.

## Risks / Trade-offs

### R1: IDE availability not checked

The sandbox does not verify that the selected IDE is
installed on the host. If a user passes `--ide vscode`
but VS Code is not installed, `devpod up` will fail
with its own error message. This is acceptable because
DevPod's error messages for missing IDEs are clear.

### R2: IDE flag ignored for ephemeral Podman sandbox

The `--ide` flag only affects the DevPod backend. If
the user is on the ephemeral Podman path (no
`uf sandbox create`), the flag is silently ignored.
The `--ide` flag help text should note this.

## Coverage Strategy

Unit tests only. All new and modified functions are
covered via the existing `ExecCmd` injection pattern.
No integration or e2e tests required.

### IDE flag functions (D1-D6)

- `validateIDE`: table-driven valid/invalid values
- `Create` IDE passthrough: assert `--ide` in args
- `Start` IDE passthrough: assert `--ide` in args
- `DefaultConfig` IDE resolution: flag > env > default
- `applySandboxConfig` IDE wiring: config file fallback
- Ephemeral ignores IDE: `Start()` without persistent
  workspace does not pass `--ide` to `podman run`

### Lifecycle fixes (D7-D10)

- `Attach` persistent workspace detection: delegates
  to backend when `isPersistentWorkspace()` is true
- `Destroy` ephemeral mode: handles cleanup directly,
  does not call `ResolveBackend()`
- `waitForHealth`: immediate success, delayed success
  (retry path), timeout

### Manual testing bug fixes (D11-D14)

- `Create` health check (D11): verify `waitForHealth`
  called after `devpod up`, warn on timeout
- `Create` stderr suppression (D12): tunnel error
  (status=Running → suppress) vs real failure
  (status≠Running → report) vs status check failure
  (report original error)
- `Start`/`Create` SSH fallback (D13): health timeout
  triggers SSH server start, second health check;
  both-fail path prints warning; concurrent server
  start (port-in-use) is non-fatal
- `Destroy` confirmation (D14): empty input (Enter),
  explicit "n", EOF/pipe, bare `\r` — all print
  "Cancelled." via `bufio.Scanner`

### OS-aware devcontainer runArgs (D15)

`.devcontainer/devcontainer.json` is now gitignored and
generated per-user by `uf sandbox init`. The `runArgs`
are OS-specific:

- **macOS** (`darwin`): `--userns=keep-id:uid=1000,gid=1000`
  Maps the host user through the Podman VM to the
  container's dev user (UID 1000). Required because
  macOS runs Podman in a Linux VM with different UID
  semantics.
- **Linux** (default): `--userns=keep-id`
  Plain keep-id without explicit uid/gid suffix. The
  explicit range breaks container restart on Fedora
  rootless Podman where subuid-mapped UIDs (e.g.,
  4203716) fall outside the 0-1000 range.

`uf init` no longer deploys the devcontainer template
(skipped in scaffold walk). The embedded asset remains
for `DevcontainerContent()` used by `uf sandbox init`.

`json.MarshalIndent` replaced with `json.NewEncoder` +
`SetEscapeHTML(false)` to prevent shell characters
(`>`, `&`) in `postStartCommand` from being escaped
to `\u003e`/`\u0026`.

### Manual verification (not unit-testable)

- `postStartCommand` (D9): verified by task 7.4
  (`uf sandbox create --backend devpod --ide vscode`).
  Unit tests verify the `postStartCommand` value in
  the embedded devcontainer.json template.
- Tunnel error suppression (D12): DevPod Bun bug
  cannot be reliably triggered in tests

Coverage target: 100% of new functions.
