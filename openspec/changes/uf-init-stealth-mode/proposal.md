## Why

`uf init` scaffolds `.opencode/`, `openspec/`, `.specify/`, `.uf/`,
and `opencode.json`, and it unconditionally modifies tracked files
(`.gitignore`, `AGENTS.md`). For a developer who wants to use `uf`
locally to author specs and open PRs against an upstream project that
does not itself adopt unbound-force, every scaffolded file shows up in
`git status` and risks being committed or shipped in a PR unless the
developer manually reverts and hides it. This is the local-only
workflow described in discussion #523, and today it requires manual
post-init cleanup steps.

A first-class "stealth" (local-only) mode would reduce that workflow to
a single command: leave the working tree git-clean, leave tracked files
unmodified, and still let the pipeline commands operate normally.

## What Changes

Add a `--stealth` mode to `uf init` that:

- Scaffolds the tool-owned files so pipeline commands keep working.
- Leaves the working tree git-clean (no staged, tracked, or
  committed unbound-force artifacts).
- Leaves tracked files (`.gitignore`, `AGENTS.md`) unmodified —
  including any edits those files would otherwise receive from
  `uf init` itself or from tools that `uf init` installs.
- Makes scaffolded files present-but-invisible via a local-only git
  facility (`.git/info/exclude`) with no indirection required for path
  resolution. Out-of-tree storage + symlinks (#198) is explicitly
  deferred as a future consolidation point, not part of this change.
- Never leaves the working tree in a half-applied state, and provides a
  `--check` verification path plus a post-init summary (with
  machine-parseable output) that reports working-tree cleanliness.

## Capabilities

### New Capabilities
- `uf init --stealth`: a local-only init mode that scaffolds tool-owned
  files for pipeline use while keeping the working tree git-clean and
  tracked files unmodified.

### Modified Capabilities
- `uf init`: gains an opt-in stealth flag; the existing (non-stealth)
  behavior is unchanged.

### Removed Capabilities
- (none)

## Impact

- `internal/scaffold` (Options struct, `Run`, `ensureGitignore`,
  `ensureAGENTSmdPackSection`, `isToolOwned`, drift detection).
- `cmd/unbound-force` init command wiring for the new flag.
- `.gitignore`/`AGENTS.md` mutation paths and their interaction with
  tools installed during init.
- Test coverage: table-driven tests across `{normal, stealth}` modes
  must assert normal-mode behavior is byte-identical and stealth-mode
  never touches tracked files.

## Constitution Alignment

Assessed against the Unbound Force org constitution (v1.3.0).

### I. Autonomous Collaboration

**Assessment**: PASS

Stealth mode does not change how heroes communicate. Pipeline commands
continue to operate normally, so artifacts remain self-describing and
discoverable at their well-known locations. The change is scoped to
file placement and git visibility, not to inter-hero communication.

### II. Composability First

**Assessment**: PASS

This change directly strengthens composability: it lets `uf` be used
additively against a repository that does not adopt the framework,
without polluting that repository. It introduces no mandatory
dependencies and preserves standalone operation of every other hero.

### III. Observable Quality

**Assessment**: PASS

The change adds a post-init summary reporting working-tree cleanliness
and a `--check` verification path, making the git-clean invariant
observable rather than assumed. `--check` and the post-init summary
emit machine-parseable (JSON) output per Constitution III.

### IV. Testability

**Assessment**: PASS

The git-clean invariant is objectively verifiable (`git status
--porcelain` empty; tracked files byte-identical before/after). Tests
will be table-driven across `{normal, stealth}` modes in isolation
(`t.TempDir()`), with explicit assertions that normal mode is
unchanged. The coverage strategy (unit vs integration classification
and specific targets) is defined in design.md "Test Strategy" and
enforced by the existing coverage ratchet.

### V. Security by Default

**Assessment**: PASS

Stealth mode reduces the risk of accidentally committing `.uf/` local
config (which may hold embedding/model overrides and, plausibly,
secrets) into an upstream PR. This protects against accidental git
commit, not against local file disclosure. No new external dependencies
are introduced.
