## Context

The `/review-pr` command is a Markdown instruction
file (`.opencode/command/review-pr.md`, 419 lines)
that directs an AI agent through a structured PR
review pipeline. It was adapted from
`complyctl/.opencode/command/review_pr.md` with these
additions: optional PR number with auto-detection,
Step 0 prerequisites, convention pack loading,
OpenSpec spec detection, and enhanced fix-branch/
comment workflows.

During PR #139 review, 12 tool calls failed due to
structural gaps in the instructions. The failures
fall into four categories: ignored argument (1 call),
nonexistent `gh` syntax (6 calls), inaccessible PR
branch (5 calls), and missing mode check (skipped
local tools entirely). The original `complyctl`
version avoids the first category by making the PR
number required.

All changes are Markdown-only — no Go code, no new
dependencies. The scaffold copy at
`internal/scaffold/assets/opencode/command/review-pr.md`
must be synchronized.

## Goals / Non-Goals

### Goals
- Prevent argument-provided PR numbers from being
  ignored by auto-detection logic
- Eliminate tool call failures from nonexistent
  `gh pr diff` file-filter syntax
- Eliminate tool call failures from inaccessible PR
  branch refs (`git show`, `git fetch`)
- Stop the review early when the agent cannot execute
  local tools (plan/read-only mode)
- Make the CI→local tool dedup decision visible and
  auditable
- Find specs that are introduced by the PR itself,
  not only pre-existing on the base branch

### Non-Goals
- Changing the review output format
- Adding new review categories or checks
- Modifying Go source code or tests
- Changing convention pack loading behavior
- Supporting non-GitHub forges (GitLab, Bitbucket)

## Decisions

### D1: Argument Parsing Before Prerequisites

Move PR number extraction to a preamble section
before Step 0, not inside Step 1.

**Current**: Step 0 runs `which gh` + `gh auth status`.
Step 1 resolves PR number. The agent may conflate
prerequisites with argument resolution, running
`git branch` or `gh pr view` during Step 0.

**Change**: Add an "Argument Parsing" section between
the Arguments block and Execution Steps. It
instructs: "Parse the user's message for a PR number.
If found, set `PR_NUMBER` immediately." Step 1 then
has a hard gate: "If `PR_NUMBER` is already set,
skip this step entirely. Do NOT run `gh pr view`,
`git branch --show-current`, or any auto-detection
commands."

**Rationale**: The original `complyctl` version
avoids this by making the PR number required. Since
the adaptation added optional arguments, the gate
must be explicit. Placing it before Step 0 ensures
the decision is made before any tool calls.

### D2: Save-and-Navigate for Large Diffs

Replace the vague "process file-by-file" instruction
with a concrete three-step technique.

**Current**: Step 5 says "If the diff exceeds 500
lines, process file-by-file instead of loading the
entire diff." This implies `gh pr diff` supports
file filtering — it does not.

**Change**: The instruction becomes:
1. Fetch the full diff once:
   `gh pr diff <PR_NUMBER> > /tmp/pr<N>.diff`
   (or use the auto-saved truncation file path)
2. Find file boundaries:
   `grep -n '^diff --git' /tmp/pr<N>.diff`
3. Read specific sections using offset/limit on the
   saved file
4. Explicit DO NOT list: `gh pr diff <N> -- <path>`,
   `git show <remote>/<branch>:<path>`,
   `git fetch <remote> <branch>`

**Rationale**: This is the technique that ultimately
worked during PR #139 review, after 11 failed
attempts with other approaches. Encoding the working
technique directly prevents rediscovery.

### D3: GitHub API for PR File Contents

When the agent needs full file contents (not just the
diff), direct it to the GitHub API instead of git
operations.

**Current**: No guidance. The agent improvises with
`git show` and `git fetch`, which fail for fork-based
PRs and PRs where the branch is not on a configured
remote.

**Change**: Add a section after Step 5:
```
gh api repos/{owner}/{repo}/contents/<path> \
  --jq '.content' | base64 -d
```
with the `ref=<headRefName>` query parameter from
Step 2 metadata. Add DO NOT guard against `git show`
and `git fetch` for PR branch access.

**Rationale**: `gh api` works regardless of remote
configuration, fork setup, or branch naming. It uses
the authenticated `gh` CLI already required by the
command.

### D4: Mode Check as Prerequisite

Add an execution mode check to Step 0 that verifies
the agent can run commands.

**Current**: Step 0 checks `gh` availability and
authentication. No check for whether the agent can
actually execute build/test commands. In plan mode,
the agent silently skips local tools and produces an
unverified review.

**Change**: After `gh auth status`, test execution
capability with a harmless command (e.g.,
`go version` or `make --version`). If the command
cannot be executed, STOP with a message directing the
user to switch to a mode that allows command
execution. The message explicitly states that
reviews without local tool verification do not meet
the command's quality standard.

**Rationale**: The review's core value is "delegate
deterministic checks to local tools first." Without
tool execution, the review degrades to AI-only
judgment — which the command explicitly deprioritizes.
Stopping early saves all tokens that would be spent
on metadata, CI, and diff fetching for an incomplete
review.

### D5: CI Coverage Matrix

Replace the implicit dedup instruction with a
mandatory visible decision table.

**Current**: Step 4 says "If CI already ran and
passed the same checks, skip re-running them
locally." The agent must implicitly map CI check
names to local tool categories. The decision is
invisible.

**Change**: Before executing local tools, the agent
MUST build and display a coverage matrix:

| Local tool | CI check | CI status | Run locally? |
|------------|----------|-----------|--------------|
| `go test`  | CI / test | PASS | No |
| `golangci-lint` | CI / lint | NONE | Yes |

Rules:
- CI PASS → skip locally
- CI FAIL → skip locally (captured in Step 3a)
- CI NONE → MUST run locally
- No CI checks at all → MUST run ALL local tools

**Rationale**: Making the decision visible prevents
the agent from silently skipping tools when CI
returns no data. The matrix also serves as
documentation in the review output, showing the user
which verification was delegated to CI vs. run
locally.

### D6: Spec Detection from PR Changed Files

Extend Step 6 to check the PR's changed file list
for spec artifacts.

**Current**: Step 6 checks only the local filesystem
for `specs/<branch>/spec.md` and
`openspec/changes/<branch>/proposal.md`. When specs
are introduced by the PR itself, they exist only in
the diff.

**Change**: After the filesystem check fails, scan
the changed file list (from Step 2 metadata) for
paths matching spec directories. If found, read the
spec content from the saved diff file (from Step 5)
using the file boundary offsets.

**Rationale**: Many PRs include their own spec
artifacts (OpenSpec proposals, design docs, task
lists). These are valuable context for alignment
review.

## Risks / Trade-offs

### R1: Mode Check False Positives

The mode check (D4) uses a command execution test.
If the tool runtime has intermittent execution
capability (e.g., partial permissions), the check
may pass but later tool executions may fail. This
is an edge case — most runtimes are either fully
capable or fully restricted.

### R2: Instruction Length Growth

The six fixes add approximately 80-100 lines to a
419-line file. This increases the command's token
footprint when loaded. However, the fixes prevent
12+ wasted tool calls per review, which consume
far more tokens than the added instructions.

### R3: DO NOT Lists May Be Incomplete

The DO NOT guards (D2, D3) enumerate specific
failing commands observed during PR #139. Future
tool versions or alternative approaches may
introduce new failure modes not covered. The
guards are a practical minimum, not exhaustive.

### R4: GitHub API Rate Limits

D3 directs the agent to `gh api` for file contents.
For PRs with many files, this could hit GitHub API
rate limits. Mitigation: the instruction only applies
when the agent needs full file contents (rare — most
review uses the diff). The primary file access method
remains the saved diff file.
