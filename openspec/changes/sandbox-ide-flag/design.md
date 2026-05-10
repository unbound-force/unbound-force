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

## Coverage Strategy

Unit tests only. All new and modified functions
(`validateIDE`, `Create` IDE passthrough, `Start` IDE
passthrough, `DefaultConfig` IDE resolution,
`applySandboxConfig` IDE wiring, `Attach` persistent
workspace detection, `Destroy` ephemeral mode) are
covered via the existing `ExecCmd` injection pattern.
No integration or e2e tests required — IDE passthrough
is verified by asserting `devpod up` arguments contain
the expected `--ide` value. Regression tests verify
the ephemeral Podman path ignores the IDE field and
the Destroy/Attach dispatch handles both persistent
and ephemeral paths correctly.
Coverage target: 100% of new functions.
