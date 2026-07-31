## Context

The `opsx-propose` command (`.opencode/commands/opsx-propose.md`)
and the `openspec-propose` skill
(`.opencode/skills/openspec-propose/SKILL.md`) share nearly
identical instruction text. Both have two structural gaps
identified in issue #353:

1. The dirty-tree guard in Step 3a is prose-only — it describes
   the check but does not specify the `AskUserQuestion` tool
   call with concrete options. Under context compression,
   agents may skip the guard.

2. The "STOP HERE" rule appears after the full workflow, meaning
   agents under context compression may never reach it.

Both files must be updated in lockstep since they contain
parallel copies of the same workflow.

## Goals / Non-Goals

### Goals

- Make the dirty-tree guard machine-enforceable by adding an
  explicit `AskUserQuestion` call with two structured options
- Front-load the implementation prohibition as a preamble
  before the workflow steps, so it is encountered early in
  context and less likely to be dropped during compression
- Apply identical fixes to both the command file and the
  skill file to maintain parity
- Retain the existing STOP HERE block after Step 6 as
  reinforcement (belt-and-suspenders approach)

### Non-Goals

- Refactoring the command/skill to share a single source
  (that would be a separate change to address DRY concerns)
- Adding automated tests for agent instruction files (these
  are Markdown consumed by LLM agents, not executable code)
- Modifying other commands or skills that may have similar
  patterns (each should be addressed via its own issue)

## Decisions

### D1: AskUserQuestion with two options

The dirty-tree guard will use `AskUserQuestion` with exactly
two options:

1. "Stash changes and continue" — agent runs `git stash`,
   then proceeds with branch creation
2. "Abort — keep changes as-is" — agent stops and reports
   that the user needs to deal with uncommitted changes first

**Rationale**: Two concrete options eliminate ambiguity. The
agent does not need to interpret prose — it presents the
options and acts on the user's selection. The "stash" option
provides a convenient path forward without losing work.

### D2: Preamble placement before Step 1

A bolded preamble will be inserted immediately after the
`**Steps**` heading and before Step 1. This positions the
implementation prohibition at the top of the workflow
instructions, where it will be encountered early during
context processing.

**Rationale**: Rules that appear after the actions they gate
are unreliable under context compression. Front-loading the
prohibition follows the same pattern used in other commands
(e.g., the Guardrails section that already exists at the end
of both files serves as reinforcement, not as the primary
gate).

### D3: Parallel updates to both files

Both `.opencode/commands/opsx-propose.md` and
`.opencode/skills/openspec-propose/SKILL.md` will receive
identical structural changes. The wording may differ slightly
due to frontmatter differences, but the guard logic and
preamble content will be the same.

**Rationale**: These files are consumed in different contexts
(slash command vs. skill invocation) but implement the same
workflow. Divergence between them would create confusion and
risk one being fixed while the other remains vulnerable.

## Risks / Trade-offs

### R1: Preamble adds to context length

Adding a preamble before Step 1 increases the total token
count of both files. Under aggressive context compression,
this could push other content out.

**Mitigation**: The preamble is concise (3 lines). The
trade-off is acceptable because the prohibition it conveys
is the single most important rule in the file — it must
survive compression even if other instructions are trimmed.

### R2: Dual-file maintenance burden

Maintaining two files with the same workflow content doubles
the surface area for future fixes.

**Mitigation**: This is a known limitation documented as a
non-goal. A future change could extract shared content, but
that is out of scope here.

### R3: "Stash" option may surprise users

The "Stash changes and continue" option performs
`git stash --include-untracked` on the user's behalf,
which may be unexpected.

**Mitigation**: The option text clearly states "stash" and
the AskUserQuestion dialog requires explicit user selection.
No action is taken without consent. If the stash command
fails, the agent stops and reports the error rather than
proceeding.

## Verification Protocol

Since this change modifies Markdown instruction files (not
executable source code), automated unit tests are not
applicable. Manual acceptance verification:

1. Run `/opsx-propose <name>` with a dirty working tree
   (staged, unstaged, or untracked files present)
2. Confirm AskUserQuestion appears with exactly two options
3. Select "Stash changes and continue" — confirm
   `git stash --include-untracked` runs and the branch
   is created
4. Select "Abort — keep changes as-is" — confirm execution
   stops and the user is informed
5. Verify the preamble appears before Step 1 in both files
   by visual inspection
6. Run `diff` on the guard and preamble sections of both
   files to confirm parity (accounting for known
   invocation syntax divergence)
