## Context

The two PR-mode review commands freeze `head_sha` in their input
context. Their local gates, though, run wherever the agent happens
to be:

- `/uf.review-pr` Step 3.7: `pre-flight` in `ci-aware` mode.
- `/uf.review-council` Phase 1a: `pre-flight` in `soft-gate` mode.
- `/uf.review-council` Phase 1b: `gaze-reporter`.

The `pre-flight` skill runs detected tools (`make check`,
`go test`, ...) through the agent's shell. It also calls
`uf_preflight_declared_checks`, whose `cwd` is fixed at plugin
start (`input.worktree || input.directory`).

`preflight-declared.ts` already has the pattern this change needs.
`classifyFailures` adds a detached worktree at the merge-base, runs
the primary checkout's scripts with that worktree as `cwd`, and
removes it in a `finally`. This change applies the same pattern to
the PR head.

The proposal's Constitution Alignment is PASS on all four
principles. D3, D5, and D7 below are where Composability First and
Observable Quality show up concretely.

## Goals / Non-Goals

### Goals
- In PR mode, every local gate (detected tools, declared checks,
  baseline, Gaze) runs against `head_sha`, or does not run.
- The user's checkout is never switched, stashed, or modified.
- Worktree choreography lives in a tested tool, not command prose
  (Determinism gate, DE-001/DE-005).
- If the head cannot be materialised, the gates fail closed.

### Non-Goals
- Fixing #1063 (diff scope and baseline when a branch is behind
  main).
- Changing the `ci-aware` / `soft-gate` decision rules.
- The `/uf.review-pr` CI-fix branch flow (Steps 3b+), which checks
  out branches on purpose.
- Editing `gaze-reporter.md`. It is scaffolded by the gaze repo
  (`<!-- scaffolded by gaze dev -->`), so a root argument belongs
  there.
- Sandboxing PR code. Running local gates on a PR already executes
  its code. See Risks.

## Decisions

### D1: One `review_head_root` tool with `prepare` / `release`

It goes in `uf-workflow/index.ts` beside `checkout_pr_head`, and
reuses `GitExec`, `GhRunner`, `defaultGitExec`, and
`defaultGhRunner`. A single tool with an `action` enum keeps the
pairing visible in prose and lets one registry serve both actions.
No new `lib/` module: there is one consumer, and it would add a
scaffold asset.

Args: `action` (`prepare` | `release`), `pr_number` (positive int,
prepare), `head_sha` (`^[0-9a-f]{40}$`, prepare), `root` (release).

Success data:

```json
{"root":"/tmp/uf-review-head-1a2b3c/wt","head_sha":"<40-hex>",
 "method":"worktree","created":true}
```

`method` is `checkout` (current tree reused, `created: false`) or
`worktree`.

### D2: `prepare` decision sequence

1. Validate inputs. Then prune stale tool-owned worktrees: any
   `git worktree list --porcelain` entry under the
   `uf-review-head-` temp prefix whose directory is older than 24h
   is removed with `git worktree remove --force` (best effort).
2. `git rev-parse HEAD` equals `head_sha`, and
   `git status --porcelain --untracked-files=no` is empty: return
   `{root: cwd, method: "checkout", created: false}`.
3. `git cat-file -e <head_sha>^{commit}`. If the commit is missing:
   - Resolve the base repo with
     `gh pr view <N> --json url`.
   - Pick the first remote whose `git remote get-url` names the
     same `owner/repo`, for SSH or HTTPS, with or without `.git`.
   - `git fetch --no-tags <remote> refs/pull/<N>/head`.
   - `git rev-parse FETCH_HEAD` differs from `head_sha`:
     `HEAD_MOVED`. The PR was pushed after the review froze its
     input. The remediation is to re-run the review.
   - No matching remote, or the fetch failed: `HEAD_UNAVAILABLE`
     (`retryable: true` for fetch errors).
4. `mkdtemp(<os tmp>/uf-review-head-)`, then
   `git worktree add --detach <dir>/wt <head_sha>`. On error,
   remove the temp dir and return `WORKTREE_FAILED` with git's
   message.
5. Record `<dir>/wt` in the in-process live-root set and return it.

### D3: `release` removes only what `prepare` created

A `root` that is not in the live set is refused with
`INVALID_ROOT`. The user's checkout passed back from the
`checkout` method is a no-op success. Otherwise it runs
`git worktree remove --force <root>`, removes the temp parent,
and drops the root from the set. A failed `git worktree remove`
still removes the directory; `git worktree prune` reclaims the
entry, as in `classifyFailures`. The tool never deletes a path it
did not create, which keeps it composable with the user's own
worktrees.

### D4: Declared checks get an execution root, not a new config root

`uf_preflight_declared_checks` gains an optional
`root: string`. `runDeclaredChecks(cwd, deps, mode)` becomes
`runDeclaredChecks({configRoot, execRoot}, deps, mode)`:

- `.uf/preflight.yaml`, `run` executables, and `requires` probes
  resolve from `configRoot`, the plugin directory. They are
  project policy, and stealth `.uf/` trees are absent from
  worktrees. A PR therefore cannot rewrite the gate it is judged
  by.
- Checks execute with `cwd = execRoot`.
- `computeDiffScope` and the soft-gate merge-base use
  `execRoot`'s `HEAD`, which is `head_sha`.
- `PROJECT_ROOT_ENV` continues to name `configRoot`.

When `root` is omitted, `execRoot = configRoot` and behavior is
byte-for-byte today's. A `root` that is neither the plugin
directory nor a live `review_head_root` root returns
`INVALID_ROOT`. The two tools share the live set through a
module-level registry in `uf-workflow`, injected into both
factories for tests.

### D5: Skill and commands thread one value

The `pre-flight` skill takes an optional `root` from the consuming
command:

- Phase 4 runs each detected command with that working directory.
- Phase 2b passes `root` to the declared-checks tool.
- Phase 5 adds a `**Gates ran at**: <head_sha> (<method>)` line.

This is a single parameter, not choreography.

The command flow:

- `/uf.review-pr`: `prepare` runs right after Step 3 and before
  Step 3.7.
- `/uf.review-council`: `prepare` runs at the start of Code Review
  Mode step 1, for PR input only.
- In both commands, `release` runs before the finalization
  protocol on success, STOP, and INCONCLUSIVE paths alike. The
  prose states one rule ("release before finalizing whenever
  prepare succeeded") instead of repeating it at each stop.
- Phase 1b invokes `gaze-reporter` unchanged (`"full"`) only when
  `method` is `checkout`. For a `worktree` root it prints "Gaze
  skipped: the PR head is in a temporary worktree and
  gaze-reporter has no root argument" and continues without Gaze
  data, the same as the existing not-installed path. A package
  pattern cannot stand in for a root: Go rejects absolute patterns
  outside the main module.

### D6: Fail closed per command

- `/uf.review-pr`: `prepare` failure is a STOP, like the Execution
  Mode Check. A review without trustworthy local evidence does not
  meet the command's bar. Each code maps to one remediation line,
  e.g. `HEAD_MOVED` means re-run, and `WORKTREE_FAILED` suggests
  `gh pr checkout <N>` and re-running.
- `/uf.review-council`: `prepare` failure uses the existing Phase
  1a FAIL path. Non-child planning and finalization record
  coverage `FAIL` and `INCONCLUSIVE`; no children are dispatched.

### D7: Provenance

The `prepare` result (`head_sha`, `method`) is echoed in the
pre-flight section of both reports. That is the Observable Quality
hook: gate evidence now carries the commit it describes.

## Risks / Trade-offs

- **Slower first run in a worktree.** A fresh worktree has no
  `node_modules`, so `make check` installs plugin dependencies;
  Go's module cache is shared. This is accepted: correct evidence
  over fast wrong evidence. Users already on the PR head take the
  `checkout` path and pay nothing.
- **Executing PR code.** This is unchanged in kind. Today the
  local gates already run whatever is checked out; now they run
  the PR. Declared-check scripts still come from the primary
  checkout (D4). Reviewing untrusted fork PRs locally remains the
  user's call.
- **Leaked worktrees** if the agent dies between prepare and
  release. They are bounded by D2 step 1 pruning and by living
  under the OS temp dir.
- **Live-root set is per process.** A `release` in a different
  plugin process than `prepare` returns `INVALID_ROOT`. The stale
  pruning reclaims that worktree later. Accepted: both commands
  run prepare and release in one session.
- **#1063 still applies** to the head root when the PR branch is
  behind base. Out of scope; this change neither fixes nor worsens
  it.
