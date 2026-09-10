## Context

The `/uf.review-pr` command is a 971-line Markdown file that
serves as both the command instructions and the conversational
context for the reviewing agent. When OpenCode's context
compression activates, it treats the entire file as
conversation content and may discard the procedural steps for
future work (Steps 4-11) while summarizing completed steps
(Steps 0-3). This causes the agent to lose awareness of the
structured review workflow.

The current file structure is monolithic — all 11 steps live in
a single document with no separation between "orchestration"
(what the parent agent manages) and "analysis" (what could run
independently).

## Goals / Non-Goals

### Goals
- Prevent context compression from discarding future step
  instructions by keeping the parent context small
- Isolate token-heavy analysis (diff reading, convention pack
  loading, full file reads) in a subagent so it never bloats
  the parent context
- Preserve all interactive user prompts (CI wait, large diff
  focus, fix-branch offer, verdict posting) in the parent
  where the user can respond
- Maintain identical user-facing behavior — the split is
  invisible to the end user
- Preserve the PR number in the session title through
  compression cycles

### Non-Goals
- Changing the review methodology or steps (8a-8g remain
  identical)
- Splitting the command into multiple files (it remains one
  `.md` file)
- Adding new review capabilities or changing severity
  definitions
- Modifying the scaffold engine or command loading mechanism
- Addressing compression behavior itself (that's OpenCode's
  responsibility)

## Decisions

### D1: Single-file parent/subagent split

The command stays as one `.md` file. The parent section
contains Steps 0-3 (lightweight prerequisites and metadata)
and Steps 9-11 (interactive output and posting). Between them,
a delegation block constructs a Task subagent prompt containing
Steps 4-8 with the PR metadata gathered in Steps 0-3.

**Rationale**: Keeping a single file preserves the existing
scaffold sync mechanism (drift detection tests compare the
command file against its scaffold copy). Multiple files would
require changes to the scaffold engine.

### D2: Subagent prompt is inline, not a separate template

The subagent prompt text (Steps 4-8) is written directly in the
command file as a fenced block that the parent copies into the
Task tool's `prompt` parameter. This avoids introducing a
second file or a template-expansion mechanism.

**Rationale**: The command file is already a template (it uses
`$ARGUMENTS` for PR number injection). Adding a second layer of
templating would increase complexity. The subagent prompt is
static except for the PR metadata values the parent injects.

### D3: Move the large-diff interactive prompt to pre-delegation

Step 5 currently asks the user "Review all files or focus on
specific files?" for PRs with 2000+ diff lines. Since the
subagent cannot interact with the user, this prompt must move
to the parent phase. The parent fetches a lightweight diff
summary (file count and total line count) before delegating,
and if the threshold is exceeded, asks the user which files to
focus on. The file-focus list is then passed to the subagent
as part of its prompt.

**Rationale**: This is the only interactive prompt in Steps 4-8.
Moving it preserves the user's ability to scope the review
while keeping the subagent non-interactive.

### D4: File-based findings with compact summary

The subagent writes full findings to a temporary file
(created via `mktemp`) with structured Markdown sections
matching the Step 9 output format: CI Coverage Matrix,
Local Tool Results, Walkthrough, Linked Issues, Summary,
Alignment findings, Security findings, Constitution
Compliance findings, CI Failure Analysis, Existing
Review State, and Verdict recommendation.

The subagent returns only a compact summary (under 4 KB)
containing the findings file path, verdict, severity
counts, top findings, user login, review count, and
justification. The parent reads needed sections from the
findings file using scoped offset/limit reads, then
renders into the Step 9 output.

**Rationale**: Returning the full report inline caused
tool_output truncation and a double context pass. Writing
to a file and returning a compact summary eliminates
both issues while preserving the structured format.

### D5: Session title preservation via frontmatter description

Add the PR number to the frontmatter `description` field:

```yaml
---
description: "Review PR #$ARGUMENTS"
---
```

This provides a secondary anchor for the session title that
survives compression, since OpenCode uses the frontmatter
description for session metadata.

**Rationale**: Minimal change with immediate effect. The
frontmatter is processed before the content enters the
conversation, so it is not subject to content compression.

### D6: Subagent type selection

Use `subagent_type: "general"` for the Task tool invocation.
The general agent has access to all the tools needed for the
analysis phase: Bash (for `gh` CLI, local tool execution),
Read (for convention packs, spec files), Grep/Glob (for file
discovery), and the skill tool (for pre-flight and
review-context skills).

**Rationale**: No specialized agent type exists for code review
analysis. The general agent is the most appropriate since it
needs diverse tool access.

## Risks / Trade-offs

### R1: Subagent cannot ask interactive questions

**Trade-off accepted**. Only one interactive prompt exists in
Steps 4-8 (large-diff file focus in Step 5). This is moved to
the parent pre-delegation phase (see D3). If future steps add
interactive prompts, they must also be hoisted to the parent.

### R2: Subagent prompt size

The subagent prompt will contain Steps 4-8 (~500 lines of
instructions) plus the injected PR metadata. This is well
within the Task tool's prompt capacity and is smaller than the
current monolithic approach (which loads all 971 lines into the
parent context).

### R3: Skill loading in subagent context

Steps 4 and 6 load skills (`pre-flight` and `review-context`)
via the skill tool. The subagent will need to load these skills
in its own context. This works because the skill tool is
available to the general agent type. The skills are read-only
(they provide instructions, not state).

### R4: Single-file maintenance complexity

Having the parent orchestration and subagent prompt template in
one file makes the file structure less obvious. Mitigated by
clear section headings and comments delineating the parent vs.
subagent sections.

### R5: Scaffold sync

Both `.opencode/commands/uf.review-pr.md` and its scaffold copy
at `internal/scaffold/assets/opencode/commands/uf.review-pr.md`
must be updated in lockstep. The existing drift detection tests
enforce this.
