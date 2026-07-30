## Context

The Muti-Mind PO agent (`muti-mind-po.md`) has two CLI invocation points that execute irreversible actions:

1. **Story generation** (line 60): Protected by an "Interactive Approval" rule requiring user confirmation before `go run cmd/mutimind/main.go add`.
2. **Acceptance decisions** (lines 69-72): Unprotected. The `go run cmd/mutimind/main.go decide` command fires without confirmation.

The `askuser-tool-confirmations` change hardened 15 interaction points across three slash commands but was scoped to commands only. This change addresses the remaining gap in the agent file.

## Goals / Non-Goals

### Goals
- Add a mandatory AskUserQuestion confirmation gate before acceptance decision CLI execution
- Follow the established pattern from the existing Interactive Approval rule at line 60
- Present decision details (item ID, decision type, rationale) for human review
- Maintain consistency with the AskUserQuestion patterns established in the `askuser-tool-confirmations` change

### Non-Goals
- Modifying the Go CLI backend (`cmd/mutimind/main.go`) -- behavior unchanged
- Changing the `acceptance-decision` artifact schema
- Adding confirmation gates to other agent files (scope limited to `muti-mind-po.md`)
- Converting the existing story generation approval to AskUserQuestion (that could be a separate change)

## Decisions

### D1: Confirmation gate placement

Add the AskUserQuestion gate in the "Acceptance Authority" section (between lines 68 and 69), immediately before the CLI command block. This mirrors how the Interactive Approval rule at line 60 sits before the `add` command.

**Rationale**: Placing the gate directly above the CLI invocation block makes the requirement visually and logically adjacent to the action it guards. This is the same pattern used by line 60 for story generation.

### D2: AskUserQuestion options

Use structured options: `["Confirm decision", "Abort"]`.

**Rationale**: The issue explicitly specifies these two options. Binary choice is appropriate because the decision parameters (item, verdict, rationale) are already determined before this gate. The user's role is to approve or reject the action, not to modify parameters at this point.

### D3: Information presented before confirmation

The agent MUST present to the user before the AskUserQuestion prompt:
- The target backlog item identifier (`BI-NNN`)
- The proposed decision (`accept`, `reject`, or `conditional`)
- A summary of the rationale

**Rationale**: The user needs sufficient context to make an informed confirmation. Presenting just "Confirm decision?" without context would be a rubber-stamp gate with no real safety value.

### D4: No scaffold asset sync required

Agent files under `.opencode/agents/` do not have corresponding copies under `internal/scaffold/assets/`. Unlike slash commands (which require byte-identical sync), this change only touches a single file.

**Rationale**: Confirmed by searching `internal/scaffold/assets/` -- no `muti-mind-po.md` scaffold asset exists. The proposal's constitution alignment (section IV) already notes this.

## Risks / Trade-offs

### R1: Additional friction in automated workflows

When Muti-Mind runs in swarm mode (`execution_mode: swarm`), the confirmation gate will block autonomous execution of acceptance decisions. This is the intended behavior -- acceptance decisions are governance artifacts that should require human oversight even in automated pipelines. If fully autonomous acceptance is needed in the future, it should be gated by the execution mode flag, not by removing the confirmation.

### R2: Asymmetric confirmation patterns

The existing Interactive Approval at line 60 uses conversational language ("present the proposed stories to the user and ask for their confirmation"), while this change uses the structured AskUserQuestion tool. This creates a stylistic inconsistency within the same agent file. Converting line 60 to AskUserQuestion is explicitly a non-goal to keep this change minimal, but should be considered as follow-up work.
<!-- scaffolded by uf vdev -->
