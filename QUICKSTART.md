# Quick Start

Unbound Force adds AI-powered development workflows to
your project:

- **Code review council** -- 5 AI reviewer personas
  audit your code for security, architecture, testing,
  operations, and intent drift
- **Specification-driven development** -- structured
  workflows from idea to implementation
- **Quality analysis** -- CRAP scores, coverage metrics,
  and test generation (Go projects)

Designed for [OpenCode](https://opencode.ai). The
scaffolded files are portable Markdown that can be
adapted for other AI coding tools.

## Choose Your Path

### Option A: OpenPackage (no binary required)

```bash
opkg install @unbound-force/review-council
# or for the full workflow suite:
opkg install @unbound-force/workflows
```

### Option B: `uf` binary (full tool suite)

```bash
brew install unbound-force/tap/unbound-force
```

With `opkg` on your `PATH`, `uf init` runs the same OpenPackage installs
(OpenPackage bundles under `packages/`) instead of rewriting those paths
from the embedded fallback inside the binary. Without `opkg`, behavior
matches Option B as before.
## Prerequisites

- **git** -- version control (required)
- **LLM API key** -- OpenCode needs an LLM provider.
  See [OpenCode provider docs](https://opencode.ai/docs/providers)
  for setup (Anthropic, OpenAI, Google, AWS Bedrock,
  and others supported).
- **Go 1.24+** -- only if your project is Go-based
  (used by review council CI checks and Gaze quality
  analysis)

## Install

### macOS (Homebrew)

```bash
brew install unbound-force/tap/unbound-force
```

#### Podman Machine Setup (for `uf sandbox`)

If you plan to use `uf sandbox` for containerized
development sessions, ensure your Podman machine is
configured for correct UID mapping:

```bash
podman machine stop
podman machine rm
podman machine init --rootful=false
podman machine start
```

This ensures virtiofs maps file ownership correctly
inside sandbox containers. Without this, files appear
as `root:nobody` and `direct` mode is non-functional.

If you cannot reconfigure your Podman machine, use
the `--uidmap` escape hatch:

```bash
uf sandbox start --uidmap
```

### Fedora / RHEL (Homebrew -- recommended)

Homebrew provides access to all companion tools via
`uf setup`. Install Homebrew for Linux first, then use
the same command as macOS:

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
brew install unbound-force/tap/unbound-force
```

### Fedora / RHEL (dnf -- minimal)

Installs the `uf` binary only. Additional tools must be
installed separately or via Homebrew later.

```bash
# Install uf (latest RPM, auto-resolved)
sudo dnf install -y "$(
  curl -fsSL \
    https://api.github.com/repos/unbound-force/unbound-force/releases/latest |
  grep -o 'https://[^"]*linux_amd64\.rpm'
)"

# Install OpenCode
curl -fsSL https://opencode.ai/install | bash
```

For ARM64 systems, replace `amd64` with `arm64` in the
grep pattern.

## For Project Maintainers

Add Unbound Force to your project:

```bash
cd your-project
uf init
```

This scaffolds agents, commands, convention packs, and
workflow configuration into your project. Tool-owned files
are auto-updated on re-run; user-owned files (like custom
convention packs) are never overwritten.

Commit and push the scaffolded files:

```bash
git add .opencode/ openspec/ .specify/ opencode.json
git commit -m "chore: add Unbound Force framework"
git push
```

For code review only (no spec workflows), use the subset:

```bash
uf init --divisor
```

## For Contributors

Set up your development environment in a project that
uses Unbound Force:

```bash
uf setup        # installs recommended tools
uf doctor       # verify everything works
```

Preview what `uf setup` will install before running:

```bash
uf setup --dry-run
```

Most tools are optional. The core experience (code review,
spec workflows) requires only `uf` and `opencode`.

## Your First Review

Start OpenCode and run the Divisor review council:

```bash
opencode
```

Inside OpenCode:

```
/review-council
```

The council discovers available reviewer agents and runs
them in parallel. Each persona focuses on a different
aspect -- security, architecture, testing, operations,
and intent alignment. You receive an **APPROVE** or
**REQUEST CHANGES** verdict with specific findings.

## Next Steps

- **[USAGE.md](USAGE.md)** -- Common workflows, agents,
  and command reference for daily use
- **Specification workflows** -- `/opsx-propose` for
  small changes, `/speckit.specify` for features
- **Autonomous pipeline** -- `/unleash` runs the full
  workflow from spec to code review in one command
- **Full tool suite** -- `uf setup` installs all
  companion tools (Gaze, Dewey, Replicator)
- **[AGENTS.md](AGENTS.md)** -- Full reference for AI
  agents and power users
