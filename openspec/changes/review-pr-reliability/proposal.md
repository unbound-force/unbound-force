## Why

The `/review-pr` command (`review-pr.md`) was adapted
from the `complyctl` repository's `review_pr.md` for
this project. During PR #139 review, six categories
of failure were observed — totaling 12 wasted tool
calls and a review that lacked local tool verification.
The failures stem from structural gaps in the command
instructions that allow the AI agent to deviate from
the intended execution path:

1. **Argument ignored**: The PR number was provided
   as an argument (`139`) but the agent ran
   auto-detection commands anyway (`git branch`,
   `gh pr view`) because Step 1 comes after Step 0
   with no argument-first parsing gate.

2. **Nonexistent `gh` syntax used**: The agent tried
   `gh pr diff 139 -- <path>` (6 times) to read
   individual file diffs. `gh pr diff` does not
   support path filters — the instruction to "process
   file-by-file" is vague and misleading.

3. **PR branch access failures**: The agent tried
   `git show origin/<branch>:<file>` (4 times) and
   `git fetch upstream <branch>` (1 time) to read
   files from the PR branch. Fork-based PRs and PRs
   pushed to GitHub PR refs are not available on
   locally configured remotes.

4. **No execution mode check**: The agent was in
   plan/read-only mode and could not run local tools
   (build, test, lint). Instead of stopping early,
   it skipped Step 4 entirely and produced a review
   without deterministic verification — defeating the
   command's core value proposition of "delegate
   deterministic checks first."

5. **CI-first dedup not applied**: CI returned "no
   checks reported" but the agent did not derive the
   inverse conclusion ("CI covers nothing → ALL local
   tools are required"). The dedup logic is implicit
   and the decision is not made visible.

6. **PR-introduced specs not found**: The spec
   artifacts exist only in the PR diff, not on the
   base branch filesystem. Step 6 checks only the
   local filesystem, missing specs introduced by the
   PR itself.

The original `complyctl` version treats the PR number
as required (not optional), has no auto-detection, and
is structurally simpler. The adaptation introduced
optional arguments and auto-detection without adequate
guards, and did not add mode or capability checks.

## What Changes

Six targeted fixes to the `/review-pr` command
instructions. All changes are to Markdown command
files only — no Go code, no new files.

### Fix 1: Argument-First Parsing Gate

Add explicit argument parsing instruction before
Step 0. When a PR number is provided, set
`PR_NUMBER` immediately and add a DO NOT guard to
Step 1 preventing any auto-detection commands.

### Fix 2: Large Diff Handling Rewrite

Replace the vague "process file-by-file" instruction
in Step 5 with concrete technique: save the full diff
to a temp file, use `grep -n '^diff --git'` to find
file boundaries, use `read` with offset/limit for
specific sections. Add explicit DO NOT list for
commands that fail: `gh pr diff <N> -- <path>`,
`git show <remote>/<branch>:<path>`,
`git fetch <remote> <branch>`.

### Fix 3: PR Branch Access Guidance

Add a section after Step 5 explaining how to access
full file contents from a PR branch using the GitHub
API (`gh api repos/.../contents/...?ref=<headRefName>`)
instead of git operations. Add DO NOT guard against
`git show` and `git fetch` for PR branches.

### Fix 4: Execution Mode Check

Add a mode/capability check to Step 0 (Prerequisites)
that verifies the agent can execute commands. If in
plan/read-only mode, STOP with a message directing
the user to switch modes before invoking the command.
This prevents wasting tokens on metadata, CI, and
diff fetching when the review will be incomplete.

### Fix 5: PR-Introduced Spec Detection

Extend Step 6 to check the PR's changed file list
(from Step 2 metadata) for spec artifacts when they
are not found on the local filesystem. Read spec
content from the saved diff instead of the filesystem.

### Fix 6: Explicit CI Coverage Matrix

Replace the implicit "skip if CI already covers it"
instruction in Step 4 with a mandatory coverage
matrix that maps each CI check to its local tool
equivalent. The agent must display the matrix showing
which tools CI covers and which must run locally,
making the dedup decision visible and auditable.

## Capabilities

### New Capabilities
- `argument-gate`: Argument parsing occurs before
  any tool calls, preventing unnecessary auto-detect.
- `mode-check`: Execution mode verified in
  prerequisites, stopping early when local tools
  cannot run.
- `ci-coverage-matrix`: Visible decision table
  mapping CI checks to local tools, making dedup
  auditable.

### Modified Capabilities
- `large-diff-handling`: Replaced with concrete
  save-and-navigate technique, explicit DO NOT list.
- `spec-detection`: Extended to find specs introduced
  by the PR itself via changed file list.
- `pr-branch-access`: GitHub API path added, git
  operations blocked with DO NOT guard.

### Removed Capabilities
- None.

## Impact

### Files Modified

| File | Change |
|------|--------|
| `.opencode/command/review-pr.md` | All six fixes applied to command instructions |
| `internal/scaffold/assets/opencode/command/review-pr.md` | Scaffold copy synchronized |

### Behavioral Changes

- The command now stops immediately if invoked in
  plan/read-only mode, instead of producing an
  incomplete review.
- PR number arguments are consumed before any tool
  calls, eliminating unnecessary auto-detection.
- Large diffs are navigated via saved temp file
  instead of attempting nonexistent file-filter
  syntax.
- The CI→local tool deduplication decision is
  displayed as a visible matrix.
- Specs introduced by the PR are discoverable.
- Multiple `git show`/`git fetch` failures on PR
  branches are prevented by explicit DO NOT guards.

### No Breaking Changes

- The command output format is unchanged.
- All existing review steps remain in the same order.
- Convention pack loading, security review, and
  constitution compliance steps are unaffected.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

This change modifies an OpenCode command file (agent
instructions). It does not affect inter-hero artifact
formats, communication protocols, or the artifact
envelope schema.

### II. Composability First

**Assessment**: PASS

The `/review-pr` command remains independently usable.
No new mandatory dependencies are introduced. The
mode check uses a simple command execution test — no
special tooling required. The `gh` CLI remains the
only external dependency.

### III. Observable Quality

**Assessment**: PASS

The CI coverage matrix (Fix 6) makes the dedup
decision visible and auditable — the agent must
display which tools CI covers and which must run
locally. This directly improves observability of the
review process. The mode check (Fix 4) prevents
producing reviews that lack deterministic verification
— a direct quality improvement.

### IV. Testability

**Assessment**: N/A

This change modifies Markdown instruction files only.
No Go code is added or changed. The command's
behavioral correctness is verified by re-running
`/review-pr` after the changes and confirming the
failure categories no longer occur. No unit tests
apply to Markdown command files.
