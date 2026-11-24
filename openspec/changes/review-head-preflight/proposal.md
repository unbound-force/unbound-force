## Why

`/uf.review-pr <N>` and `/uf.review-council <N>` pin the diff,
review context, and reviewer prompts to the PR's immutable
`base_sha...head_sha`, and both say "Never substitute the current
checkout". Their local quality gates do not follow that rule:

- `/uf.review-pr` Step 3.7 runs the `pre-flight` skill in
  `ci-aware` mode from the agent's working directory.
- `/uf.review-council` Phase 1a runs `pre-flight` in `soft-gate`
  mode, and Phase 1b runs `gaze-reporter`, from the same
  directory. The command is also told to fetch the PR's commits
  "without checking out or switching branches".

Nothing checks that the working tree is at `head_sha`. When it is
not (usually `main`), the local results describe the wrong code:

- A broken PR gets a local PASS. In `/uf.review-pr` that also
  drops the category from the AI review.
- In `/uf.review-council`, a failure that belongs to the local
  branch is classified branch-caused against the `main` baseline.
  The council then stops CRITICAL/INCONCLUSIVE on a PR that did not
  cause it.
- Gaze scores the local code and reports it as the PR's.

Issues #1203 (`/uf.review-pr`) and #1204 (`/uf.review-council`)
record the gap. The no-argument council run from `/uf.unleash`
Step 6 is unaffected: there the checkout is the code under review.

## What Changes

1. **New `review_head_root` tool** in the `uf-workflow` plugin, with
   injectable git and gh executors (argument arrays, no shell).
   - `prepare` takes `pr_number`, `base_sha`, and `head_sha`. It
     returns the directory where the local gates must run, the PR's
     `changed_files` (`base_sha...head_sha`), and the config commit.
   - If `HEAD` already equals `head_sha` with no tracked or
     untracked changes, the current checkout is reused.
   - Otherwise it applies a trust gate: a fork PR from an author
     who is not OWNER, MEMBER, or COLLABORATOR fails
     `UNTRUSTED_HEAD` before any fetch.
   - For a trusted head it fetches missing commits from the base
     repo's remote, verifies them, and adds a detached worktree
     under a self-ignoring `.uf/review-head/<id>`.
   - `release` removes only such worktrees. Root validation is
     exact.
   - Failure codes: `UNTRUSTED_HEAD`, `HEAD_UNAVAILABLE`,
     `HEAD_MOVED`, `WORKTREE_FAILED`, `INVALID_ROOT`.
2. **`uf_preflight_declared_checks` gains optional `root` and
   `base_sha`.** Declared checks run in the root and are scoped to
   `merge-base(base_sha, HEAD)`. Config and scripts still come
   from the plugin directory. An invalid root returns
   `INVALID_ROOT`.
3. **`pre-flight` skill**: given `root` and `changed_files`, every
   phase runs against the root, and Phase 2a uses `changed_files`
   as the branch diff:
   - workflow parsing
   - tool detection
   - execution
   - the baseline fallback
   - declared checks

   This closes the empty-diff-on-`main` false PASS. Phase 5
   records `Gates ran at: <head_sha> (<method>); config from
   <config_sha>`.
4. **`/uf.review-pr`** calls `prepare` once after Step 3 and threads
   its result into Step 3.7. If `prepare` fails, it records
   `Local gates: not run (<code>)` and continues on CI evidence.
   It calls `release` once at the end; a release failure is a
   warning.
5. **`/uf.review-council`**, for PR input only, does the same
   around Phase 1a.
   - `UNTRUSTED_HEAD` skips the local gates and continues.
   - Other `prepare` errors take the existing Phase 1a FAIL path:
     coverage `FAIL`, `INCONCLUSIVE`, and no reviewer children.
   - Phase 1b runs Gaze only when the checkout was reused.
     `gaze-reporter` is scaffolded by the gaze repo and has no
     root argument.
6. **`.uf/preflight.yaml` runtime-contract comment** (both copies)
   names the review-head root as the PR-mode working directory.
7. **Scaffold mirrors** stay byte-identical for both commands, the
   skill, the plugin, and `lib/preflight-declared.ts`.

## Capabilities

### New Capabilities
- `review-head-preflight`: local quality gates in PR-mode reviews
  run against the PR's frozen head commit, never an unrelated
  checkout.

### Modified Capabilities
- None. `review-council` requirements (discovery, verdict policy,
  fix loop) are unchanged. The Phase 1a/1b behavior moves into
  the new capability.

### Removed Capabilities
- None.

## Impact

- `.opencode/plugins/uf-workflow/index.ts`: new tool factory,
  plus the optional `root` on the declared-checks tool.
- `.opencode/lib/preflight-declared.ts`: `runDeclaredChecks` takes
  separate config and execution roots.
- `.opencode/commands/uf.review-pr.md` and
  `.opencode/commands/uf.review-council.md`: prepare/release calls,
  per-code handling, root threading.
- `.opencode/skills/pre-flight/SKILL.md`: root parameter.
- Scaffold mirrors under `internal/scaffold/assets/opencode/`.
- Tests: vitest units for the tool and for the declared-checks root,
  a real-git round trip, and Go command-contract clauses for both
  commands.
- Behavior: when the user is not on the PR head, a review of a
  trusted PR creates a worktree under `.uf/review-head/`. A first
  `make check` there also installs plugin dependencies, so it is
  slower but correct. Untrusted fork PRs no longer get local gates
  unless the user checks out the head themselves; until now they
  got gates that described the wrong code.
- `.uf/preflight.yaml` and its scaffold mirror: comment-only
  update.
- Follow-up in unbound-force/gaze: an analysis-root argument for
  `gaze-reporter`, so Gaze can run on a worktree root.
- Related: #1063 (diff scope when a branch is behind main). PR
  input sidesteps it through `base_sha`; non-PR input is
  unchanged.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

The review artifacts gain provenance: pre-flight output names the
commit the local gates ran against, so a reader can tell whether
the gate evidence matches the reviewed `head_sha`. No runtime
coupling is added between heroes.

### II. Composability First

**Assessment**: PASS

The new tool lives in the existing `uf-workflow` plugin, and `root`
is optional on the declared-checks tool. Callers that omit it keep
today's behavior. No change is made to the gaze-owned
`gaze-reporter` agent. Gaze stays optional and is skipped, with a
note, when it cannot target the head.

### III. Observable Quality

**Assessment**: PASS

Both tools return structured JSON envelopes with stable failure
codes. The gate results reported for a PR come from that PR's code,
which corrects quality evidence that is currently misattributed.

### IV. Testability

**Assessment**: PASS

The worktree tool takes injectable git and gh executors, so every
branch (reuse, fetch, create, each failure code, release) is unit
testable without a network. A real-git round trip verifies the
worktree side effects, and Go contract tests pin the command prose.
