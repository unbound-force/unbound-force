## ADDED Requirements

### Requirement: Argument validation

The `/uf.rebase-pr` command MUST require a PR number argument.
The command MUST validate the argument against
`^[1-9][0-9]*$` before any tool call. If the argument is
missing, invalid, zero, negative, or contains non-numeric
characters, the command MUST STOP with the error "Invalid PR
number. Must be a positive integer." and MUST NOT invoke any
`gh` CLI command. All subsequent steps MUST use the validated
`PR_NUMBER` value.

#### Scenario: Valid PR number

- **GIVEN** the user invokes `/uf.rebase-pr 42`
- **WHEN** the command parses the argument
- **THEN** validation passes and `PR_NUMBER` is set to `42`
  before any tool call

#### Scenario: Non-numeric argument

- **GIVEN** the user invokes `/uf.rebase-pr abc`
- **WHEN** the command parses the argument
- **THEN** the command reports "Invalid PR number. Must be a
  positive integer." and stops without running any `gh`
  command

#### Scenario: Zero or negative argument

- **GIVEN** the user invokes `/uf.rebase-pr 0` or
  `/uf.rebase-pr -3`
- **WHEN** the command parses the argument
- **THEN** the command reports "Invalid PR number. Must be a
  positive integer." and stops

#### Scenario: Leading zeros

- **GIVEN** the user invokes `/uf.rebase-pr 042`
- **WHEN** the command parses the argument
- **THEN** validation fails (leading zeros are not accepted)
  and the command stops

### Requirement: Prerequisites and repository detection

The command MUST verify the `gh` CLI is installed (`which gh`)
and authenticated (`gh auth status`) before any GitHub API
call. If `gh` is missing, the command MUST STOP with an
installation pointer. If authentication fails, the command
MUST STOP with guidance to run `gh auth login`. The command
MUST detect the current repository via
`gh repo view --json nameWithOwner` and MUST NOT hardcode any
repository identifier.

#### Scenario: gh not installed

- **GIVEN** the `gh` binary is not on PATH
- **WHEN** the user invokes `/uf.rebase-pr 42`
- **THEN** the command reports that the GitHub CLI is not
  installed, with its installation URL, and stops

#### Scenario: gh not authenticated

- **GIVEN** `gh` is installed but `gh auth status` fails
- **WHEN** the user invokes `/uf.rebase-pr 42`
- **THEN** the command reports that the CLI is not
  authenticated and directs the user to run `gh auth login`

#### Scenario: Repository detected dynamically

- **GIVEN** the command runs inside the repository
  `unbound-force/unbound-force`
- **WHEN** the prerequisite phase completes
- **THEN** the recorded repository identifier is
  `unbound-force/unbound-force`, resolved from `gh repo view`
  rather than any hardcoded value

### Requirement: PR state gating

The command MUST fetch the PR via
`gh pr view <PR_NUMBER> --json number,title,state,headRefName,
baseRefName,mergeable,mergeStateStatus,baseRepository,
headRepository`. The command MUST STOP with the `gh` error if
the PR does not exist, MUST STOP if the PR state is CLOSED or
MERGED, and MUST STOP with a no-conflicts report when
`mergeable` is `MERGEABLE` and `mergeStateStatus` is `CLEAN`.
Otherwise the command MUST record `HEAD_BRANCH`
(`headRefName`), `BASE_BRANCH` (`baseRefName`), `BASE_REPO`
(`baseRepository.owner.login/baseRepository.name`), and
`HEAD_REPO` (`headRepository.owner.login/headRepository.name`).

#### Scenario: PR does not exist

- **GIVEN** no PR #999 exists in the repository
- **WHEN** the user invokes `/uf.rebase-pr 999`
- **THEN** the command stops and surfaces the `gh` error
  output verbatim

#### Scenario: PR already merged

- **GIVEN** PR #42 is in state MERGED
- **WHEN** the user invokes `/uf.rebase-pr 42`
- **THEN** the command reports "PR #42 is MERGED. Nothing to
  resolve." and stops

#### Scenario: PR has no conflicts

- **GIVEN** PR #42 reports `mergeable: MERGEABLE` and
  `mergeStateStatus: CLEAN`
- **WHEN** the user invokes `/uf.rebase-pr 42`
- **THEN** the command reports "PR #42 has no conflicts to
  resolve." and stops without checking out or rebasing
  anything

#### Scenario: Conflicted fork PR records fork metadata

- **GIVEN** PR #42 targets `unbound-force/unbound-force` from
  fork `alice/unbound-force` with `mergeable: CONFLICTING`
- **WHEN** the user invokes `/uf.rebase-pr 42`
- **THEN** the command records `BASE_REPO` as
  `unbound-force/unbound-force` and `HEAD_REPO` as
  `alice/unbound-force`, and proceeds to checkout

### Requirement: Head branch checkout

The command MUST check out the PR's head branch via
`gh pr checkout <PR_NUMBER>`. If that fails (for example, the
head branch was deleted), the command MUST fall back to
fetching the head ref directly from the PR's head repository
(`git fetch https://github.com/<HEAD_REPO>.git
<HEAD_BRANCH>`) and checking it out
(`git checkout -B <HEAD_BRANCH> FETCH_HEAD`). If both paths
fail, the command MUST STOP with the failure.

#### Scenario: Checkout succeeds

- **GIVEN** PR #42's head branch exists on its head remote
- **WHEN** the checkout phase runs
- **THEN** `gh pr checkout 42` succeeds and the working tree
  is on `HEAD_BRANCH`

#### Scenario: Checkout fails, fallback succeeds

- **GIVEN** `gh pr checkout 42` fails because the local
  checkout cannot be configured from the PR metadata
- **WHEN** the fallback runs
- **THEN** the head ref is fetched directly from
  `https://github.com/<HEAD_REPO>.git` and checked out as
  `HEAD_BRANCH` from `FETCH_HEAD`

### Requirement: Fork-aware remote resolution

The command MUST resolve `BASE_REMOTE` and `HEAD_REMOTE` by
matching `git remote -v` fetch/push URLs against the PR's
recorded `BASE_REPO` and `HEAD_REPO`. The command MUST NOT
assume `origin` is the base repository — for a fork PR,
`origin` often points to the fork. If no existing remote
matches `BASE_REPO` (or `HEAD_REPO`), the command MUST add
one under a name that does not collide with existing remotes.
For a same-repo PR, `BASE_REMOTE` and `HEAD_REMOTE` resolve to
the same remote.

#### Scenario: Fork PR with base remote named upstream

- **GIVEN** PR #42 targets `unbound-force/unbound-force` from
  fork `alice/unbound-force`
- **AND** `git remote -v` shows `origin` →
  `alice/unbound-force` and `upstream` →
  `unbound-force/unbound-force`
- **WHEN** remotes are resolved
- **THEN** `BASE_REMOTE` is `upstream` and `HEAD_REMOTE` is
  `origin`

#### Scenario: No remote matches the base repository

- **GIVEN** PR #42 targets `unbound-force/unbound-force` but
  no configured remote points at it
- **WHEN** remotes are resolved
- **THEN** the command adds a new remote for
  `https://github.com/unbound-force/unbound-force.git` under
  a non-colliding name and uses it as `BASE_REMOTE`

#### Scenario: Same-repo PR

- **GIVEN** PR #42's head and base are both
  `unbound-force/unbound-force` and `origin` points at it
- **WHEN** remotes are resolved
- **THEN** `BASE_REMOTE` and `HEAD_REMOTE` both resolve to
  `origin`

### Requirement: Rebase onto base

The command MUST fetch the latest base branch from
`BASE_REMOTE` and rebase the head branch onto
`<BASE_REMOTE>/<BASE_BRANCH>`. The command MUST use the
existing git plugin tools (`git_fetch` with the resolved
remote, `git_rebase` with the resolved ref) rather than raw
bash git invocations.

#### Scenario: Clean rebase

- **GIVEN** the head branch is checked out and
  `BASE_REMOTE/BASE_BRANCH` is fetched
- **WHEN** `git_rebase` runs with ref
  `upstream/main`
- **THEN** the rebase completes with no conflicts and the
  flow proceeds to the push phase

#### Scenario: Rebase stops with conflicts

- **GIVEN** the head branch and base branch both modified
  `internal/foo.go`
- **WHEN** `git_rebase` runs
- **THEN** the rebase stops with unmerged paths and the flow
  proceeds to conflict resolution

### Requirement: Conflict resolution discipline

When the rebase stops with conflicts, the command MUST list
unmerged paths (`git status`), resolve ONLY the files the
rebase flagged, and preserve the PR author's intended code
without unrelated changes. For each conflicted file the
command MUST resolve the conflict markers, stage the file, and
continue the rebase, repeating until the rebase completes. If
a conflict cannot be resolved with confidence, the command
MUST NOT guess — it MUST STOP, explain what is blocking, and
offer the user the choice to abort with `git rebase --abort`.
If stuck mid-rebase, the command MUST back out cleanly via the
`git_rebase_abort` plugin tool and report the situation rather
than forcing a bad resolution.

#### Scenario: Resolve a flagged file

- **GIVEN** the rebase flags `internal/foo.go` as conflicted
- **WHEN** the command resolves the markers preserving the PR
  author's intent and stages the file
- **THEN** `git rebase --continue` advances the rebase and no
  unrelated file is modified

#### Scenario: Ambiguous conflict stops

- **GIVEN** a conflict in `internal/foo.go` cannot be
  resolved with confidence
- **WHEN** the command evaluates the conflict
- **THEN** it stops, explains the blocker, and offers
  `git rebase --abort` rather than fabricating a resolution

#### Scenario: Clean backout

- **GIVEN** the command is stuck mid-rebase
- **WHEN** the user chooses to abort
- **THEN** `git_rebase_abort` restores the pre-rebase branch
  state and the command reports the situation

### Requirement: Bounded force-push

The rebase rewrites history, so the head branch MUST be
force-pushed with lease protection: `git push
--force-with-lease <HEAD_REMOTE> <HEAD_BRANCH>`. The command
MUST NOT use `--force`. The command MUST push only to
`HEAD_REMOTE` (the remote resolved for `HEAD_REPO`) and only
the PR's head branch. The command MUST NOT push to
`BASE_REMOTE`, `main`, `master`, or any protected branch.
When the checked-out branch's configured upstream is verifiably
`<HEAD_REMOTE>/<HEAD_BRANCH>`, the command SHOULD use the
`git_push_force_with_lease` plugin tool; when the upstream does
not match (or is unset, as after the checkout fallback), the
command MUST fall back to the explicit raw refspec command to
guarantee the push target.
(Superseded by `retroactive-push-skill-contracts`,
openspec/changes/retroactive-push-skill-contracts/: the force
push MUST always go through the `git_push_force_with_lease`
plugin tool with `{branch: <HEAD_BRANCH>, remote:
<HEAD_REMOTE>, expected_sha: <HEAD_SHA>}` — including after
the checkout fallback — via an explicit refspec independent of
any configured upstream; raw shell force-push forms are
forbidden. That change MUST archive after this change.)

#### Scenario: Plugin tool push with verified upstream

- **GIVEN** after `gh pr checkout` the branch's upstream is
  `origin/feature-x` where `HEAD_REMOTE` is `origin` and
  `HEAD_BRANCH` is `feature-x`
- **WHEN** the push phase runs
- **THEN** the command uses the `git_push_force_with_lease`
  plugin tool, which pushes `feature-x` to `origin` under
  lease protection

#### Scenario: Explicit refspec push after fallback checkout

- **GIVEN** the head branch was checked out via the
  fetch-from-head-repo fallback and has no upstream configured
- **WHEN** the push phase runs
- **THEN** the command runs
  `git push --force-with-lease <HEAD_REMOTE> <HEAD_BRANCH>`
  explicitly, guaranteeing the push target

#### Scenario: Never pushes the base branch

- **GIVEN** `BASE_REMOTE` is `upstream` and `BASE_BRANCH` is
  `main`
- **WHEN** the push phase runs
- **THEN** the push targets only `<HEAD_REMOTE>
  <HEAD_BRANCH>`; no push to `upstream` or `main` is issued

### Requirement: Post-push verification

After the push, the command MUST re-check the PR via
`gh pr view <PR_NUMBER> --json mergeable,mergeStateStatus`
and report whether the conflict is resolved. The command MUST
NOT merge, approve, or close the PR — it only makes the PR
ready to merge with a rebase once approved.

#### Scenario: Conflict resolved

- **GIVEN** the rebased branch was pushed successfully
- **WHEN** the verification runs
- **THEN** `mergeable` is `MERGEABLE`, the command reports
  success, and the PR remains open and unmerged

#### Scenario: PR still reports conflicts

- **GIVEN** the push succeeded but GitHub still computes
  `mergeable: CONFLICTING`
- **WHEN** the verification runs
- **THEN** the command reports the persisted conflict state
  rather than claiming success

### Requirement: Command file conventions

The ported command file `.opencode/commands/uf.rebase-pr.md`
MUST follow the conventions of `uf.review-pr.md`: YAML
frontmatter with a `description` field, a `<protect>` block
wrapping the instructions, a session-resume guard at the top,
an Arguments section with the pre-validation contract, phased
execution steps, and a Guardrails section carrying the
no-merge/no-approve, never-push-base, resolve-remotes,
force-with-lease, conflicts-only, no-guessing, clean-backout,
gh-verification, and determinism rules from the verified
source command.

#### Scenario: Frontmatter and protect block

- **GIVEN** the ported command file is written
- **WHEN** its structure is inspected
- **THEN** it begins with YAML frontmatter containing a
  description, and its instruction body is wrapped in a
  `<protect>` block

#### Scenario: Session-resume guard present

- **GIVEN** the ported command file is written
- **WHEN** a session resumed from compressed context
  re-invokes the command
- **THEN** the guard directs the agent to re-read the template
  before continuing rather than inferring step completion from
  compressed summaries

### Requirement: Scaffold mirror and drift tests

The command file MUST be mirrored byte-identically at
`internal/scaffold/assets/opencode/commands/uf.rebase-pr.md`
as a tool-owned asset. The `expectedAssetPaths` list in
`internal/scaffold/scaffold_test.go` MUST include the new
asset path (updating the commands count comment from 11 to
12). The scaffold drift tests (`TestEmbeddedAssets_MatchSource`,
`TestAssetPaths_MatchExpected`) MUST pass, and `make check`
MUST be green.

#### Scenario: Mirror matches canonical source

- **GIVEN** the command is ported to
  `.opencode/commands/uf.rebase-pr.md`
- **WHEN** the scaffold drift tests run
- **THEN** the embedded asset at
  `internal/scaffold/assets/opencode/commands/uf.rebase-pr.md`
  matches the canonical file and both tests pass

#### Scenario: Asset list updated

- **GIVEN** the new command file exists
- **WHEN** `expectedAssetPaths` is inspected
- **THEN** it contains `opencode/commands/uf.rebase-pr.md`
  and the count comment reads 12

#### Scenario: uf init deploys the command

- **GIVEN** a project scaffolded by `uf init` at a version
  that includes this asset
- **WHEN** `uf init` deploys embedded assets
- **THEN** `.opencode/commands/uf.rebase-pr.md` is written to
  the project as a tool-owned file

### Requirement: Documentation updates

The change MUST document the new command in four places:
a `/uf.rebase-pr [N]` row in the `AGENTS.md` PR Review
Commands table (Post-PR timing, rebase-conflicted-PR scope)
together with the `AGENTS.md` Project Structure command-count
bump ("Slash commands (32 files)" → "(33 files)"), a row in
the `docs/usage.md` Quick Reference table, a row in the
`docs/architecture.md` Review Commands table, and a
`port-uf-rebase-pr:` entry in `CHANGELOG.md` under
"## Unreleased" → "### Added" carrying
"(Spec: openspec/changes/port-uf-rebase-pr/)" and, once
filed, "(Docs: #N, unbound-force/website#N)" references.
The `docs/architecture.md` row is a consistency extension
beyond the routed scope's explicit file list: that table
already enumerates the sibling post-PR commands, and
omitting the new command would leave it stale. The change
MUST also satisfy the documentation gate before the
implementing PR merges: an in-repo docs issue with
per-surface dispositions (`CHANGELOG.md`, `AGENTS.md`,
`docs/usage.md`, `docs/architecture.md`, `README.md`,
each updated-or-no-change-plus-why, plus tutorial and
blog judgments) and a `unbound-force/website` sync issue,
with both issue numbers recorded in the CHANGELOG entry
and in proposal.md's Impact section.

#### Scenario: AGENTS.md table row

- **GIVEN** the change is implemented
- **WHEN** the AGENTS.md PR Review Commands table is read
- **THEN** it contains
  `/uf.rebase-pr [N] | Post-PR (GitHub) | Rebase a conflicted
  PR's head branch onto its base`

#### Scenario: docs tables updated

- **GIVEN** the change is implemented
- **WHEN** `docs/usage.md` and `docs/architecture.md` are read
- **THEN** both command tables include a `/uf.rebase-pr`
  entry consistent with their existing row formats

#### Scenario: CHANGELOG entry present

- **GIVEN** the change is implemented
- **WHEN** the `CHANGELOG.md` "## Unreleased" →
  "### Added" section is read
- **THEN** it contains a `- port-uf-rebase-pr:` entry
  carrying the Spec reference and, once filed, the Docs
  issue references

#### Scenario: AGENTS.md command count bumped

- **GIVEN** the change is implemented
- **WHEN** the `AGENTS.md` Project Structure tree is read
- **THEN** the `.opencode/commands/` annotation reads
  "Slash commands (33 files)"

#### Scenario: Documentation gate issues filed

- **GIVEN** the change's PR is ready to merge
- **WHEN** the documentation gate is checked
- **THEN** an in-repo docs issue with per-surface
  dispositions and tutorial/blog judgments exists, a
  `unbound-force/website` sync issue exists, and both
  issue numbers appear in the CHANGELOG entry and
  proposal.md's Impact section

## Test Strategy

This change ports an agent-instruction command (declarative
markdown) and its scaffold mirror. The testable surface is
the scaffold asset machinery, per constitution principle IV:

### Scaffold Asset Tests (Go)

- `TestAssetPaths_MatchExpected` — the new asset appears in
  `expectedAssetPaths`
- `TestEmbeddedAssets_MatchSource` — drift detection: the
  embedded mirror matches the canonical command byte-for-byte
- `TestRun_CreatesFiles` — the deployed file count reflects
  the added asset
- `TestCanonicalSources_AreEmbedded` — reverse drift: the
  canonical command file is embedded

### Coverage Target

- No new compiled code is introduced; the coverage strategy
  is the existing scaffold test surface above, which MUST
  remain green under `make check` (including the scope-based
  Go coverage gate and the TypeScript plugin suite)

## MODIFIED Requirements

None. This change does not modify existing command behavior.
`/uf.finale` step 4 and its conflict recovery (step 7b) are
unchanged.

## REMOVED Requirements

None. No existing functionality is removed.
