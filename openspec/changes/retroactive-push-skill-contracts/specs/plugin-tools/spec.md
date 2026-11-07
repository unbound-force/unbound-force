<!-- capability: plugin-tools
     change: retroactive-push-skill-contracts
     Archive ordering: this change MUST archive after
     finale-token-reduction (whose text it restates) and after
     fix-tui-git-output-leak (whose inline annotation it
     preserves). Precondition discovered during validation:
     finale-token-reduction's delta file is mis-pathed
     (specs/plugin-tools.md instead of specs/plugin-tools/
     spec.md) and is ignored at apply/archive as-is; it must be
     re-pathed before or at its archive, or this delta's
     MODIFIED has no target spec to modify and the archive
     fails safely (see design.md "finale-token-reduction delta
     is mis-pathed").
     Supersession strategy: this MODIFIED delta restates the
     "Git operation tools" requirement from
     finale-token-reduction's text with the git_push and
     git_push_force_with_lease contracts corrected to the
     shipped required-branch structured-argument shape (PR 948,
     #947). The finale-token-reduction delta itself carries a
     light inline annotation at its two push subsections.
     Design rationale: design.md D2, D3, and D5 in this
     change. -->

Requirement numbering: this delta allocates no new requirement
ids. The plugin-tools capability's requirement in the
unarchived stack (finale-token-reduction/specs/plugin-tools.md)
uses requirement-name addressing with no FR ids; this delta
keeps that addressing unchanged. The "Git operation tools"
requirement is restated below with corrected push-tool
contracts.

## MODIFIED Requirements

### Requirement: Git operation tools

The `uf-workflow` plugin MUST provide the following git
operation tools. Each tool MUST return minimal structured
JSON to the LLM. (Superseded by `fix-tui-git-output-leak`,
openspec/changes/fix-tui-git-output-leak/: tool handlers
MUST NOT print full git output to TUI via `console.log()`;
all output flows through the JSON result envelope.)

All git tools MUST accept a `GitExec` function for
testability.

#### git_fetch

The tool MUST accept `remote` (string) and `branch`
(optional string) parameters.

The tool MUST return `{success: true}` on success or a
failure envelope on error.

#### Scenario: Fetch remote branch

- **GIVEN** a valid remote name `origin` and branch `main`
- **WHEN** `git_fetch` is called
- **THEN** the tool executes `git fetch origin main` and
  returns `{success: true}`

#### git_push

The tool MUST accept a required `branch` parameter (string),
an optional `remote` parameter (string, default `origin`), and
an optional `setUpstream` parameter (boolean). The `branch`
MUST be the currently checked-out branch: the tool pushes the
HEAD commit to `refs/heads/<branch>` with an explicit refspec,
and when the checkout is on a different branch (or detached,
or the current branch cannot be determined) it MUST refuse
with a non-retryable `BRANCH_MISMATCH` envelope naming what
the checkout is actually on, without running any push. The
`remote` value MUST be a git config remote nickname: it MUST
NOT contain `/` (a slash-bearing push target would be resolved
by git as a local repository path, not a remote name), and it
MUST NOT be `.` or `..` (git resolves those as the working
directory and its parent, so a push would target a different
local repository and execute its hooks), while
`branch` continues to allow `/` for names such as `feat/x`.
A malformed, option-like, or empty `branch`, or a
malformed `remote`, MUST fail input validation before any git
command runs. When `setUpstream` is set, the tool MUST
configure tracking to `<remote>/<branch>` after the push; when
the push lands but the tracking configuration fails, the tool
MUST still return success with a `warning` field describing
the failure and the remediation command (`git branch
--set-upstream-to=<remote>/<branch> <branch>`).

The tool MUST return a structured `GitPushData` result —
`{remote, ref, old_sha, new_sha, updated, warning?}` — on
success (`old_sha` is null for a newly created remote ref;
`warning` is present only when a post-push step failed), or a
failure envelope with the error message on error.

#### Scenario: Push with set-upstream

- **GIVEN** `setUpstream: true`, `branch: "feat/x"` naming the
  currently checked-out branch, and the default remote
- **WHEN** `git_push` is called
- **THEN** the tool executes the push of HEAD to
  `refs/heads/feat/x` with an explicit refspec and then
  configures upstream tracking to `origin/feat/x`, returning
  the structured push result

#### Scenario: Upstream configuration failure warns

- **GIVEN** `setUpstream: true`, the push lands, and the
  tracking configuration command fails
- **WHEN** `git_push` returns
- **THEN** the result is success with the structured push data
  and a `warning` field carrying the manual
  `git branch --set-upstream-to=...` remediation

#### Scenario: Branch mismatch refuses to push

- **GIVEN** `branch: "feat/x"` and the checkout is on
  `main`
- **WHEN** `git_push` is called
- **THEN** the tool returns a non-retryable `BRANCH_MISMATCH`
  envelope naming the checked-out branch and executes no push

#### git_push_force_with_lease

The tool MUST accept a required `branch` parameter (string,
the currently checked-out branch — same `BRANCH_MISMATCH`
refusal as `git_push`), an optional `remote` parameter
(string, default `origin`, same nickname validation), and an
optional `expected_sha` parameter (7-64 character hex commit
SHA). The tool MUST push the HEAD commit to
`refs/heads/<branch>` with an explicit refspec under lease
protection: leasing on the supplied `expected_sha` when given
(a locally unknown `expected_sha` MUST fail with a
non-retryable `EXPECTED_SHA_UNKNOWN` envelope); otherwise on
the remote-tracking ref `<remote>/<branch>`; otherwise — when
no remote-tracking ref exists — on absence. The tool MUST
return the structured `GitPushData` result on success or a
failure envelope on error.

#### Scenario: Force push with lease

- **GIVEN** a branch with diverged history and
  `branch` naming the currently checked-out branch
- **WHEN** `git_push_force_with_lease` is called
- **THEN** the tool pushes HEAD to `refs/heads/<branch>` with
  an explicit refspec under `--force-with-lease` protection
  and returns the structured push result

#### git_merge

The tool MUST accept a `ref` (string) parameter.

The tool MUST return
`{success: true, conflicted: false}` on clean merge or
`{success: false, conflicted: true, conflictingFiles: [...]}`.

#### Scenario: Clean merge

- **GIVEN** a ref `origin/main` that merges cleanly
- **WHEN** `git_merge` is called with `ref: "origin/main"`
- **THEN** the tool returns
  `{success: true, conflicted: false}`

#### Scenario: Merge with conflicts

- **GIVEN** a ref `origin/main` that conflicts
- **WHEN** `git_merge` is called
- **THEN** the tool returns `{success: false,
  conflicted: true, conflictingFiles: ["file.go"]}`

#### git_merge_abort

The tool MUST execute `git merge --abort` and return
`{success: true}`.

#### git_rebase

The tool MUST accept a `ref` (string) parameter.

The tool MUST return
`{success: true, conflicted: false}` on clean rebase or
`{success: false, conflicted: true, conflictingFiles: [...]}`.

#### git_rebase_abort

The tool MUST execute `git rebase --abort` and return
`{success: true}`.

#### git_conflict_files

The tool MUST return `{files: [...]}` listing all files
with unresolved merge conflicts.

#### git_commit_no_edit

The tool MUST execute `git commit --no-edit` and return
`{success: true, sha: "..."}` or a failure envelope.

Previously: the requirement (as added by
finale-token-reduction) accepted an optional `branch` (string)
on `git_push` — pushing whatever git's default refspec chose
for the checkout — and `git_push_force_with_lease` executed a
bare `git push --force-with-lease` with no structured
arguments, both returning `{success: true}`. The shipped
tools (#947) require `branch` (the currently checked-out
branch, pushed via an explicit refspec), validate `remote` as
a nickname, and return structured `GitPushData` with an
optional post-push `warning`.

## REMOVED Requirements

None.
