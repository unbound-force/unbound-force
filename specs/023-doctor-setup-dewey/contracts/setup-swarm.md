# Contract: Setup Swarm Plugin Fork Installation

**Package**: `internal/setup`  
**File**: `setup.go`

## Function: `installSwarmPlugin` (modified)

```go
func installSwarmPlugin(
    opts *Options,
    env doctor.DetectedEnvironment,
) stepResult
```

### Changes from Current Behavior

1. **Install source**: Changed from npm registry package
   (`opencode-swarm-plugin@latest`) to GitHub fork
   (`github:unbound-force/swarm-tools`)
2. **Idempotent update**: The function no longer returns
   early with "already installed" when the `swarm` binary
   exists. Instead, it always runs the install command to
   ensure the fork version is current.

### Behavior Matrix

| Condition | Action | Result |
|-----------|--------|--------|
| swarm not installed, bun available | `bun add -g github:unbound-force/swarm-tools` | installed via bun |
| swarm not installed, npm only | `npm install -g github:unbound-force/swarm-tools` | installed via npm |
| swarm installed (any source), bun available | `bun add -g github:unbound-force/swarm-tools` | updated via bun |
| swarm installed (any source), npm only | `npm install -g github:unbound-force/swarm-tools` | updated via npm |
| bun install fails | fall through to npm | installed via npm |
| both fail | return failed | failed |
| dry-run mode | report what would happen | dry-run |

### Dry-Run Messages

| Condition | Detail |
|-----------|--------|
| bun available | "Would install: bun add -g github:unbound-force/swarm-tools" |
| npm only | "Would install: npm install -g github:unbound-force/swarm-tools" |

### Install Hint Updates

The following install hints in `environ.go` must also be
updated to reference the fork:

| Function | Current | New |
|----------|---------|-----|
| `managerInstallCmd("swarm", ManagerBun)` | `bun add -g opencode-swarm-plugin@latest` | `bun add -g github:unbound-force/swarm-tools` |
| `homebrewInstallCmd("swarm")` | `npm install -g opencode-swarm-plugin@latest` | `npm install -g github:unbound-force/swarm-tools` |
| `genericInstallCmd("swarm")` | `npm install -g opencode-swarm-plugin@latest` | `npm install -g github:unbound-force/swarm-tools` |
| `checkSwarmPlugin` install hint | `npm install -g opencode-swarm-plugin@latest` | `npm install -g github:unbound-force/swarm-tools` |

### Postconditions

- The `swarm` binary is installed from the
  `unbound-force/swarm-tools` fork
- The binary name remains `swarm` (unchanged)
- Existing `swarm` binary detection (`LookPath("swarm")`)
  continues to work
- The `opencode.json` plugin config check
  (`"opencode-swarm-plugin"` in plugin array) is
  unchanged — the plugin name in the config is the same
  regardless of install source
