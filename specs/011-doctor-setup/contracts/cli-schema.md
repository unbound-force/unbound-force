# CLI Contract: Doctor and Setup Commands

**Branch**: `011-doctor-setup` | **Date**: 2026-03-21

## `unbound doctor`

### Synopsis

```
unbound doctor [flags]
```

### Description

Diagnose the Unbound Force development environment. Checks
for required tools, version managers, scaffolded files, hero
availability, Swarm plugin status, MCP server configuration,
and agent/skill integrity. Produces a colored terminal report
by default, or structured JSON for CI pipelines.

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format` | string | `text` | Output format: `text` or `json` |
| `--dir` | string | `.` | Target directory to check |

### Exit Codes

| Code | Meaning |
|------|---------|
| 0    | All checks pass or only warnings |
| 1    | One or more checks failed |

### Text Output Format

```
Unbound Force Doctor
====================

Detected Environment
  ✓ goenv           Go version manager (/opt/homebrew/bin/goenv)
  ✓ nvm             Node version manager (~/.nvm)
  ✓ Homebrew        Package manager (/opt/homebrew/bin/brew)

Core Tools
  ✓ go              1.24.3 via goenv (~/.goenv/shims/go)
  ✓ opencode        0.2.15 via Homebrew (/opt/homebrew/bin/opencode)
  ✓ gaze            0.10.0 via Homebrew (/opt/homebrew/bin/gaze)
  ✓ mxf             0.5.0 via Homebrew (/opt/homebrew/bin/mxf)
  ○ graphthulhu     not found (optional)
                     Install: brew install unbound-force/tap/graphthulhu
  ✓ node            22.15.0 via nvm (~/.nvm/versions/node/v22.15.0/bin/node)
  ○ gh              not found (optional)
                     Install: brew install gh
  ✓ swarm           installed (/usr/local/bin/swarm)

Swarm Plugin
  ✓ swarm           installed (/usr/local/bin/swarm)
  ── swarm doctor ──────────────────────────
  ✓ OpenCode plugin configured
  ✓ Hive storage: libSQL (embedded SQLite)
  ✓ Semantic memory: ready
  ✓ Dependencies: all installed
  ────────────────────────────────────────
  ✓ .hive/          initialized
  ✓ plugin config   opencode-swarm-plugin in opencode.json

Scaffolded Files
  ✓ .opencode/agents/       12 agent files
  ✓ .opencode/command/      11 command files
  ✓ .specify/               present
  ✓ AGENTS.md               present

Hero Availability
  ✓ Muti-Mind (PO)          agent: muti-mind-po.md
  ✓ Cobalt-Crush (Dev)      agent: cobalt-crush-dev.md
  ✓ Gaze (Tester)           binary: /opt/homebrew/bin/gaze
  ✓ The Divisor (Reviewer)  agent: divisor-guard.md (+4 personas)
  ✓ Mx F (Manager)          binary: /opt/homebrew/bin/mxf

MCP Server Config
  ✓ opencode.json           valid
  ✓ knowledge-graph         graphthulhu binary found

Agent/Skill Integrity
  ✓ 12 agents validated     all frontmatter valid
  ✓ 1 skill validated       unbound-force-heroes

Summary: 25 passed, 0 warnings, 0 failed
```

### Text Output Symbols

| Symbol | Color  | Plain Fallback | Meaning |
|--------|--------|----------------|---------|
| ✓      | Green  | `[PASS]`       | Check passed |
| !      | Yellow | `[WARN]`       | Warning (non-critical) |
| ✗      | Red    | `[FAIL]`       | Check failed (required) |
| ○      | Gray   | `[INFO]`       | Optional item absent |

### JSON Output Schema

See `data-model.md` for the complete Report structure.
The JSON output is the Report struct serialized with
`json.MarshalIndent` using 2-space indentation.

---

## `unbound setup`

### Synopsis

```
unbound setup [flags]
```

### Description

Install and configure the Unbound Force development tool
chain. Detects existing version and package managers, installs
missing tools through the appropriate manager, configures the
Swarm plugin in `opencode.json`, and scaffolds project files.
Idempotent -- safe to run multiple times.

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--dir` | string | `.` | Target directory for setup |
| `--dry-run` | bool | `false` | Print actions without executing |
| `--yes` | bool | `false` | Skip confirmation prompts (required for non-interactive `curl \| bash` installs) |

### Exit Codes

| Code | Meaning |
|------|---------|
| 0    | Setup completed (some steps may have been skipped) |
| 1    | One or more critical steps failed |

### Installation Order

```
1. Detect environment (managers)
2. Install OpenCode     (if missing, via Homebrew or curl)
3. Install Gaze         (if missing, via Homebrew)
4. Check Node.js >= 18  (if missing, install via nvm/fnm/brew)
5. Install Swarm plugin (if missing, via npm or bun)
6. Run swarm setup      (if not configured)
7. Configure opencode.json (add plugin entry if missing)
8. Run swarm init       (if .hive/ missing)
9. Run unbound init     (if .opencode/ missing)
```

### Text Output Format

```
Unbound Force Setup
===================

Detected Environment
  goenv (Go), nvm (Node.js), Homebrew (packages)

Installing...
  ✓ OpenCode        already installed (0.2.15 via Homebrew)
  ✓ Gaze            already installed (0.10.0 via Homebrew)
  ✓ Node.js         already installed (22.15.0 via nvm)
  ✓ Swarm plugin    installed via npm (opencode-swarm-plugin@1.2.3)
  ✓ swarm setup     completed
  ✓ opencode.json   plugin configured
  ✓ .hive/          initialized
  ✓ unbound init    already scaffolded

Setup complete! Run `unbound doctor` to verify.
```

### Error Output Example

```
Unbound Force Setup
===================

Detected Environment
  Homebrew (packages)

Installing...
  ✓ OpenCode        installed via Homebrew
  ✓ Gaze            installed via Homebrew
  ! Node.js         not found
                     Install: brew install node
                     Skipping Swarm setup (requires Node.js)
  - Swarm plugin    skipped (no Node.js)
  - swarm setup     skipped (no swarm)
  - opencode.json   skipped (no swarm)
  - .hive/          skipped (no swarm)
  ✓ unbound init    scaffolded (47 files)

Setup partially complete. Install Node.js, then re-run
`unbound setup` to complete Swarm configuration.
```
