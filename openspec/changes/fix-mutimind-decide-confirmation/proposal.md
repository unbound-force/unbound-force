## Why

The `muti-mind-po.md` agent file contains a `go run cmd/mutimind/main.go decide ...` command block (lines 70-72) that fires acceptance decisions (`accept`/`reject`/`conditional`) without any AskUserQuestion confirmation gate. Acceptance decisions are irreversible governance artifacts that affect downstream workflow -- rejecting a backlog item or conditionally accepting it has material consequences on team priorities and sprint planning.

The "Interactive Approval" rule at line 60 covers story generation (`go run ... add`) but has no parallel for the `decide` block. This gap was identified in issue #351 as part of the root cause analysis from the parent audit (issue #346), which uncovered agents executing irreversible external actions without human confirmation.

The related `askuser-tool-confirmations` change (completed) hardened 15 interaction points across three slash commands (`/review-pr`, `/address-feedback`, `/triage-issue`), but agent files were out of scope. This change closes the remaining gap in the Muti-Mind PO agent.

## What Changes

Add a mandatory AskUserQuestion confirmation gate immediately before the `go run cmd/mutimind/main.go decide` command block in the Acceptance Authority section of `muti-mind-po.md`. The gate presents the target backlog item, proposed decision, and rationale summary for human review before execution.

## Capabilities

### New Capabilities
- None (no new functionality)

### Modified Capabilities
- `Acceptance Authority`: Acceptance decision CLI invocation now requires explicit human confirmation via AskUserQuestion before execution

### Removed Capabilities
- None

## Impact

**Files modified** (1 agent file):
- `.opencode/agents/muti-mind-po.md`

**Behavioral change**: The acceptance decision workflow gains a mandatory confirmation step. The agent must present the decision details (backlog item ID, decision type, rationale summary) and receive user confirmation before invoking the Go CLI backend. This matches the pattern already established by the Interactive Approval rule for story generation at line 60.

**No scaffold drift**: Unlike slash commands, agent files under `.opencode/agents/` do not have corresponding scaffold assets under `internal/scaffold/assets/`. No sync step is required.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

This change modifies agent prompt instructions (a persona markdown file), not inter-hero artifact exchange. The `acceptance-decision` JSON artifact schema and its structure remain unchanged. Only the human confirmation step before CLI invocation is affected.

### II. Composability First

**Assessment**: N/A

No hero dependencies are introduced or modified. The Muti-Mind PO agent remains independently usable. The AskUserQuestion tool is a built-in OpenCode capability, not a hero dependency.

### III. Observable Quality

**Assessment**: N/A

No machine-parseable outputs or provenance metadata are affected. The `acceptance-decision` artifact schema, its fields (`decision`, `rationale`, `criteria_met`, `criteria_failed`), and the Go CLI backend behavior remain unchanged.

### IV. Testability

**Assessment**: N/A

No testable components are modified. This is an agent instruction change in a markdown file. There are no scaffold assets to keep in sync, so no drift detection tests apply.
<!-- scaffolded by uf vdev -->
