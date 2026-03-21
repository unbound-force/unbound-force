# Research: Doctor and Setup Commands

**Branch**: `011-doctor-setup` | **Date**: 2026-03-21

## R1: Version Manager Detection Heuristics

**Decision**: Use a priority-ordered chain of path pattern
matching and environment variable checks to detect which
version/package manager installed each binary.

**Rationale**: Binary paths contain reliable signals
(e.g., `/.goenv/shims/`, `/.nvm/versions/`, `/Cellar/`).
Path-based detection works in non-interactive contexts
where environment variables may not be set.

**Detection order** (most specific first):
1. goenv — path contains `/.goenv/shims/` or `/.goenv/versions/`
2. pyenv — path contains `/.pyenv/shims/` or `/.pyenv/versions/`
3. nvm — path contains `/.nvm/versions/`
4. fnm — path contains `/fnm_multishells/` or `/fnm/node-versions/`
5. mise — path contains `/mise/installs/` or `/mise/shims/`
6. bun — path contains `/.bun/bin/`
7. Homebrew — resolved symlink path contains `/Cellar/`
8. Direct install — path starts with `/usr/local/go/bin/`
9. System — path starts with `/usr/bin/`, `/snap/bin/`
10. Unknown — no match

**Alternatives considered**:
- Shelling out to each manager (`goenv version-name`, etc.)
  — rejected because managers may not be in PATH in
  non-interactive contexts and adds subprocess overhead.
- Checking only environment variables — rejected because
  env vars may not be set when shell init hasn't run.
- Using `brew list --formula` to detect Homebrew — rejected
  because it's slow (shells out to brew) and not needed when
  symlink resolution works reliably.

### Critical implementation notes

- **goenv and pyenv use shims** (small bash scripts), not
  symlinks. `filepath.EvalSymlinks` does NOT resolve through
  them. Use `strings.Contains` path matching.
- **nvm is a bash function**, not a binary. Cannot be invoked
  via `exec.Command`. Detect by path pattern and `NVM_DIR`.
- **Homebrew `/usr/local/bin` is ambiguous** on Intel macOS.
  MUST resolve symlinks and verify `/Cellar/` in the target.
- **Version extraction from shim paths** is not possible.
  For goenv/pyenv, extract version from env vars
  (`GOENV_VERSION`, `PYENV_VERSION`) or version files
  (`.go-version`, `~/.goenv/version`). For nvm/fnm, extract
  version from the path itself (e.g., `/.nvm/versions/node/v22.15.0/`).

## R2: Colored Terminal Output with lipgloss

**Decision**: Use `charmbracelet/lipgloss` (already an
indirect dependency) for colored styling. Use
`muesli/termenv` (transitive dependency) for color profile
detection and `NO_COLOR` support.

**Rationale**: lipgloss is already in the dependency tree
(via `charmbracelet/log`). It handles `NO_COLOR` env var
and terminal capability detection automatically through
its `ColorProfile()` function. No additional dependencies
needed.

**Implementation pattern**:
```
renderer := lipgloss.NewRenderer(stdout)
pass := renderer.NewStyle().Foreground(lipgloss.Color("2"))
warn := renderer.NewStyle().Foreground(lipgloss.Color("3"))
fail := renderer.NewStyle().Foreground(lipgloss.Color("1"))
dim  := renderer.NewStyle().Foreground(lipgloss.Color("8"))
```

When the color profile is `Ascii` (no color support, pipe,
or `NO_COLOR` set), lipgloss renders unstyled text
automatically. The fallback plain indicators (`[PASS]`,
`[WARN]`, `[FAIL]`) should be used when the color profile
is `Ascii`, detected via the renderer's color profile
(automatic when using `lipgloss.NewRenderer(writer)`).

**Alternatives considered**:
- `fatih/color` — rejected, adds a new dependency.
- Plain ANSI escape codes — rejected, no `NO_COLOR` support.
- `charmbracelet/log` for output — rejected, log is for
  structured logging, not report formatting.

## R3: opencode.json Manipulation

**Decision**: Use `encoding/json` with `json.RawMessage` to
preserve unknown fields when adding the `plugin` array entry.

**Rationale**: `opencode.json` has a rich schema that we
don't fully model. Using `json.RawMessage` for unknown
fields preserves any config (MCP servers, agents, themes)
that our code doesn't know about.

**Implementation pattern**:
1. Read file → `json.Unmarshal` into
   `map[string]json.RawMessage`
2. Extract `"plugin"` key → unmarshal to `[]string`
3. Append `"opencode-swarm-plugin"` if not present
4. Marshal back → `json.MarshalIndent` with 2-space indent
5. Write atomically (write to temp, rename)

**Alternatives considered**:
- Full struct modeling of `opencode.json` — rejected, too
  fragile if OpenCode adds new config keys.
- String manipulation / sed-style editing — rejected, not
  safe for JSON.
- `tidwall/sjson` — rejected, adds a new dependency.

## R4: Swarm Doctor Output Integration

**Decision**: Shell out to `swarm doctor` with a 10-second
timeout via `context.WithTimeout`, capture combined
stdout/stderr, and embed verbatim.

**Rationale**: The `swarm doctor` command outputs 4 lines
of status checks. Its output contract is: stdout for
status, exit code 0 for healthy. Embedding verbatim avoids
duplicating Swarm's validation logic and stays in sync with
Swarm updates.

**Expected output format** (from Swarm docs):
```
✓ OpenCode plugin configured
✓ Hive storage: libSQL (embedded SQLite)
✓ Semantic memory: ready
✓ Dependencies: all installed
```

**Implementation pattern**:
```
ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
defer cancel()
cmd := exec.CommandContext(ctx, "swarm", "doctor")
output, err := cmd.CombinedOutput()
```

**Alternatives considered**:
- Parsing Swarm's internal state files — rejected, fragile
  and would drift from Swarm's actual checks.
- Running `swarm doctor --json` — not available; Swarm
  doctor only outputs text.

## R5: Reusing DetectHeroes()

**Decision**: Reuse `internal/orchestration.DetectHeroes()`
directly from the doctor package for the Hero Availability
check group.

**Rationale**: DetectHeroes already implements the canonical
hero detection logic with injectable `lookPath`. No need to
duplicate. The doctor package imports `orchestration` and
calls `DetectHeroes(agentDir, lookPath)`.

**Alternatives considered**:
- Copying hero detection logic into doctor — rejected,
  violates DRY and could drift.
- Creating a shared interface — rejected, the function
  signature is already clean and injectable.

## R6: YAML Frontmatter Parsing

**Decision**: Parse YAML frontmatter manually by splitting
on `---` delimiters, then unmarshal with `gopkg.in/yaml.v3`
(already a dependency).

**Rationale**: Agent `.md` files use standard YAML
frontmatter between `---` delimiters. The project already
uses `yaml.v3` for convention pack validation. No new
dependency needed.

**Implementation pattern**:
1. Read file content
2. Check starts with `---\n`
3. Find second `---\n`
4. Extract YAML between delimiters
5. `yaml.Unmarshal` into `map[string]interface{}`
6. Check for required keys (`description`, `name`, etc.)

**Alternatives considered**:
- `adrg/frontmatter` — rejected, adds a new dependency.
- Regex-based extraction — rejected, fragile.

## R7: Setup Installation Order and Error Handling

**Decision**: Install in dependency order (OpenCode, Gaze,
Node.js, Swarm plugin, Swarm config, hive init,
unbound init) with continue-on-independent-failure
semantics.

**Rationale**: Some steps depend on prior steps (Swarm
requires Node.js), but others are independent (OpenCode
and Gaze can be installed regardless of each other). The
setup command should maximize successful installs even when
some fail.

**Dependency chain**:
```
OpenCode  (independent)
Gaze      (independent)
Node.js   (independent, but required for below)
  └── Swarm plugin   (requires npm from Node.js)
       └── swarm setup  (requires swarm binary)
       └── opencode.json config (requires swarm binary)
       └── swarm init   (requires swarm binary)
unbound init (independent, runs last)
```

**Error handling rules**:
- OpenCode install failure → warn, continue to Gaze
- Gaze install failure → warn, continue to Node.js check
- Node.js missing → warn, skip all Swarm steps, continue
  to unbound init
- npm/Swarm install failure → warn, skip swarm setup/init
- swarm setup failure → warn, continue to opencode.json
  config and swarm init
- unbound init failure → warn (final step, nothing to skip)

**Alternatives considered**:
- Fail-fast on any error — rejected, a developer who has
  Homebrew but not Node.js would get no installs at all.
- Parallel installs — rejected, subprocess output would
  interleave and confuse the user.

## R8: Manager Precedence for Competing Managers

**Decision**: When multiple managers exist for the same
tool category, prefer the manager whose binary is the one
actually resolved by PATH (i.e., whichever appears first
in PATH wins). For install actions, prefer the manager
that currently manages the active binary if one exists;
otherwise prefer version managers over package managers
(nvm over Homebrew for Node.js).

**Rationale**: The binary that `exec.LookPath` returns is
the one the developer is actually using. Respecting PATH
order is the most honest representation of the developer's
environment.

**Precedence for Node.js install**:
1. nvm (if `NVM_DIR` set or `.nvm/` in path)
2. fnm (if `FNM_DIR` set or fnm detected)
3. mise (if mise active)
4. Homebrew (fallback)

**Precedence for Go install**:
1. goenv (if `GOENV_ROOT` set or `.goenv/` in path)
2. mise (if mise active for Go)
3. Homebrew (fallback)
4. Direct download (no managers detected)

**Alternatives considered**:
- Asking the user which manager to use — rejected, adds
  interactive prompts to a non-interactive command.
- Always preferring Homebrew — rejected, contradicts the
  spec requirement to respect existing managers.
