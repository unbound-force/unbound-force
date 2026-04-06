# Contract: Setup — Replicator Installation

**Spec**: 024-replicator-migration | **FR**: FR-001 through FR-006

## Scope

Defines the behavioral contract for `uf setup` after the
Replicator migration. Covers installation, setup delegation,
bun removal, OpenSpec CLI changes, and step count.

## Step Flow (12 steps)

```text
[1/12] OpenCode        — installOpenCode (unchanged)
[2/12] Gaze            — installGaze (unchanged)
[3/12] Mx F            — installMxF (unchanged)
[4/12] GitHub CLI      — installGH (unchanged)
[5/12] Node.js         — ensureNodeJS (unchanged)
[6/12] OpenSpec CLI    — installOpenSpec (MODIFIED: npm-only)
[7/12] Replicator      — installReplicator (NEW)
[8/12] Replicator setup — runReplicatorSetup (NEW)
[9/12] Ollama          — installOllama (unchanged)
[10/12] Dewey          — installDewey (unchanged)
[11/12] golangci-lint  — installGolangciLint (unchanged)
[12/12] govulncheck    — installGovulncheck (unchanged)
```

## installReplicator Behavior

| Condition | Action | Result |
|-----------|--------|--------|
| `replicator` in PATH | Skip | `{action: "already installed"}` |
| DryRun + Homebrew | Report | `{action: "dry-run", detail: "Would install: brew install unbound-force/tap/replicator"}` |
| DryRun + no Homebrew | Report | `{action: "dry-run", detail: "Would install: download from GitHub releases"}` |
| No Homebrew | Skip | `{action: "skipped", detail: "Homebrew not available. Download from https://github.com/unbound-force/replicator/releases"}` |
| Homebrew available | Install | `brew install unbound-force/tap/replicator` |
| Install succeeds | Report | `{action: "installed", detail: "via Homebrew"}` |
| Install fails | Report | `{action: "failed", detail: "brew install failed"}` |

## runReplicatorSetup Behavior

| Condition | Action | Result |
|-----------|--------|--------|
| DryRun | Report | `{action: "dry-run", detail: "Would run: replicator setup"}` |
| Not interactive + no --yes | Skip | `{action: "skipped", detail: "interactive — run replicator setup manually or use --yes"}` |
| `replicator setup` succeeds | Report | `{action: "completed"}` |
| `replicator setup` fails | Report | `{action: "failed", detail: "replicator setup failed"}` |

## installOpenSpec Behavior (Modified)

| Condition | Action | Result |
|-----------|--------|--------|
| `openspec` in PATH | Skip | `{action: "already installed"}` |
| DryRun | Report | `{action: "dry-run", detail: "Would install: npm install -g @fission-ai/openspec@latest"}` |
| npm install succeeds | Report | `{action: "installed", detail: "via npm"}` |
| npm install fails | Report | `{action: "failed", detail: "npm install failed — ..."}` |

Note: No bun preference. npm is the only install method.

## Removed Functions

The following functions MUST be deleted from `setup.go`:

- `ensureBun()` — bun is no longer a prerequisite
- `installSwarmPlugin()` — replaced by `installReplicator()`
- `runSwarmSetup()` — replaced by `runReplicatorSetup()`
- `initializeHive()` — `.hive/` creation moves to `uf init`
  via `replicator init` delegation in `initSubTools()`
- `const swarmForkSource` — no longer needed

## Removed Step Logic

The Node.js-dependent block (lines 195–231) that gates steps
6–10 on `nodeAvailable` MUST be restructured:

- OpenSpec CLI still depends on Node.js (npm)
- Replicator does NOT depend on Node.js
- Replicator steps run unconditionally (after Node.js step)

## uf init Removal from Setup

Step 15 (`runUnboundInit`) is removed from setup. Rationale:
setup is per-machine; init is per-repo. The Dewey pattern
already handles per-repo initialization in `initSubTools()`
(called by `uf init`). Replicator init follows the same
pattern. Users run `uf init` separately after `uf setup`.

## Embedding Model Note

The completion message referencing "Swarm and Dewey" MUST be
updated to reference "Replicator and Dewey" (or just "Dewey"
since Replicator may not use the same embedding model).
