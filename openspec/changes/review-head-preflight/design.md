## Context

The two PR-mode review commands freeze `head_sha` in their input
context. Their local gates, though, run wherever the agent happens
to be:

- `/uf.review-pr` Step 3.7: `pre-flight` in `ci-aware` mode.
- `/uf.review-council` Phase 1a: `pre-flight` in `soft-gate` mode.
- `/uf.review-council` Phase 1b: `gaze-reporter`.

The agent-directory dependence reaches into every phase of the
`pre-flight` skill:

- Phase 1 reads `.github/workflows/`.
- Phase 2 probes `Makefile`, `go.mod`, and the other config files.
- Phase 2a computes `git diff ${MERGE_BASE}...HEAD`. On `main`
  that diff is empty, and the empty-diff fallback does not fire on
  the default branch, so every scoped tool is SKIPped and the PR
  passes.
- Phase 4 executes the tools.
- Phase 2b calls `uf_preflight_declared_checks`, whose `cwd` is
  fixed at plugin start.

`preflight-declared.ts` already has the needed pattern:
`classifyFailures` adds a detached worktree, runs the primary
checkout's scripts with that worktree as `cwd`, and removes it in
a `finally`.

The proposal's Constitution Alignment is PASS on all four
principles. D3, D4, and D8 below carry Composability First and
Observable Quality.

## Goals / Non-Goals

### Goals
- In PR mode, every pre-flight phase (workflow parsing, tool
  detection, diff scope, execution, declared checks, baseline)
  runs against `head_sha`, or does not run, and the report says
  which.
- The diff scope for a PR is `base_sha...head_sha`, not the
  default branch's merge-base.
- The user's branch, index, and working tree are never switched,
  stashed, or modified.
- Untrusted fork code is never fetched and executed implicitly.
- Worktree choreography lives in a tested tool, not command prose
  (Determinism gate, DE-001/DE-005).

### Non-Goals
- Fixing #1063 for non-PR input (diff scope when a branch is
  behind main).
- Changing the `ci-aware` / `soft-gate` decision rules.
- The `/uf.review-pr` CI-fix branch flow (Steps 3b+), which checks
  out branches on purpose.
- Editing `gaze-reporter.md`, which is scaffolded by the gaze repo
  (`<!-- scaffolded by gaze dev -->`). A root argument belongs
  there.
- Submodules and LFS. A fresh worktree has uninitialised
  submodules and LFS pointers; a repo that needs them may see
  local failures in the worktree path. This is documented, not
  handled.
- Re-regenerating pre-flight evidence when council fix-loop
  iterations move `head_sha`. Evidence is captured once per
  invocation.

## Decisions

### D1: One `review_head_root` tool with `prepare` / `release`

It goes in `uf-workflow/index.ts` beside `checkout_pr_head`, and
reuses `GitExec`, `GhRunner`, `defaultGitExec`, and
`defaultGhRunner`. One tool with an `action` enum keeps the
pairing visible in the command text. No new `lib/` module: there
is one consumer.

Args:
- `action`: `prepare` | `release`.
- `pr_number`: positive int (prepare).
- `base_sha`, `head_sha`: `^[0-9a-f]{40}$` (prepare).
- `root`: string (release).

Success data for `prepare`:

```json
{"root":"/repo/.uf/review-head/k3v9x2qa","head_sha":"<40-hex>",
 "base_sha":"<40-hex>","method":"worktree","created":true,
 "config_sha":"<40-hex>","config_dirty":false,
 "changed_files":["internal/a.go"]}
```

- `method` is `checkout` (the current tree is reused,
  `created: false`) or `worktree`.
- `config_sha` is the plugin directory's `HEAD`: the tree that
  `.uf/preflight.yaml` and its scripts come from.
  `config_dirty` is true when that tree had uncommitted changes,
  so the config may differ from `config_sha`.
- Failure codes: `UNTRUSTED_HEAD`, `HEAD_UNAVAILABLE`,
  `HEAD_MOVED`, `WORKTREE_FAILED`, `INVALID_ROOT`, and
  `REVIEW_HEAD_ROOT_ERROR` for an unexpected error (the message
  carries the detail).
- `changed_files` is
  `git diff --name-only <base_sha>...<head_sha>`, sorted.

### D2: Worktrees live in `<cwd>/.uf/review-head/<id>`

The location is in the workspace, beside the
`.uf/sibling-clones/` precedent. This avoids OpenCode's
out-of-workspace permission asks, which parented child sessions
cannot answer. `<id>` is 8 random characters matching
`^[a-z0-9]{8}$`.

On first use the tool writes `.uf/review-head/.gitignore` with
`*`. The directory is then ignored even in repos whose managed
`.gitignore` block predates this change: it never shows in
`git status`, and it never trips the reuse check. Go's `./...`
already skips dot-directories.

Because the location is fixed, root validation needs no
in-process registry, and a `release` from a restarted process
still works.

### D3: Root validation is exact

A root is valid for `release`, and for the declared-checks
`root` argument, only when all of these hold:

- It is the exact string `<cwd>/.uf/review-head/<id>`, with `<id>`
  matching the pattern. There is no normalisation and no prefix
  match, so `..`, a trailing slash, or a symlinked spelling is
  refused.
- `realpath(root)` equals that same string resolved through
  `realpath(cwd)`.
- `git worktree list --porcelain` (from `cwd`) lists it, compared
  by `realpath`.

Anything else returns `INVALID_ROOT` and deletes nothing. The
plugin directory itself is also accepted, for the `checkout`
method.

### D4: `prepare` decision sequence

1. Validate inputs, then prune stale worktrees. Each
   `.uf/review-head/<id>` whose directory mtime is older than 24
   hours, and whose `realpath` appears in
   `git worktree list --porcelain`, is removed with
   `git worktree remove --force -- <path>`, and the directory is
   then deleted. Pruning is best effort and never fails `prepare`.
2. Reuse check: if `git rev-parse HEAD` equals `head_sha` and
   `git status --porcelain --untracked-files=normal
   --ignore-submodules=none` is empty
   (ignored files do not count), return
   `{root: cwd, method: "checkout", created: false}`. The user
   chose this checkout, so no trust check applies.
3. Trust gate. Read `gh pr view <N> --json url,headRefOid,
   isCrossRepository,headRepository,headRepositoryOwner,author`.
   This is the same repository resolution the commands used to
   freeze `head_sha`; the base repo comes from the PR URL, never
   from a `{owner}/{repo}` placeholder, which can resolve to the
   user's fork.
   - `headRefOid` MUST equal `head_sha`, or `prepare` fails
     `HEAD_MOVED` before any fetch or worktree. This binds the
     vetted PR to the commit that will run: a trusted author's
     PR cannot vouch for another PR's head that happens to be
     local.
   - The head is trusted when it is not cross-repository. A null
     head repository (deleted fork) is never same-repo.
   - Otherwise it is trusted only when
     `gh api repos/<base>/collaborators/<login>/permission`
     reports `admin`, `maintain`, or `write`: people who can
     already push to the base repo. `author_association` is not
     used, because `COLLABORATOR` includes read-only invitees and
     `MEMBER` includes any org member.
   - A failed permission lookup counts as untrusted (fail closed).
   - Untrusted: fail `UNTRUSTED_HEAD`. No fetch happens and no
     worktree is created.
   - A failure of the PR lookup itself is `HEAD_UNAVAILABLE`,
     retryable.
4. Ensure both commits exist. For each of `head_sha` and
   `base_sha`, run `git cat-file -e <sha>^{commit}`. If one is
   missing:
   - Pick the first remote whose `git remote get-url` matches the
     base repo by host plus `owner/repo`, case-insensitively.
     Accept `git@host:o/r(.git)`, `ssh://[user@]host/o/r(.git)`,
     and `https://[userinfo@]host/o/r(.git)`.
   - Run `git fetch --no-tags <remote> refs/pull/<N>/head` for
     the head, or `<base_sha>` for the base.
   - Re-check with `cat-file -e`. `FETCH_HEAD` is never read, so
     concurrent reviews cannot race on it.
   - Still missing after a successful fetch: `HEAD_MOVED` (the
     PR was force-pushed and the frozen commit is gone; re-run
     the review). No matching remote, or a fetch error:
     `HEAD_UNAVAILABLE` (retryable for fetch errors).
5. Create `.uf/review-head/<id>` with
   `git worktree add --detach -- <path> <head_sha>`. On error,
   remove the directory and return `WORKTREE_FAILED` with git's
   stderr.
6. `changed_files` is computed before step 5, so a diff
   failure cannot leave a worktree behind; return.

### D5: `release`

- `root` equal to the plugin directory: a no-op success.
- `root` that passes D3: run
  `git worktree remove --force -- <root>`. If that fails, delete
  the directory anyway (`git worktree prune` reclaims the entry,
  as in `classifyFailures`). Success.
- Anything else: `INVALID_ROOT`.
- A second release of the same root fails D3 and returns
  `INVALID_ROOT`. The commands treat every release failure as a
  non-fatal warning (D7).

### D6: Declared checks get an execution root and a base

`uf_preflight_declared_checks` gains optional `root` and
`base_sha`. `runDeclaredChecks(cwd, deps, mode, execRoot = cwd,
baseSha?)`:

- `.uf/preflight.yaml`, `run` executables, `requires` probes, and
  `UF_PREFLIGHT_PROJECT_ROOT` resolve from `cwd`, the config root.
- Checks execute with `cwd = execRoot`.
- With `baseSha`, the diff scope is
  `git merge-base <baseSha> HEAD` in `execRoot`, plus the diff
  from it, skipping default-branch resolution. That merge-base
  also feeds the soft-gate baseline. Without `baseSha`,
  `computeDiffScope` is unchanged.
- `root` must pass D3, else `INVALID_ROOT` with no check run.
- Omitting both arguments is byte-for-byte today's behavior.

Config provenance differs by path. On the `checkout` path the
plugin directory *is* the PR head, so the PR's own
`.uf/preflight.yaml` applies. On the `worktree` path the config
comes from the user's current branch. Neither is hidden:
`config_sha` is reported (D8).

### D7: Skill and commands

The `pre-flight` skill takes optional `root` and `changed_files`
from the consuming command:

- Phases 1, 2, 4, and the Phase 4a worktree fallback: every shell
  command runs with working directory `root`.
- Phase 4a, when `base_sha` is supplied: Tier 1 reads CI check
  runs for `base_sha` instead of the default branch tip. Tier 2
  adds its worktree at `git merge-base <base_sha> HEAD`, run in
  `root`. A stacked PR is then compared against its own base.
- Phase 2a: when `changed_files` is supplied, it *is* the branch
  diff. The skill's own merge-base computation and its
  fallbacks do not run.
- Phase 2b: passes `root` and `base_sha` to the declared-checks
  tool, and handles `INVALID_ROOT` like `INVALID_PREFLIGHT_CONFIG`
  (a pre-flight FAIL, no declared check ran).
- Phase 5 adds the provenance line (D8).

The command flow:

- `prepare` runs once per invocation: in `/uf.review-pr` after
  Step 3, and in `/uf.review-council` at the start of Code Review
  Mode step 1 for PR input only. Local and uncommitted council
  input never calls the tool.
- `release` runs once, right after the last local gate (the end
  of `/uf.review-pr` Step 3.7, or council Phase 1b), and before
  any STOP after a successful `prepare`. Nothing later reads the
  root: the light check and the reviewers read the PR diff
  through `gh`. Releasing early keeps the worktree's lifetime
  short and spares every later STOP path its own release. A
  release failure is a warning in the report and never changes
  the verdict.
- Phase 1b (Gaze) runs only when `method` is `checkout`. For a
  worktree root it notes "Gaze skipped: the PR head is in a
  temporary worktree and gaze-reporter has no root argument" and
  continues, like the existing not-installed path. A package
  pattern cannot stand in for a root, because Go rejects absolute
  patterns outside the main module.

`prepare` failures:

| Code | `/uf.review-pr` | `/uf.review-council` |
|---|---|---|
| `UNTRUSTED_HEAD` | Skip local gates and continue | Coverage `FAIL`, `INCONCLUSIVE`, STOP |
| `HEAD_MOVED` | Skip local gates and continue | Coverage `FAIL`, `INCONCLUSIVE`, STOP |
| `HEAD_UNAVAILABLE` | Skip local gates and continue | Coverage `FAIL`, `INCONCLUSIVE`, STOP |
| `WORKTREE_FAILED` | Skip local gates and continue | Coverage `FAIL`, `INCONCLUSIVE`, STOP |

- `/uf.review-pr` records
  `Local gates: not run (<code>): <message>` and continues with CI
  evidence. In `ci-aware` mode CI is already the primary gate, and
  blocking the review on a missing remote would be a regression.
- `/uf.review-council` keeps Phase 1a mandatory. Every code,
  including `UNTRUSTED_HEAD`, takes the existing Phase 1a FAIL
  path: non-child finalization, coverage `FAIL`, `INCONCLUSIVE`.
  The council has no CI-evidence fallback, so skipping its only
  local gate would let a PR reach APPROVE with no gate evidence.
  That weakens a gate (AGENTS.md Gatekeeping).
- For an untrusted fork PR, the user opts in by checking out the
  head themselves (`gh pr checkout <N>`) and re-running. The reuse
  path then applies, just as running gates on a checkout of the
  user's choosing does today. The `UNTRUSTED_HEAD` message names
  this step.

### D8: Provenance

Both reports' pre-flight section carries:

```
**Gates ran at**: <head_sha> (<method>); config from <config_sha>
```

or the `Local gates: not run (...)` line. Gate evidence now names
the commit it describes and the commit its rules came from.

### D9: Determinism classification

Mechanical (now in tools):
- head identity and the dirty check
- the trust read
- remote matching, fetch, and commit verification
- worktree add, remove, and prune
- root validation
- the diff list

Judgment or agent-shell (stays in prose):
- Running the Phase 4 detected tools, which must run in the
  agent's shell with `root` as the working directory.
- The per-code handling choice.

The prepare → gates → release chain is the minimal chain DE-005
allows: the middle step cannot move into a tool without moving
the whole skill.

## Risks / Trade-offs

- **Slower first worktree run.** `make check` installs plugin
  dependencies in a fresh worktree. Accepted: correct evidence
  over fast wrong evidence. The `checkout` path costs nothing.
- **Trusted PR code runs locally.** This is unchanged in kind for
  same-repo and collaborator PRs. Fork PRs from others are gated
  by D4 step 3.
- **Leaked worktrees** if the agent dies between prepare and
  release. They are bounded by 24h pruning and hidden by the
  self-ignoring directory.
- **The trust gate costs one extra API call** for fork PRs. A
  token that cannot read collaborator permissions makes every fork
  PR untrusted. That is fail closed, and the message names the
  opt-in.
- **#1063 still applies** to non-PR input. PR input sidesteps it
  through `base_sha`.
