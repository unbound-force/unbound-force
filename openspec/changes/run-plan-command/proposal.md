## Why

The Unbound Force pipeline (`/opsx-propose` -> `/uf.unleash` -> `/uf.finale`) is powerful for single changes but requires manual invocation for each work item. When executing a prioritized plan of multiple work items (e.g., a sprint backlog or a list of GitHub issues), the operator must manually sequence three commands per item, each in a fresh subagent context to prevent context leakage. This is tedious, error-prone (context bleeds between items), and violates the zero-waste principle by requiring repetitive human orchestration of a mechanical process.

A `/uf.run-pipeline-plan` command codifies this orchestration pattern: it reads a prioritized work plan, then for each item spawns fresh subagents for each pipeline phase, relays mandatory human confirmation gates through the `question` tool, and tracks progress via `todowrite`. This turns a multi-hour manual sequencing task into a single command invocation with human gates preserved.

## What Changes

- **New command**: `.opencode/commands/uf.run-pipeline-plan.md` — the canonical slash command definition
- **Embedded copy**: `internal/scaffold/assets/opencode/commands/uf.run-pipeline-plan.md` — byte-identical copy for `uf init` scaffold deployment
- **Drift test manifest**: Add `"opencode/commands/uf.run-pipeline-plan.md"` to `expectedAssetPaths` in `internal/scaffold/scaffold_test.go` (~line 153) so `TestEmbeddedAssets_MatchSource` validates the new asset

The command implements:
1. **Phase 0 (orchestrator context)**: Resolve input (plan file path or GitHub issue refs), derive change names/descriptions, present work-item table for human confirmation, flag non-mechanical items (investigations, policy decisions) for explicit resolution before pipeline runs.
2. **Per-item protocol (sequential)**: For each work item, spawn three fresh `general` subagents in sequence:
   - Subagent 1: `/opsx-propose` (spec artifacts)
   - Subagent 2: `/uf.unleash` (OpenSpec mode; Step 8 review council never skipped)
   - Subagent 3: `/uf.finale` (commit + PR with `Closes #N`)
   - Verify disk handoff between subagents; confirm return to `main` + clean tree between items.
3. **Gate relay protocol**: Subagents STOP and return `GATE:<name>` + exact question + full proposed text + options when encountering a mandatory human confirmation. The orchestrator relays through `question` and resumes the same subagent via `task_id`. Subagents are forbidden from invoking `question` directly.
4. **Guardrails**: Scaffold drift sync, explicit staging (never `git add -A`), no force-push (only `--force-with-lease`), PR targets parent repo head `jflowers:<branch>`, `Closes #N` in commit message and PR body, never merge the PR.
5. **Tracking**: `todowrite` one entry per work item, marked complete only after PR created and CI green.

## Capabilities

### New Capabilities
- `/uf.run-pipeline-plan`: Orchestrates a prioritized work plan through the full Unbound Force pipeline with fresh subagent isolation per phase and human gate relay

### Modified Capabilities
- `expectedAssetPaths` in `internal/scaffold/scaffold_test.go`: Add new command path to drift test manifest

### Removed Capabilities
- None

## Impact

- **New files (3)**:
  - `.opencode/commands/uf.run-pipeline-plan.md` — canonical command definition
  - `internal/scaffold/assets/opencode/commands/uf.run-pipeline-plan.md` — embedded copy (byte-identical)
  - Line addition to `internal/scaffold/scaffold_test.go` `expectedAssetPaths` (~line 153)
- **No changes to existing commands**: `/opsx-propose`, `/uf.unleash`, `/uf.finale` remain unchanged; `/uf.run-pipeline-plan` invokes them as-is via subagents
- **No changes to scaffold engine**: The embed.FS walk in `internal/scaffold/` automatically picks up new files under `assets/`; only the test manifest needs updating
- **CI impact**: `TestEmbeddedAssets_MatchSource` will validate the new asset; `TestAssetPaths_MatchExpected` will verify the manifest is current

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

The command communicates through well-defined artifacts: plan files (markdown), subagent return values (`GATE:<name>` protocol), and `todowrite` entries. Each subagent produces self-describing outputs (spec artifacts, commits, PRs) that any consumer can interpret without consulting the producing subagent. The orchestrator does not require synchronous interaction with subagents — it reads their return values from disk and resumes via `task_id`. The gate relay protocol (`GATE:<name>` + question + text + options) is a structured artifact exchange format, not ad-hoc messaging.

### II. Composability First

**Assessment**: PASS

`/uf.run-pipeline-plan` delivers value only when the existing pipeline commands (`/opsx-propose`, `/uf.unleash`, `/uf.finale`) are present — but it does not modify or require patching any of them. It invokes them as-is via subagents. A user who only needs single-item execution can continue using the individual commands directly. The command auto-detects input format (plan file vs. issue refs) without requiring manual configuration. No new mandatory dependencies are introduced.

### III. Observable Quality

**Assessment**: PASS

Every pipeline execution produces machine-parseable, traceable outputs: spec artifacts on disk, git commits with conventional messages, GitHub PRs with `Closes #N` links, and `todowrite` entries with completion status. The gate relay protocol produces structured return values (`GATE:<name>` + exact question + full text + options) that are auditable. The `todowrite` tracker provides a machine-readable progress log. All quality claims (e.g., "CI green") are backed by verifiable CI status checks.

### IV. Testability

**Assessment**: PASS

The command itself is a markdown file (not executable code), so testability applies to its guardrails and the scaffold drift test. The new `expectedAssetPaths` entry is validated by `TestEmbeddedAssets_MatchSource`, which verifies byte-identity between canonical and embedded copies. The command's gate protocol is testable by inspection: each subagent's return value is a structured string that can be pattern-matched. The per-item isolation (fresh subagent per phase) is testable by verifying no context leakage between items — a property enforced by the subagent spawn mechanism, not by the command itself.

### V. Security by Default

**Assessment**: PASS

The command enforces explicit staging (never `git add -A`/`git add .`), preventing accidental secret commits. It forbids `git push --force` (only `--force-with-lease` allowed), protecting against history rewrites. PRs target the parent repo's head branch with a user-scoped ref (`jflowers:<branch>`), preventing cross-repo injection. The gate relay protocol ensures no mandatory human confirmation is bypassed — subagents STOP and return, they do not auto-approve. Input validation (plan file existence check, issue ref parsing) prevents path traversal and injection.
