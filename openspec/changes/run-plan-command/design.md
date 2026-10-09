## Context

The Unbound Force pipeline consists of three sequential commands — `/opsx-propose` (spec artifacts), `/uf.unleash` (autonomous implementation + review), `/uf.finale` (commit + PR) — each designed for single-change execution. When executing a prioritized plan of multiple work items, the operator must manually invoke this three-command sequence per item, ensuring each invocation runs in a fresh subagent context to prevent context leakage between items. This is mechanical, repetitive, and error-prone.

The `/uf.run-pipeline-plan` command codifies this orchestration as a reusable slash command. It reads a prioritized work plan, then for each item spawns fresh `general` subagents for each pipeline phase, relays mandatory human confirmation gates through the `question` tool, and tracks progress via `todowrite`.

This design aligns with the constitution's Autonomous Collaboration principle: subagents communicate through disk artifacts (spec files, git commits, PRs), not runtime coupling. The gate relay protocol (`GATE:<name>` + question + text + options) is a structured artifact exchange format that preserves the human-in-the-loop discipline without violating subagent isolation.

## Goals / Non-Goals

### Goals
- Provide a single `/uf.run-pipeline-plan` command that executes a prioritized work plan through the full pipeline
- Enforce fresh subagent context per pipeline phase (no context leakage between items or phases)
- Relay all mandatory human confirmation gates through the orchestrator's `question` tool — subagents MUST NOT invoke `question` directly
- Support two input modes: plan file path OR space-separated GitHub issue refs (`#NNN` or URLs)
- Auto-derive kebab-case change names and descriptions from GitHub issue metadata (Phase 0)
- Flag non-mechanical items (investigations, policy decisions) for explicit human resolution before pipeline execution
- Verify disk handoff between subagents (artifacts on disk, branch checked out) before proceeding
- Confirm return to `main` + clean working tree between items
- Enforce all existing guardrails: scaffold drift sync, explicit staging, no force-push, PR targeting, `Closes #N` linkage
- Track progress via `todowrite` with one entry per work item

### Non-Goals
- Parallel execution of work items (items run sequentially to maintain isolation and simplify gate relay)
- Modifying existing commands (`/opsx-propose`, `/uf.unleash`, `/uf.finale`) — `/uf.run-pipeline-plan` invokes them as-is
- Automatic merge of PRs (the PR stays open for human review, per `/uf.finale` behavior)
- Supporting plan formats other than numbered markdown lists or GitHub issue refs
- Automatic resolution of investigation/policy items (these are surfaced for human decision, not auto-executed)

## Decisions

### D1: Fresh subagent per phase (not per item)

**Decision**: Each pipeline phase (`/opsx-propose`, `/uf.unleash`, `/uf.finale`) runs in a fresh `general` subagent. Three subagents per work item, not one.

**Rationale**: Context leakage between phases is the primary failure mode. `/opsx-propose` produces spec artifacts that `/uf.unleash` reads from disk; if they share context, the unleash subagent may skip reading artifacts (because it "already knows" them from the propose context). Fresh subagents force disk-based handoff, which is the constitutionally correct artifact-based communication pattern. This also means each subagent starts with a clean context window, reducing the risk of context-compression artifacts.

**Alternative considered**: Single subagent per item running all three phases sequentially. Rejected because context leakage between phases would undermine the disk-handoff verification step.

### D2: Gate relay protocol (GATE:<name> return value)

**Decision**: When a subagent encounters a mandatory human confirmation gate, it MUST STOP execution and return a structured value: `GATE:<name>` + the exact question + full proposed text + options. The orchestrator then invokes `question` and resumes the same subagent via `task_id` with the human's answer.

**Rationale**: Subagents are forbidden from invoking `question` directly because the orchestrator owns the human interaction surface. This prevents subagents from racing for user input, ensures consistent gate formatting, and makes the gate relay auditable (each gate has a name, question, text, and options). The `task_id` resume mechanism ensures the same subagent continues with the human's answer in its context.

**Alternative considered**: Subagents invoke `question` directly. Rejected because it violates the orchestrator's role as the single point of human interaction, makes gates harder to track, and risks context pollution from multiple subagents awaiting user input.

### D3: Phase 0 orchestrator context (no subagents)

**Decision**: Phase 0 runs in the orchestrator context (no subagents). It resolves input, derives change names/descriptions from GitHub issues via `gh issue view`, presents the work-item table for confirmation, and flags non-mechanical items.

**Rationale**: Phase 0 is purely investigative and confirmatory — it does not produce spec artifacts or execute pipeline phases. Running it in the orchestrator context avoids the overhead of subagent spawn for what is essentially input parsing and human confirmation. The `gh issue view` calls are lightweight and do not risk context leakage.

### D4: Plan file contract (numbered list format)

**Decision**: A plan file is a numbered markdown list where each item carries: issue refs (`#NNN`), a kebab-case `Change:` name, and an optional one-line description. Non-mechanical items (investigations, policy decisions) are marked with a tag (e.g., `[INVESTIGATION]` or `[POLICY]`) and surfaced for explicit human resolution.

**Rationale**: The numbered list format is human-readable, diff-friendly, and simple to parse. The `Change:` name convention matches the kebab-case naming used by `/opsx-propose`. Marking non-mechanical items prevents the pipeline from auto-executing work that requires human judgment (e.g., "investigate why CI is flaky" is not a implementable change).

**Alternative considered**: YAML/JSON plan format. Rejected because markdown is more readable in PR descriptions and GitHub issue bodies, and the parsing requirements are simple enough for line-by-line reading.

### D5: Scaffold drift sync enforcement

**Decision**: The canonical command file (`.opencode/commands/uf.run-pipeline-plan.md`) and the embedded copy (`internal/scaffold/assets/opencode/commands/uf.run-pipeline-plan.md`) MUST be byte-identical. The `expectedAssetPaths` list in `internal/scaffold/scaffold_test.go` MUST include the new path. `TestEmbeddedAssets_MatchSource` will fail if they diverge.

**Rationale**: This follows the established pattern for all UF commands (10 existing commands follow this same dual-file pattern). The drift test is the safety net that prevents the canonical and embedded copies from diverging silently. This is the "agent change requires a spec" case from AGENTS.md — adding a new command is a structural change that touches both the canonical location and the scaffold assets.

### D6: Sequential item execution (not parallel)

**Decision**: Work items execute sequentially, one at a time. Each item completes its full three-phase pipeline before the next item begins.

**Rationale**: Sequential execution simplifies gate relay (only one subagent awaiting user input at a time), ensures clean tree verification between items, and prevents branch conflicts (each item creates its own `opsx/<name>` branch). Parallel execution would require worktree isolation and complex merge coordination, which is out of scope for this change.

### D7: Input detection (file vs. issue refs)

**Decision**: The orchestrator detects input type by checking if `$ARGUMENTS` is an existing file path. If it is, parse as a plan file. Otherwise, treat as a space-separated list of GitHub issue numbers (`#NNN`) or URLs.

**Rationale**: File existence check is a simple, unambiguous discriminator. This avoids requiring a flag (e.g., `--file` vs `--issues`) and matches the principle of least surprise — if you pass a file path, it reads the file; if you pass issue refs, it resolves them.

## Risks / Trade-offs

### R1: Subagent spawn overhead

**Risk**: Spawning three subagents per work item introduces latency (subagent initialization, context loading). For a plan with 10 items, that is 30 subagent spawns.

**Mitigation**: Subagent spawns are fast (seconds, not minutes). The sequential execution model means the total wall-clock time is dominated by pipeline execution (spec generation, implementation, CI), not subagent overhead. The isolation benefit outweighs the latency cost.

### R2: Gate relay complexity

**Risk**: The `GATE:<name>` return protocol requires subagents to recognize mandatory gates and format their return values correctly. A malformed gate return could cause the orchestrator to hang or skip a gate.

**Mitigation**: The command definition MUST include explicit examples of gate return formatting. The orchestrator MUST validate gate returns (check for `GATE:` prefix, required fields) before invoking `question`. If a gate return is malformed, the orchestrator MUST stop and report the error rather than guessing.

### R3: Context compression during long plans

**Risk**: For plans with many items, the orchestrator's context may compress, losing track of which items are complete.

**Mitigation**: The `todowrite` tracker is the source of truth for item completion, not the orchestrator's context. After each item, the orchestrator MUST read `todowrite` to determine the next incomplete item. This makes the command resumable — if the orchestrator's context is lost, a fresh invocation can read `todowrite` and resume from the next incomplete item.

### R4: Scaffold drift test failure on incomplete implementation

**Risk**: If the implementation adds the canonical file but forgets to update `expectedAssetPaths`, or vice versa, `TestEmbeddedAssets_MatchSource` will fail in CI.

**Mitigation**: The tasks.md MUST list the scaffold test manifest update as a task. The implementation MUST run `make test` before marking complete. This is the established pattern for all command additions.

### R5: Non-mechanical item handling

**Risk**: An item marked `[INVESTIGATION]` might be accidentally executed by the pipeline if the marker is not parsed correctly.

**Mitigation**: Phase 0 MUST present all flagged items to the human for explicit resolution before the pipeline runs. The human MUST confirm that flagged items are either resolved (and can be removed from the plan) or converted to mechanical items (with a clear change name and description). The pipeline MUST NOT auto-execute flagged items.
