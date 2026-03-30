# Research: Unleash Command

## R1: Resumability Detection Strategy

**Decision**: Detect step completion from filesystem
artifacts using a sequential probe. Each step has a
"done" condition derived from the presence and state
of files in the feature directory.

**Rationale**: Filesystem-based detection is stateless
-- no separate state file to maintain, corrupt, or
get out of sync. The feature directory IS the state.
Each probe is a simple file existence or content check.

**Detection sequence**:

| Step | Done When | Resume At |
|------|-----------|-----------|
| Clarify | No `[NEEDS CLARIFICATION]` markers in spec.md AND (Clarifications section exists OR plan.md exists) | Plan |
| Plan | plan.md exists | Tasks |
| Tasks | tasks.md exists | Spec Review |
| Spec Review | `<!-- spec-review: passed -->` marker in tasks.md | Implement |
| Implement | All tasks `[x]` in tasks.md | Code Review |
| Code Review | CI build+test commands pass (from `.github/workflows/`) | Retrospective |
| Retrospective | Always runs (idempotent -- stores learnings) | Demo |
| Demo | Terminal state | N/A |

**Alternatives considered**:
- JSON state file at `.unbound-force/unleash-state.json`:
  rejected because it adds a file to maintain and can
  get out of sync with actual progress.
- Hive cell status: rejected because it adds a dependency
  on the Swarm plugin for state management.

## R2: Parallel Worker Orchestration

**Decision**: Use Swarm's `swarm_spawn_subtask` +
`swarm_worktree_create` for `[P]`-marked tasks within
each phase. Each worker gets a dedicated worktree.
After all workers complete, merge worktrees back via
`swarm_worktree_merge`. Attempt auto-resolution on
merge conflicts.

**Rationale**: Worktrees provide git-level isolation
so parallel workers can't interfere with each other's
file changes. The `[P]` marker in tasks.md already
identifies which tasks are safe to parallelize (they
touch different files). Swarm's worktree tools handle
the creation/merge lifecycle.

**Workflow per phase**:
1. Parse tasks for `[P]` markers
2. Group: parallel set + sequential set
3. Run sequential tasks first (they may create files
   the parallel tasks depend on)
4. For parallel tasks: create worktree per task,
   spawn worker, wait for all
5. Merge all worktrees (auto-resolve conflicts)
6. Run phase checkpoint (build + test)
7. If checkpoint fails: exit to human

**Alternatives considered**:
- File reservation only (no worktrees): rejected
  because `swarmmail_reserve` prevents conflicts but
  doesn't provide git isolation for clean merges.
- Branch-per-task (no worktrees): rejected because
  branch management is more complex than worktrees
  and doesn't integrate with Swarm's tooling.

## R3: Dewey Clarification Strategy

**Decision**: For each `[NEEDS CLARIFICATION]` marker,
the orchestrating agent extracts the question and
surrounding spec context, formulates a targeted Dewey
semantic search query, evaluates the results using
agent judgment, and either auto-resolves (writes the
answer to the spec) or marks the question as
unanswerable (presents to human).

**Rationale**: The agent is already reading the full
spec during the clarify step. It has the context to
formulate better queries than the raw marker text.
Agent judgment avoids brittle numeric thresholds.

**Dewey query pattern**:
1. Extract the `[NEEDS CLARIFICATION: ...]` text
2. Read 3-5 surrounding lines for context
3. Formulate a query: combine the topic keywords with
   the project domain (from spec title/description)
4. Call `dewey_semantic_search` with the query
5. If results are relevant and sufficiently answer the
   question (agent judgment): auto-resolve
6. If results are empty, off-topic, or ambiguous:
   add to unanswerable list

**Alternatives considered**:
- Numeric similarity threshold (e.g., > 0.7): rejected
  because different question types have different
  "good enough" thresholds. Agent judgment adapts.
- Multiple query strategies (semantic + full-text +
  page lookup): considered but deferred. Start with
  semantic search only. If insufficient, can add
  full-text fallback in a future iteration.

## R4: Demo Instruction Generation

**Decision**: Generate demo instructions by reading the
spec's user stories, the tasks.md completion state, the
test output, and the quickstart.md (if exists). The
demo is a structured Markdown output with sections:
What Was Built, How to Verify, Key Files Changed, Test
Results, and Next Steps.

**Rationale**: The demo must be self-contained -- a
developer should be able to verify the implementation
without reading the full spec or plan. Drawing from
multiple artifacts ensures comprehensive coverage.

**Sources for demo content**:

| Section | Source |
|---------|--------|
| What Was Built | Spec user stories (titles + descriptions) |
| How to Verify | quickstart.md if exists, otherwise generated from acceptance scenarios |
| Key Files Changed | `git diff --name-only main...HEAD` |
| Test Results | Output of `go test -race -count=1 ./...` |
| Quality Summary | Gaze report if available |
| Next Steps | Always: `/finale` to merge, `/speckit.clarify` to iterate |

**Alternatives considered**:
- Interactive demo (run the tool and show output):
  rejected because this is a CLI tool -- the demo
  instructions tell the human what commands to run.
- Auto-generated README section: rejected because
  the demo is ephemeral (per-session), not a
  permanent document.

## R5: Retrospective Content

**Decision**: Store learnings via `hivemind_store` with
tags linking to the feature branch name and the current
date. Content includes: patterns that worked,
gotchas discovered, review findings that required
fixes, and file-specific learnings.

**Rationale**: Future `/unleash` sessions can query
Hivemind via `hivemind_find` for relevant prior
learnings. Tagging with branch name enables
traceability. The learnings are structured as natural
language paragraphs (Hivemind's semantic search works
best on narrative text, not bullet lists).

**Learning categories**:
1. **Patterns**: What coding/design patterns worked
   well in this session
2. **Gotchas**: Unexpected issues or edge cases
   discovered during implementation
3. **Review Insights**: What the review council found
   and how it was fixed
4. **File-Specific**: Learnings about specific files
   that future workers should know

**Alternatives considered**:
- Structured JSON learning records: rejected because
  Hivemind's semantic search works on text, not JSON.
- Mx F coaching agent for retrospective: considered
  but deferred. The coaching agent adds value but
  `/unleash` should work without Mx F installed
  (Composability First).

## R6: Exit Point Design

**Decision**: Each exit point accumulates context up to
that point and presents it in a structured format with
actionable next steps. The exit message always includes:
what step failed, why, what the human should do, and
how to resume.

**Rationale**: A developer encountering an exit should
never be confused about what happened or what to do
next. The exit message is the primary UX of `/unleash`
for non-happy-path scenarios.

**Exit point format**:
```
## /unleash paused at: [step name]

**Reason**: [what happened]
**Context**: [relevant details]

### What to do next
[specific instructions]

### Then resume
Run `/unleash` to continue from [next step].
```

**Alternatives considered**:
- Silent exit with status code: rejected because this
  is an interactive agent command, not a script.
- Offer multiple options at each exit: rejected for
  simplicity. Each exit has one clear recommended
  action.
