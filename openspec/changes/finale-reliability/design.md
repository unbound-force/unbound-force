## Context

`/uf.finale` is the primary end-of-branch workflow command. It
operates as an 8-step pipeline with mandatory human-confirmation
gates at Steps 3 (commit message) and 5f (PR title/body). The
command uses an execution checklist maintained via the Edit tool
to survive context compression.

Two reliability issues have been identified:

1. The `question` tool renders only a short question string and
   short option labels. Multi-line proposed text (commit messages,
   PR bodies) shown as blockquotes in the instruction text may
   not be visible to the user after context compression, because
   the `question` tool UI does not reproduce the blockquote. The
   user approves text they cannot see.

2. After Step 6 (Watch CI), the agent has a natural "report CI
   results" moment. After context compression or sub-task
   boundaries, this moment becomes a stopping point — the agent
   reports CI results and does not proceed to Steps 7-8.

`/uf.review-council` Step 7f uses the same `question` tool pattern
for its verdict-confirmation gate and has the same visibility gap.

## Goals / Non-Goals

### Goals

- Ensure the full proposed text (commit message, PR title/body,
  review verdict) is visible as plain assistant output BEFORE
  the `question` tool is invoked, so it appears in the
  transcript regardless of context compression.
- Ensure Steps 7-8 execute reliably after Step 6 by adding an
  explicit momentum checkpoint that prevents the agent from
  stopping after CI results.
- Apply the confirmation-gate visibility fix consistently
  across both `/uf.finale` and `/uf.review-council`.

### Non-Goals

- Changing the `question` tool itself (it is an external tool
  provided by the agent runtime).
- Adding new confirmation gates or removing existing ones.
- Changing the execution checklist structure or step ordering.
- Modifying the CI watch logic in Step 6 itself — only the
  post-CI momentum directive is added.

## Decisions

### D1: Print-before-question directive (not tool modification)

**Decision**: Add an explicit directive to the command templates
instructing the agent to print the full proposed text as plain
assistant output BEFORE invoking the `question` tool.

**Rationale**: The `question` tool is external and cannot be
modified by this change. The directive approach works within the
existing tool contract: the agent produces output (visible in the
transcript) and then invokes the tool. This is composable and
does not introduce runtime dependencies.

**Alternative considered**: Embedding the full text in the
question string. Rejected because the `question` tool's question
field is designed for short strings and may truncate or reformat
multi-line content.

### D2: Explicit momentum checkpoint after Step 6

**Decision**: Add a checkpoint instruction immediately after the
Step 6 CI pass/fail block that directs the agent to:
(a) mark Step 6 complete in the execution checklist,
(b) proceed immediately to Step 7 WITHOUT producing any
user-facing output about CI results until Step 8 (Summary) is
reached.

**Rationale**: The root cause is that the agent's natural
"report results" behavior creates a stopping point. By making
the continuation explicit and mandatory in the template, the
agent is instructed to defer reporting until the summary is
complete. This is consistent with the existing session-resume
guard pattern.

**Alternative considered**: Merging Steps 6-8 into a single
compound step. Rejected because it reduces granularity of the
execution checklist and makes resume behavior harder to reason
about.

### D3: Scope the audit to Step 7f of /uf.review-council only

**Decision**: The confirmation-gate visibility audit applies to
`/uf.review-council` Step 7f (verdict confirmation) only. Other
`question` tool usages in `/uf.review-council` (e.g., Step 7c
duplicate-review detection) present short questions with
well-defined options and do not have multi-line content that
could be hidden.

**Rationale**: Focused scope minimizes change surface. The
Step 7f gate is the only other gate that presents multi-line
review content (verdict + inline comments) for approval.

## Risks / Trade-offs

- **Verbose output**: The print-before-question directive adds
  visible output to the transcript. This is intentional — the
  verbosity is the fix. Users who found the output redundant
  can ignore it; users who need to verify what they approved
  now have it in the transcript.

- **Directive enforcement**: Directives in command templates
  rely on agent compliance. If an agent ignores the directive,
  the gate content may still be hidden. Mitigation: the
  session-resume guard already requires re-presentation of
  gate content after compression, providing a second chance.

- **Post-CI deferral**: Deferring CI result reporting until
  Step 8 means the user waits slightly longer for CI feedback.
  Trade-off accepted because the alternative (agent stopping
  at Step 6) leaves the workflow incomplete, which is worse.
