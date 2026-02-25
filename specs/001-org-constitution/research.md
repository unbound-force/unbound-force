# Research: Unbound Force Organization Constitution

**Spec**: [spec.md](spec.md) | **Date**: 2026-02-25

## Research Task 1: Alignment Agent Design Patterns

### Decision

Adapt the Gaze reviewer agent pattern for the constitution alignment agent. The agent operates as a subagent (`mode: subagent`) with read-only tool permissions, uses low temperature (0.1) for deterministic output, and produces structured findings using the same `### [SEVERITY] Finding Title` format used by The Divisor prototype agents.

### Rationale

The Gaze reviewer agents (`reviewer-guard.md`, `reviewer-architect.md`, `reviewer-adversary.md`) provide a proven pattern for structured document comparison and analysis:

- **Read-only**: Reviewers have no write/edit/bash tools. The alignment agent should be read-only too -- it reads both constitutions and reports findings, never modifies them.
- **Subagent mode**: Reviewers use `mode: subagent` to be invoked by a command. The alignment agent should follow the same pattern, invoked by a `/constitution-check` command.
- **Deterministic output**: Reviewers use `temperature: 0.1` for consistent, reproducible analysis. Alignment checking must be deterministic -- the same two constitutions should always produce the same findings.
- **Structured findings**: Reviewers output `### [SEVERITY] Finding Title` with standardized fields. The alignment agent uses the same pattern with alignment-specific fields (Org Principle, Hero Principle, Status, Rationale).

Key differences from the reviewer agents:

| Aspect | Reviewer Agents | Alignment Agent |
|--------|----------------|-----------------|
| Input | Git diff (code changes) | Two constitution files |
| Scope | Code quality, security, intent | Principle alignment only |
| Output fields | File, Line, Convention | Org Principle, Hero Principle, Status |
| Severity levels | CRITICAL/HIGH/MEDIUM/LOW | CONTRADICTION/GAP/ALIGNED |
| Verdict | APPROVE/REQUEST CHANGES | ALIGNED/NON-ALIGNED |

### Alternatives Considered

- **Script-based checking**: A bash script that greps for keywords. Rejected because alignment is semantic, not syntactic -- a principle can support an org principle without using the same words.
- **JSON schema validation**: Define constitutions as structured JSON and validate schemas. Rejected because constitutions are prose documents with nuanced rules that resist mechanical parsing.
- **Manual checklist only**: A markdown checklist for human review. Rejected per clarification (user chose agent-assisted over manual).

## Research Task 2: Constitution Check Integration

### Decision

The `/constitution-check` command is a standalone command, separate from `/speckit.plan`. It can be run independently at any time, but `/speckit.plan` includes a Constitution Check gate that can invoke it.

### Rationale

Separation of concerns:

- `/constitution-check` performs a specific task: compare a hero constitution against the org constitution. It produces a standalone report.
- `/speckit.plan` includes a Constitution Check section that validates the planned work against the active constitution principles. This is a different kind of check -- it validates a spec/plan against principles, not a constitution against another constitution.
- The `/constitution-check` command is useful outside the speckit pipeline (e.g., when bootstrapping a new hero repo or auditing existing repos).

Integration point: The `/constitution-check` command SHOULD be referenced in the speckit pipeline documentation as a recommended pre-flight check when working in a hero repository.

### Alternatives Considered

- **Integrated into `/speckit.plan`**: The plan command would also check constitution alignment. Rejected because this conflates two concerns: plan compliance (does the plan follow principles?) and constitution alignment (does the hero constitution match the org?).
- **Integrated into `/speckit.constitution`**: The constitution command would check alignment after updates. This has merit but limits when alignment can be checked. A standalone command is more flexible.

## Research Task 3: Structured Report Format

### Decision

Use a simple Markdown report format with structured finding blocks. Do not attempt to conform to the inter-hero artifact envelope (Spec 002/009) since those specs are not yet finalized. When Spec 009 defines the artifact envelope, the alignment report can be adapted to produce conforming JSON as a secondary output.

### Rationale

- Specs 002 and 009 are still in draft status. Defining a dependency on an unfinished schema creates circular dependency risk (Spec 001 depends on Spec 009 which depends on Spec 001).
- The alignment report's primary consumer is a human maintainer or the speckit pipeline. Markdown is the natural format for this audience.
- The report structure is simple enough that converting to JSON later (when Spec 009 is finalized) is trivial.

Report structure:

```markdown
# Constitution Alignment Report

**Hero**: [hero name]
**Hero Constitution Version**: [version]
**Org Constitution Version**: [version]
**Checked**: [timestamp]
**Overall Status**: ALIGNED | NON-ALIGNED

## Findings

### [STATUS] [Org Principle Name] ↔ [Hero Principle Name]

**Org Principle**: [principle name and summary]
**Hero Principle**: [principle name and summary]
**Status**: ALIGNED | GAP | CONTRADICTION
**Rationale**: [explanation]

## Summary

- Principles checked: [count]
- Aligned: [count]
- Gaps: [count]
- Contradictions: [count]
- Parent constitution reference: PRESENT | MISSING
```

### Alternatives Considered

- **JSON artifact envelope**: Produce output conforming to the Spec 002 envelope format. Rejected because Spec 002 is not yet finalized; premature dependency.
- **Plain text list**: A simple pass/fail list. Rejected because findings need rationale to be actionable.
- **HTML report**: A formatted HTML document. Rejected because it adds complexity without value for the primary use case (CLI/agent consumption).
