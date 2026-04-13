# Quickstart: Sandbox CDE Lifecycle

**Branch**: `029-sandbox-cde-lifecycle` | **Date**: 2026-04-13

## For Engineers (Podman Backend)

### 1. Create a Persistent Workspace

```bash
# From your project directory
uf sandbox create

# Output:
# Creating persistent workspace "uf-sandbox-myproject"...
# Creating named volume...
# Starting container...
# Copying project source...
# Waiting for OpenCode server...
# ✓ Workspace ready.
# Server: http://localhost:4096
# Attaching to sandbox...
```

### 2. Work Iteratively

```bash
# Run /unleash for the first iteration
# (inside the sandbox TUI)
/unleash

# Review the demo in your browser:
#   http://localhost:3000 (web app)
#   http://localhost:8080 (API)

# Provide feedback
/speckit.clarify

# Re-run /unleash — picks up from clarify step
/unleash

# Detach when done (Ctrl+C or /quit)
```

### 3. Stop and Resume

```bash
# Stop the workspace (state preserved)
uf sandbox stop
# Output: Sandbox stopped. State preserved in volume
#         "uf-sandbox-myproject".

# Later: resume exactly where you left off
uf sandbox start
# Output: Resuming workspace "uf-sandbox-myproject"...
#         ✓ Workspace ready.

# Reattach
uf sandbox attach
```

### 4. Pull Agent Changes

```bash
# On the host, pull changes the agent pushed
git pull

# Or push spec edits from the host
git push
# The workspace picks them up on next /unleash run
```

### 5. Clean Up

```bash
# When done with the feature
uf sandbox destroy
# Output: Destroy sandbox "uf-sandbox-myproject"?
#         This will permanently delete all workspace state.
#         [y/N] y
#         ✓ Workspace destroyed.
```

---

## For Engineers (CDE Backend)

### 1. Prerequisites

```bash
# Install chectl (Eclipse Che CLI)
# See: https://github.com/che-incubator/chectl#installation
bash <(curl -sL https://che-incubator.github.io/chectl/install.sh)
# or: npm install -g chectl

# Configure Che URL
export UF_CHE_URL=https://che.example.com

# Ensure your project has a devfile
ls devfile.yaml
```

### 2. Create a CDE Workspace

```bash
uf sandbox create --backend che

# Output:
# Creating CDE workspace "uf-myproject"...
# Provisioning from devfile.yaml...
# Waiting for workspace to start...
# ✓ Workspace ready.
# IDE:    https://che.example.com/#/uf-myproject
# Server: https://uf-myproject-opencode.apps.che.example.com
# Demo:   https://uf-myproject-demo-web.apps.che.example.com
#         https://uf-myproject-demo-api.apps.che.example.com
```

### 3. Work in the CDE

```bash
# Attach from host terminal
uf sandbox attach

# Or open the Che IDE in your browser:
#   https://che.example.com/#/uf-myproject

# Run /unleash in the attached TUI
/unleash

# Review demos in the Che IDE:
# - Terminal tab: run CLI commands
# - Browser tab: open demo endpoints
# - Editor: view/edit files
```

### 4. Bidirectional Sync

```bash
# Agent pushes from workspace automatically
# Pull on host to review:
git pull

# Push spec edits from host:
git push
# Workspace picks them up on next /unleash
```

### 5. Clean Up

```bash
uf sandbox destroy --yes
# Output: ✓ CDE workspace "uf-myproject" destroyed.
```

---

## Configuration File

Create `.uf/sandbox.yaml` for persistent settings:

```yaml
# Default backend (auto, podman, che)
backend: auto

# Eclipse Che configuration
che:
  url: https://che.example.com
  # token: ${UF_CHE_TOKEN}  # for REST API auth

# Ollama endpoint (for CDE where
# host.containers.internal doesn't resolve)
ollama:
  host: http://ollama.internal:11434

# Demo ports to expose (Podman only)
demo_ports:
  - 3000
  - 8080
```

---

## Status Check

```bash
uf sandbox status

# Podman persistent:
# Sandbox Status
#   Workspace:  uf-sandbox-myproject
#   Backend:    podman (persistent)
#   Image:      quay.io/unbound-force/opencode-dev:latest
#   State:      running
#   Project:    /Users/dev/myproject
#   Server:     http://localhost:4096
#   Demo:       http://localhost:3000 (demo-web)
#               http://localhost:8080 (demo-api)
#   Started:    2026-04-13T10:00:00Z

# CDE:
# Sandbox Status
#   Workspace:  uf-myproject
#   Backend:    che (persistent)
#   Che URL:    https://che.example.com
#   State:      running
#   Server:     https://uf-myproject-opencode.apps.che.example.com
#   Demo:       https://uf-myproject-demo-web.apps.che.example.com
#   Started:    2026-04-13T10:00:00Z

# No workspace:
# No sandbox workspace found.
# Run `uf sandbox create` to provision one.
```

---

## Backward Compatibility

Existing Spec 028 workflows continue to work:

```bash
# Ephemeral mode (no create step)
uf sandbox start          # creates ephemeral container
uf sandbox attach         # attach TUI
uf sandbox extract --yes  # extract changes
uf sandbox stop           # removes container (no state preserved)
```

The ephemeral mode is the fallback when no persistent
workspace exists. No behavior change for engineers who
don't use `create`/`destroy`.

---

## Troubleshooting

### Manual Cleanup (after crash or interrupted destroy)

```bash
# List orphaned sandbox volumes
podman volume ls | grep uf-sandbox

# Remove a specific orphaned volume
podman rm -f uf-sandbox-myproject 2>/dev/null
podman volume rm uf-sandbox-myproject

# List orphaned CDE workspaces
chectl workspace:list
```

### Common Issues

- **"sandbox already exists"**: Run `uf sandbox destroy`
  first, or `uf sandbox start` to resume.
- **"cannot reach Che"**: Verify `UF_CHE_URL` is set and
  the Che instance is accessible.
- **"devfile.yaml not found"**: Add a devfile to your
  project or use `--backend podman`.
