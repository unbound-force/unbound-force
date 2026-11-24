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
   - `prepare`: given `pr_number` and `head_sha`, it returns the
     directory where the local gates must run.
     - If `HEAD` already equals `head_sha` and the tracked tree is
       clean, the current checkout is reused.
     - Otherwise it fetches the head commit if missing, verifies
       the SHA, and adds a detached worktree at `head_sha` under a
       tool-owned temp prefix.
   - `release`: removes a worktree that `prepare` created, and
     nothing else.
   - Failure codes: `HEAD_UNAVAILABLE`, `HEAD_MOVED`,
     `WORKTREE_FAILED`, `INVALID_ROOT`.
2. **`uf_preflight_declared_checks` gains an optional `root`
   argument.** Declared checks, diff scope, and the soft-gate
   baseline then run against that root instead of the plugin
   directory. Config and scripts still come from the primary
   checkout, as the baseline already does.
   - A `root` that is neither the plugin directory nor a live
     `review_head_root` worktree is refused with `INVALID_ROOT`.
3. **`pre-flight` skill**: Phase 4 runs detected tools from the
   consuming command's root when one is supplied, and Phase 2b
   passes it to the declared-checks tool. Phase 5 output records
   the commit the gates ran against.
4. **`/uf.review-pr`** calls `prepare` before Step 3.7, runs
   pre-flight with the returned root, and calls `release` before
   finalization on every path.
5. **`/uf.review-council`** does the same around Phases 1a and 1b
   for PR input only. Local and uncommitted input is unchanged.
   - Phase 1b runs `gaze-reporter` only when `prepare` returned
     `method: "checkout"` (the current tree is the head).
   - For a worktree root, Gaze is skipped with an informational
     note. `gaze-reporter` is scaffolded by the gaze repo and has
     no root argument, and running it from the primary checkout
     would score the wrong code.
6. **Fail closed.** If `prepare` fails, neither command runs local
   gates against the wrong tree.
   - `/uf.review-pr` stops with the code-specific remediation,
     consistent with its Execution Mode Check.
   - `/uf.review-council` records coverage `FAIL` and an
     `INCONCLUSIVE` verdict through the shared finalization
     protocol, then stops.
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
- Behavior: when the user is not on the PR head, a PR review now
  creates a temporary worktree. A first `make check` there also
  installs plugin dependencies, so it is slower but correct.
- Follow-up in unbound-force/gaze: an analysis-root argument for
  `gaze-reporter`, so Gaze can run on a worktree root.
- Related, not addressed: #1063 (diff scope when a branch is behind
  main). Running from the head root inherits that behavior
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
