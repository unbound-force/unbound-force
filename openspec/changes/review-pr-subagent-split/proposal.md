## Why

The `/uf.review-pr` command (971 lines) injects its entire
step-by-step workflow as conversation content. When OpenCode's
context compression activates — especially on small PRs where
the command template dominates the token budget — the
compression system correctly identifies that early steps are
"done" but incorrectly discards the instructions for future
steps (Steps 4-11) along with them.

The agent then improvises a review workflow from general
knowledge, losing structured sub-steps (8a-8g), output format
requirements (Step 9), the interactive verdict-posting flow
(Step 11), and the session title's PR number context.

This is filed as [#373](https://github.com/unbound-force/unbound-force/issues/373).

## What Changes

Split the `/uf.review-pr` command into a parent/subagent
architecture where the parent retains lightweight orchestration
instructions and the token-heavy analysis runs in an isolated
Task subagent.

**Parent context** (stays in conversation, never compressed):
- Steps 0-3: prerequisites, PR metadata, CI checks (lightweight)
- Subagent delegation instruction with structured prompt
- Steps 9-11: output formatting, fix-branch offer, interactive
  verdict posting (requires user interaction)

**Task subagent** (isolated context, writes findings to file):
- Steps 4-8: pre-flight, diff fetch, context discovery,
  convention pack loading, full AI review (8a-8g)
- Writes full findings to a temporary file and returns a
  compact summary (under 4 KB) to the parent

**Session title fix**: Add the PR number to the frontmatter
`description` field so compression preserves it.

## Capabilities

### New Capabilities
- `subagent-delegation`: Parent delegates Steps 4-8 to a Task
  subagent, keeping the parent context small enough that
  compression never triggers on the command instructions
- `file-based-findings-return`: Subagent writes full findings
  to a temporary file and returns a compact summary (under
  4 KB) that the parent uses to read needed sections via
  scoped offset/limit reads

### Modified Capabilities
- `/uf.review-pr`: Same user-facing behavior — the split is
  invisible to the user. All interactive prompts (pending CI
  wait, fix-branch offer, verdict posting) remain in the parent
  context where the user can respond.
- `session-title`: PR number preserved in frontmatter
  `description` field, surviving compression

### Removed Capabilities
- None

## Impact

- **Primary file**: `.opencode/commands/uf.review-pr.md` (971
  lines, restructured into parent orchestrator)
- **Scaffold copy**: `internal/scaffold/assets/opencode/commands/uf.review-pr.md`
  (must stay in sync)
- **Interactive steps moved to pre-delegation**: The "Wait for
  pending CI checks?" prompt (Step 3) already runs before
  delegation. The "Focus on specific files?" prompt for large
  diffs (Step 5) must move to the parent pre-delegation phase,
  since the subagent cannot ask the user questions.
- **No Go code changes**: This is a command template
  restructuring, not a code change
- **No test changes**: Command files are not unit-tested; drift
  detection tests verify scaffold sync

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

The subagent communicates with the parent through a two-part
output contract: full findings written to a temporary file
and a compact summary (under 4 KB) returned inline. This is
artifact-based communication — the subagent produces a
self-describing output (findings with severity, category,
file/line references) that the parent reads from the file
using scoped offset/limit reads without synchronous coupling
during the analysis phase.

### II. Composability First

**Assessment**: PASS

The command remains a single standalone file. The subagent
delegation uses OpenCode's built-in Task tool — no new hero
dependencies are introduced. The command works identically
whether or not context compression is enabled; the split
simply makes it resilient to compression.

### III. Observable Quality

**Assessment**: PASS

The findings file written by the subagent is machine-parseable
(severity levels, file paths, line numbers, category tags),
and the compact summary provides structured key-value fields
for verdict, counts, and top findings. The output format
(Step 9) is unchanged.

### IV. Testability

**Assessment**: N/A

This change modifies a command template (Markdown instructions),
not executable code. The existing scaffold drift detection tests
verify that the command file and its scaffold copy stay in sync.
No new testable components are introduced.

### V. Security by Default

**Assessment**: PASS

The subagent inherits the same security review steps (8b) and
the same shell injection guards (temp file for commit messages,
temp file for review JSON payloads). No new input vectors are
introduced. The subagent prompt is constructed from the command
template, not from user-controlled input.
