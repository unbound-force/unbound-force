# Quickstart: Doctor and Setup Commands

**Branch**: `011-doctor-setup` | **Date**: 2026-03-21

## Prerequisites

- `unbound` CLI installed (`brew install unbound-force/tap/unbound`)
- A project directory (ideally a git repository)
- Node.js >= 18 recommended (required for Swarm plugin
  installation). If missing, `unbound setup` will install
  other tools and provide Node.js install instructions.

## Full Environment Setup (New Developer)

Run `unbound setup` to install the entire Unbound Force
tool chain in one command:

```bash
cd /path/to/your/project
unbound setup
```

This will:
1. Detect your existing version managers (goenv, nvm, etc.)
2. Install OpenCode (via Homebrew or curl)
3. Install Gaze (via Homebrew)
4. Install Swarm plugin (via npm)
5. Configure `opencode.json` with the Swarm plugin
6. Initialize `.hive/` for Swarm work tracking
7. Run `unbound init` to scaffold project files

## Check Environment Health

After setup (or any time), run `unbound doctor` to verify
everything is correctly configured:

```bash
unbound doctor
```

This checks 7 areas:
- **Detected Environment** -- which version managers are
  active (goenv, nvm, Homebrew, etc.)
- **Core Tools** -- go, opencode, gaze, mxf, graphthulhu,
  node, gh, swarm
- **Swarm Plugin** -- installed, configured, healthy
  (includes `swarm doctor` output)
- **Scaffolded Files** -- `.opencode/`, `.specify/`,
  `AGENTS.md`
- **Hero Availability** -- Muti-Mind, Cobalt-Crush, Gaze,
  Divisor, Mx F
- **MCP Server Config** -- `opencode.json` validity and
  server binaries
- **Agent/Skill Integrity** -- frontmatter validation

## JSON Output for CI

```bash
unbound doctor --format=json
```

Use the exit code for CI gating:
- Exit 0 = all pass or warnings only
- Exit 1 = one or more failures

## Common Workflows

### Fix a failed check

```bash
# 1. Run doctor to see what's wrong
unbound doctor

# 2. Copy the install hint from the failed check
#    (e.g., "brew install unbound-force/tap/gaze")

# 3. Run the install command

# 4. Re-run doctor to verify
unbound doctor
```

### Re-run setup after installing Node.js

If `unbound setup` skipped Swarm because Node.js was
missing:

```bash
# Install Node.js through your preferred manager
nvm install 22
# or: fnm install 22
# or: brew install node

# Re-run setup -- it will pick up where it left off
unbound setup
```

### Check a different directory

```bash
unbound doctor --dir=/path/to/other/project
unbound setup --dir=/path/to/other/project
```
