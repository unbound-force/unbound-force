## ADDED Requirements

### Requirement: Subagent Delegation for Analysis Phase

The `/uf.review-pr` command MUST delegate Steps 4-8 (pre-flight,
diff fetch, context discovery, convention pack loading, AI
review) to a Task subagent using the `general` agent type.

The parent MUST construct the subagent prompt containing:
- The full text of Steps 4-8 from the command template
- PR metadata gathered in Steps 0-3 (PR number, title,
  description, base branch, head branch, changed file list,
  CI check results, CI failure classifications)
- File-focus scope (if the user selected specific files in the
  pre-delegation large-diff prompt)

The subagent MUST write full findings to a temporary file
and return a compact summary (under 4 KB) to the parent.

#### Scenario: Normal PR review with subagent delegation

- **GIVEN** the user invokes `/uf.review-pr 42`
- **WHEN** the parent completes Steps 0-3 and gathers PR metadata
- **THEN** the parent MUST invoke the Task tool with Steps 4-8
  as the prompt, injecting the PR metadata values
- **AND** the parent MUST wait for the subagent to return
  structured findings
- **AND** the parent MUST render the findings into the Step 9
  output format

#### Scenario: Subagent returns findings

- **GIVEN** the subagent completes Steps 4-8
- **WHEN** the subagent returns its findings
- **THEN** the findings MUST contain sections for: CI Coverage
  Matrix, Local Tool Results, Walkthrough, Linked Issues,
  Summary, Alignment, Security, Constitution Compliance,
  CI Failure Analysis, and Verdict recommendation
- **AND** each finding MUST include severity level, category,
  and file/line references where applicable

### Requirement: Pre-delegation Large-Diff Prompt

The parent MUST check the diff size before delegating to the
subagent. If the diff exceeds 2000 lines or 50 files, the
parent MUST ask the user whether to review all files or focus
on specific files, using the AskUserQuestion tool.

Previously: This prompt was in Step 5, inside the analysis
phase. It MUST now execute in the parent context before
delegation.

#### Scenario: Large diff triggers file-focus prompt

- **GIVEN** a PR with a diff exceeding 2000 lines
- **WHEN** the parent prepares to delegate to the subagent
- **THEN** the parent MUST fetch a lightweight diff summary
  (file count and total additions+deletions from Step 2
  metadata)
- **AND** the parent MUST ask the user using the
  AskUserQuestion tool with options `["Review all files",
  "Focus on specific files"]`
- **AND** if the user selects "Focus on specific files", the
  parent MUST follow up with an open-ended question asking
  which files or directories to focus on
- **AND** the file-focus scope MUST be included in the
  subagent prompt

#### Scenario: Small diff skips file-focus prompt

- **GIVEN** a PR with a diff under 2000 lines and under 50
  files
- **WHEN** the parent prepares to delegate to the subagent
- **THEN** the parent MUST NOT prompt the user about file focus
- **AND** the subagent prompt MUST indicate "review all files"

### Requirement: Session Title Preservation

The command frontmatter MUST include a `description` field
containing the PR number for session title preservation:

```yaml
---
description: "Review PR #$ARGUMENTS"
---
```

#### Scenario: Session title survives compression

- **GIVEN** the user invokes `/uf.review-pr 42`
- **WHEN** context compression activates during the review
- **THEN** the session title MUST retain the PR number
  (e.g., "Review PR #42")

## MODIFIED Requirements

### Requirement: Parent Context Scope

The parent agent context MUST contain only Steps 0-3
(prerequisites, PR resolution, metadata fetch, CI checks),
the subagent delegation instruction, and Steps 9-11 (output
format, fix-branch offer, verdict posting).

Previously: The parent context contained all Steps 0-11
(971 lines).

The parent context after restructuring SHOULD NOT exceed
650 lines. The interactive steps (fix-branch offer at ~114
lines and verdict posting at ~234 lines) are irreducibly
complex and MUST remain in the parent context, making a
400-line target infeasible. The primary goal is that the
token-heavy analysis (Steps 4-8, ~470 lines) runs in the
subagent, keeping the parent small enough that compression
does not discard future step instructions.

#### Scenario: Parent context stays lean

- **GIVEN** the `/uf.review-pr` command is invoked
- **WHEN** the command template is loaded into the conversation
- **THEN** the parent-visible content (excluding the subagent
  prompt block) SHOULD NOT exceed 650 lines
- **AND** Steps 4-8 MUST appear only within the subagent prompt
  construction block, not as top-level conversation content

### Requirement: Interactive Steps Remain in Parent

All steps requiring user interaction via the AskUserQuestion
tool MUST remain in the parent context. The subagent MUST NOT
contain any AskUserQuestion prompts.

Previously: All interactive prompts were in the monolithic
command. No change to which prompts exist — only where they
execute.

#### Scenario: Subagent is non-interactive

- **GIVEN** the subagent is executing Steps 4-8
- **WHEN** the subagent encounters a decision point
- **THEN** the subagent MUST use reasonable defaults or
  information provided in its prompt by the parent
- **AND** the subagent MUST NOT attempt to use the
  AskUserQuestion tool

## REMOVED Requirements

None. All existing review steps and capabilities are preserved.
The restructuring changes where steps execute (parent vs.
subagent), not what they do.
