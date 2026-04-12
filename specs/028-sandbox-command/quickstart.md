# Quickstart: Sandbox Command

**Branch**: `028-sandbox-command` | **Date**: 2026-04-12

## Prerequisites

- Podman installed (`brew install podman` or
  https://podman.io/getting-started/installation)
- Podman machine running (`podman machine start`)
- OpenCode installed (`brew install anomalyco/tap/opencode`)
- Ollama running (recommended, not required)
- LLM API key set (e.g., `ANTHROPIC_API_KEY`)

## Basic Workflow

### 1. Start an isolated sandbox

```bash
uf sandbox start
```

This will:
1. Check for Podman and OpenCode
2. Detect your platform (macOS/Linux, arm64/amd64)
3. Pull the container image (first run only)
4. Start the container with your project mounted read-only
5. Wait for the OpenCode server to be ready
6. Attach your terminal to the OpenCode TUI

### 2. Work inside the sandbox

The agent works inside the container with full toolchain
access. Your host filesystem is protected — the project
is mounted read-only with a writable overlay.

### 3. Extract changes

When the agent has made changes you want to keep:

```bash
uf sandbox extract
```

This will:
1. Generate a patch from the container's git history
2. Show you what changed (files, insertions, deletions)
3. Ask for confirmation before applying
4. Apply the patch to your host repo

### 4. Stop the sandbox

```bash
uf sandbox stop
```

## Common Scenarios

### Start in direct mode (read-write)

```bash
uf sandbox start --mode direct
```

Changes are written directly to your host filesystem.
No extraction needed, but no isolation either.

### Start detached (background)

```bash
uf sandbox start --detach
```

Prints the server URL and exits. Attach later with:

```bash
uf sandbox attach
```

### Check sandbox status

```bash
uf sandbox status
```

Shows container name, uptime, mode, mounted project,
and server URL.

### Use a custom image

```bash
uf sandbox start --image myregistry/myimage:latest
```

Or set the environment variable:

```bash
export UF_SANDBOX_IMAGE=myregistry/myimage:latest
uf sandbox start
```

### Adjust resource limits

```bash
uf sandbox start --memory 16g --cpus 8
```

## Troubleshooting

| Problem | Solution |
|---------|----------|
| "podman not found" | Install Podman: `brew install podman` |
| "opencode not found" | Install OpenCode: `brew install anomalyco/tap/opencode` |
| "sandbox already running" | Run `uf sandbox stop` first, or `uf sandbox attach` |
| Health check timeout | Check container logs: `podman logs uf-sandbox` |
| Port 4096 in use | Stop the conflicting process or the existing sandbox |
| SELinux permission denied | The command auto-detects SELinux — file a bug if it fails |
| No API key warning | Set `ANTHROPIC_API_KEY` or equivalent in your shell |
| Image pull slow/fails | Check network; retry with `podman pull quay.io/unbound-force/opencode-dev:latest` |
