# Research: Externalize Tool Initialization

**Branch**: `027-externalize-tool-init` | **Date**: 2026-04-11

## R1: Sub-Tool Init Delegation Pattern (Existing)

**Question**: What is the established pattern for delegating
initialization to an external tool in `uf init`?

**Finding**: The `initSubTools()` function in `scaffold.go`
(lines 783–896) handles this. Two delegations exist today:

1. **Dewey** (lines 822–872):
   - Gate: `opts.LookPath("dewey")` + `os.Stat(".uf/dewey")`
   - Init: `opts.ExecCmd("dewey", "init")`
   - Post-init: `generateDeweySources()` + `opts.ExecCmd("dewey", "index")`
   - Result: `subToolResult{name: ".uf/dewey/", action: "initialized"}`
   - Force mode: re-index if workspace exists and `opts.Force`

2. **Replicator** (lines 876–889):
   - Gate: `opts.LookPath("replicator")` + `os.Stat(".uf/replicator")`
   - Init: `opts.ExecCmd("replicator", "init")`
   - Result: `subToolResult{name: ".uf/replicator/", action: "initialized"}`

Both follow the same pattern:
1. Check binary in PATH via `opts.LookPath()`
2. Check workspace directory existence via `os.Stat()`
3. If absent, run init command via `opts.ExecCmd()`
4. Report result as `subToolResult`
5. Errors are captured as warnings, not hard failures
   (Constitution Principle II — Composability First)

**Application**: The three new delegations (specify, openspec,
gaze) follow this exact pattern. The only variation is that
`openspec init` takes flags (`--tools opencode`) and the custom
schema deployment happens after `openspec init`.

**Status**: RESOLVED — use existing pattern.

---

## R2: Setup Install Pattern (Homebrew vs uv)

**Question**: What is the pattern for adding a new tool
installation step to `uf setup`?

**Finding**: `setup.go` has two installation patterns:

1. **Homebrew pattern** (used by Gaze, Mx F, Replicator,
   Dewey, Ollama, GitHub CLI):
   - Check `opts.LookPath("binary")`
   - If found: `stepResult{action: "already installed"}`
   - If dry-run: return dry-run result
   - If no Homebrew: return skipped with download link
   - Run `brew install unbound-force/tap/<binary>`
   - Return installed or failed

2. **npm pattern** (used by OpenSpec CLI):
   - Check `opts.LookPath("openspec")`
   - If found: `stepResult{action: "already installed"}`
   - If dry-run: return dry-run result
   - Run `npm install -g @fission-ai/openspec@latest`
   - Return installed or failed

**New pattern needed**: `uv tool install` for the `specify` CLI.
This is a Python tool, so the installation chain is:
1. Ensure `uv` is available (install via Homebrew or curl)
2. Install `specify` via `uv tool install specify-cli`

This is analogous to the Node.js → OpenSpec chain (steps 5→6):
first ensure the runtime/package manager, then install the tool.

**Application**:
- `installUV()`: Homebrew pattern (`brew install uv`), with
  curl fallback (`curl -LsSf https://astral.sh/uv/install.sh | sh`)
- `installSpecify()`: `uv tool install specify-cli`, gated by
  `uv` availability (similar to how OpenSpec is gated by Node.js)

**Status**: RESOLVED — Homebrew + curl for uv, uv tool install
for specify.

---

## R3: Embedded Asset Removal Impact

**Question**: What happens when embedded assets are removed from
`internal/scaffold/assets/`? What tests break?

**Finding**: The following test infrastructure is affected:

1. **`expectedAssetPaths`** (line 87): Canonical list of all
   embedded assets. Currently 55 entries. Must remove:
   - 6 specify/templates/*.md entries
   - 1 specify/config.yaml entry
   - 5 specify/scripts/bash/*.sh entries
   - 1 openspec/config.yaml entry
   Total: 13 entries removed → 42 entries remaining

2. **`TestAssetPaths_MatchExpected`** (line 156): Verifies
   embedded assets match `expectedAssetPaths`. Will pass after
   list update.

3. **`TestRun_CreatesFiles`** (line 231): Asserts
   `len(result.Created) == len(expectedAssetPaths)`. Will pass
   after list update.

4. **`TestRun_SkipsExisting`** (line 270): Same pattern.

5. **`TestRun_ForceOverwrites`** (line 326): Same pattern.

6. **`TestCanonicalSources_AreEmbedded`** (line 940): Walks
   `.opencode/`, `.specify/`, `openspec/` canonical directories
   and checks each file is either embedded or in
   `knownNonEmbeddedFiles`. After this change:
   - `.specify/` files are created by `specify init`, not
     embedded → add to `knownNonEmbeddedFiles` OR remove
     `.specify/` from `canonicalDirs` walk
   - `openspec/config.yaml` is created by `openspec init` →
     remove from `canonicalDirs` standalone check

7. **`knownNonEmbeddedFiles`** (line 876): Must add entries for
   files created by `gaze init` (already there: `gaze-reporter.md`,
   `gaze.md`) and potentially `specify init` files.

8. **`mapAssetPath`** tests: The `specify/` prefix mapping
   remains valid even with fewer files. No change needed.

9. **`isToolOwned`** tests: No change — specify files were
   never tool-owned (they were user-owned templates/scripts).

**Application**: Update `expectedAssetPaths` (remove 13),
update `TestCanonicalSources_AreEmbedded` to handle the new
reality where `.specify/` files come from `specify init`.

**Status**: RESOLVED — mechanical test updates.

---

## R4: OpenSpec Custom Schema Deployment Order

**Question**: How should the custom OpenSpec schema be deployed
when `openspec init` creates the base structure first?

**Finding**: Currently, `uf init` deploys all OpenSpec files
from embedded assets in a single pass (the `fs.WalkDir` in
`Run()`). After this change:

1. `openspec init --tools opencode` creates the base structure:
   - `openspec/config.yaml`
   - `openspec/specs/` (empty directory)
   - `openspec/changes/` (empty directory)
   - `openspec/schemas/` (base schemas)

2. `uf init`'s embedded asset walk deploys the custom schema:
   - `openspec/schemas/unbound-force/schema.yaml`
   - `openspec/schemas/unbound-force/templates/proposal.md`
   - `openspec/schemas/unbound-force/templates/spec.md`
   - `openspec/schemas/unbound-force/templates/design.md`
   - `openspec/schemas/unbound-force/templates/tasks.md`

**Ordering concern**: The embedded asset walk happens in
`Run()` BEFORE `initSubTools()`. This means the custom schema
files would be deployed before `openspec init` creates the base
structure. This is fine because:
- `os.MkdirAll` creates parent directories as needed
- The custom schema files don't depend on `openspec/config.yaml`
- `openspec init` is idempotent — it won't overwrite the custom
  schema files that `uf init` already deployed

However, there's a subtlety: `openspec init` might create its
own `openspec/config.yaml` that differs from the one we
currently embed. After this change, we remove the embedded
`openspec/config.yaml`, so `openspec init` creates it. The
custom schema is deployed on top of whatever `openspec init`
creates. This is the desired behavior per FR-007.

**Application**: The ordering is:
1. `Run()` walks embedded assets → deploys custom schema files
2. `initSubTools()` runs `openspec init` → creates base
   structure (config, empty dirs)
3. Custom schema files already exist from step 1 → `openspec
   init` doesn't overwrite them

Wait — this ordering means the custom schema is deployed
BEFORE `openspec init` creates the `openspec/` directory.
That's fine because `os.MkdirAll` in the asset walk creates
the directory. But it means `openspec init` might see an
existing `openspec/` directory and skip initialization
(idempotent check).

**Revised approach**: Move the openspec init delegation to
happen BEFORE the custom schema deployment. This means:
1. `initSubTools()` runs `openspec init` → creates base
   structure
2. `Run()` walks embedded assets → deploys custom schema on top

But `initSubTools()` runs AFTER the asset walk in `Run()`.
We can't easily reorder without restructuring `Run()`.

**Simplest approach**: Keep the current ordering. The
`initSubTools()` openspec delegation checks for `openspec/`
directory existence. If the embedded asset walk already created
`openspec/schemas/unbound-force/`, then `openspec/` exists and
the delegation is skipped. This means `openspec init` would
NOT run if the custom schema was deployed first.

**Solution**: Gate the openspec delegation on `openspec/config.yaml`
existence (not `openspec/` directory existence). The custom
schema creates `openspec/schemas/` but not `openspec/config.yaml`.
So the gate becomes: "if `openspec/config.yaml` does not exist,
run `openspec init`". This way:
1. Asset walk deploys custom schema → creates `openspec/schemas/`
2. `initSubTools()` checks `openspec/config.yaml` → absent
3. Runs `openspec init --tools opencode` → creates config.yaml
   and base structure
4. Custom schema already exists → not overwritten

**Status**: RESOLVED — gate on `openspec/config.yaml`, not
`openspec/` directory.

---

## R5: Setup Step Count and Flow

**Question**: What is the current step flow and how does it
change?

**Finding**: Current 12-step flow in `setup.go`:

| Step | Name | Function |
|------|------|----------|
| 1 | OpenCode | `installOpenCode` |
| 2 | Gaze | `installGaze` |
| 3 | Mx F | `installMxF` |
| 4 | GitHub CLI | `installGH` |
| 5 | Node.js | `ensureNodeJS` |
| 6 | OpenSpec CLI | `installOpenSpec` |
| 7 | Replicator | `installReplicator` |
| 8 | Replicator setup | `runReplicatorSetup` |
| 9 | Ollama | `installOllama` |
| 10 | Dewey | `installDewey` |
| 11 | golangci-lint | `installGolangciLint` |
| 12 | govulncheck | `installGovulncheck` |

New 15-step flow (adding uv, specify, and specify after
Node.js/OpenSpec group):

| Step | Name | Function | Change |
|------|------|----------|--------|
| 1 | OpenCode | `installOpenCode` | unchanged |
| 2 | Gaze | `installGaze` | unchanged |
| 3 | Mx F | `installMxF` | unchanged |
| 4 | GitHub CLI | `installGH` | unchanged |
| 5 | Node.js | `ensureNodeJS` | unchanged |
| 6 | OpenSpec CLI | `installOpenSpec` | unchanged |
| 7 | uv | `installUV` | **NEW** |
| 8 | Specify CLI | `installSpecify` | **NEW** |
| 9 | Replicator | `installReplicator` | renumbered |
| 10 | Replicator setup | `runReplicatorSetup` | renumbered |
| 11 | Ollama | `installOllama` | renumbered |
| 12 | Dewey | `installDewey` | renumbered |
| 13 | golangci-lint | `installGolangciLint` | renumbered |
| 14 | govulncheck | `installGovulncheck` | renumbered |
| 15 | Embedding model | (part of installDewey) | renumbered |

Wait — step 15 doesn't exist separately. The embedding model
pull is part of `installDewey()`. Let me recount:

Actually, the current code has exactly 12 `fmt.Fprintf` step
lines (`[1/12]` through `[12/12]`). Adding 2 new steps makes
it 14, not 15. Let me recount the spec:

The spec says "2 new steps, 12→15" but that's 12+3=15. Looking
at the spec more carefully: it says `uv` + `specify`
installation = 2 new steps. 12+2=14. But the spec header says
"12→15". This is likely a spec error — the actual count should
be 12+2=14.

**Revised step count**: 14 steps (12 existing + 2 new: uv,
specify).

**Placement**: After OpenSpec CLI (step 6), before Replicator
(currently step 7). This groups all package-manager-dependent
tools together: Node.js → OpenSpec, uv → Specify.

| Step | Name | Function |
|------|------|----------|
| 1/14 | OpenCode | `installOpenCode` |
| 2/14 | Gaze | `installGaze` |
| 3/14 | Mx F | `installMxF` |
| 4/14 | GitHub CLI | `installGH` |
| 5/14 | Node.js | `ensureNodeJS` |
| 6/14 | OpenSpec CLI | `installOpenSpec` |
| 7/14 | uv | `installUV` |
| 8/14 | Specify CLI | `installSpecify` |
| 9/14 | Replicator | `installReplicator` |
| 10/14 | Replicator setup | `runReplicatorSetup` |
| 11/14 | Ollama | `installOllama` |
| 12/14 | Dewey | `installDewey` |
| 13/14 | golangci-lint | `installGolangciLint` |
| 14/14 | govulncheck | `installGovulncheck` |

**Status**: RESOLVED — 14 steps (12 + 2 new).

---

## R6: Gaze Init Delegation

**Question**: What does `gaze init` create, and how does it
interact with the existing `knownNonEmbeddedFiles` list?

**Finding**: `gaze init` creates:
- `.opencode/agents/gaze-reporter.md`
- `.opencode/command/gaze.md`

These are already in `knownNonEmbeddedFiles` (lines 878, 889).
The `TestCanonicalSources_AreEmbedded` test already handles
them correctly — they are excluded from the "must be embedded"
check.

The Gaze delegation gate should check for the agent file
existence (not a directory), since Gaze doesn't create a
workspace directory like Dewey or Replicator. The idempotency
check is: "if `.opencode/agents/gaze-reporter.md` exists, skip
`gaze init`".

There's also `gaze-test-generator.md` in the exclusion list
(line 879) and `gaze-fix.md` (line 890). These may also be
created by `gaze init`. The gate should check for the primary
agent file (`gaze-reporter.md`) as the sentinel.

**Status**: RESOLVED — gate on `gaze-reporter.md` existence.

---

## R7: Specify Init Behavior

**Question**: What does `specify init` create, and is it
non-interactive?

**Finding**: Based on the spec assumptions (lines 307–309),
`specify init` creates the `.specify/` directory structure:
- `.specify/config.yaml`
- `.specify/templates/` (6 Markdown templates)
- `.specify/scripts/bash/` (5 bash scripts)
- `.specify/memory/` (empty, for constitution)

The command is assumed to be non-interactive when run without
arguments. The gate is: "if `.specify/` directory exists, skip
`specify init`".

The 12 files currently embedded in `internal/scaffold/assets/specify/`
are:
1. `specify/config.yaml`
2. `specify/templates/agent-file-template.md`
3. `specify/templates/checklist-template.md`
4. `specify/templates/constitution-template.md`
5. `specify/templates/plan-template.md`
6. `specify/templates/spec-template.md`
7. `specify/templates/tasks-template.md`
8. `specify/scripts/bash/check-prerequisites.sh`
9. `specify/scripts/bash/common.sh`
10. `specify/scripts/bash/create-new-feature.sh`
11. `specify/scripts/bash/setup-plan.sh`
12. `specify/scripts/bash/update-agent-context.sh`

All 12 will be removed from embedded assets. After removal,
the `specify/` directory under `assets/` should be deleted
entirely.

**Status**: RESOLVED — remove all 12 files and the `specify/`
directory.

---

## R8: OpenSpec Init Behavior

**Question**: What does `openspec init --tools opencode` create?

**Finding**: Based on the spec (lines 297–301), `openspec init
--tools opencode` creates:
- `openspec/config.yaml` (base config)
- `openspec/specs/` (empty directory for specs)
- `openspec/changes/` (empty directory for changes)
- `openspec/schemas/` (base schema directory)
- Various OpenCode-specific files (skills, commands)

The `--tools opencode` flag bypasses the interactive tool
selection prompt (spec assumption, line 312).

Currently embedded OpenSpec files:
1. `openspec/config.yaml` — REMOVE (created by `openspec init`)
2. `openspec/schemas/unbound-force/schema.yaml` — KEEP
3. `openspec/schemas/unbound-force/templates/proposal.md` — KEEP
4. `openspec/schemas/unbound-force/templates/spec.md` — KEEP
5. `openspec/schemas/unbound-force/templates/design.md` — KEEP
6. `openspec/schemas/unbound-force/templates/tasks.md` — KEEP

After removal: 1 file removed, 5 files remain.

The `knownAssetPrefixes` in `scaffold.go` (line 251) includes
`"openspec/"` — this remains valid since the custom schema
files still use the `openspec/` prefix.

The empty directory creation in `Run()` (lines 200–209) creates
`openspec/specs/` and `openspec/changes/`. After this change,
`openspec init` creates these directories. The empty directory
creation code can be removed or kept (it's idempotent via
`os.MkdirAll`). Keeping it is safer — if `openspec` is not
installed, the directories still get created for manual use.

**Decision**: Keep the empty directory creation code. It's
harmless when `openspec init` has already run, and useful when
`openspec` is not installed.

**Status**: RESOLVED — remove `openspec/config.yaml` only,
keep custom schema and empty dir creation.

---

## R9: Doctor Checks Impact

**Question**: Does `uf doctor` need updates for this change?

**Finding**: `uf doctor` checks for `.specify/` directory
existence (lines 466–480 of `checks.go`). This check remains
valid — it doesn't care whether the directory was created by
embedded assets or `specify init`.

No doctor changes are needed. The doctor checks are
existence-based, not provenance-based.

**Status**: RESOLVED — no doctor changes needed.

---

## R10: UV Installation Method

**Question**: How should `uv` be installed across platforms?

**Finding**: `uv` (the Python package manager by Astral) can be
installed via:
- **Homebrew**: `brew install uv` (macOS, Linux with Homebrew)
- **curl**: `curl -LsSf https://astral.sh/uv/install.sh | sh`
  (any Unix system)
- **pip**: `pip install uv` (if Python is available)

The Homebrew pattern is preferred (consistent with other tools).
The curl fallback follows the OpenCode installation pattern
(lines 290–301 of `setup.go`): requires `--yes` flag or TTY
confirmation.

**Application**: `installUV()` follows the Homebrew-first
pattern with curl fallback. The curl fallback requires the same
`--yes`/TTY guard as OpenCode's curl install.

**Status**: RESOLVED — Homebrew first, curl fallback with
interactive guard.
