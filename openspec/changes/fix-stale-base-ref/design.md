## Context

Local `/uf.review-council` runs compute the review diff using
`git diff main...HEAD` with the local `main` ref. When local
`main` is stale (behind remote-tracking refs like `upstream/main`
or `origin/main`), the diff includes already-merged commits as
phantom changes. This inflates the change profile, dispatches
agents at higher tiers than necessary, and produces findings on
code the author did not change.

The `uf-workflow` plugin scaffold already exists
(`.opencode/plugins/uf-workflow/index.ts`) with an empty tool
registry. The `review-dispatch` plugin provides a proven pattern
for tool registration, Zod argument schemas, and testable factory
functions.

## Goals / Non-Goals

### Goals
- Eliminate phantom files from local review diffs by selecting
  the most current available base ref
- Provide a reusable `resolve_base_ref` tool in the uf-workflow
  plugin for any command that needs a local diff base
- Abort hard on failure (no resolvable ref) with a clear error
- Remove the token-wasting Key Files Changed block from
  uf.unleash Step 10 demo output

### Non-Goals
- Network access (git fetch) -- the tool uses only cached
  remote-tracking refs via `git rev-parse --verify`
- Changing `uf.review-pr` or `uf.address-feedback` -- these
  already use PR-provided `baseRefOid` from the GitHub API
- Adding a user-configurable base ref override
- Changing the diff algorithm or three-dot merge-base semantics

## Decisions

**D1: Three-level fallback chain (not configurable)**

The tool tries refs in this order:
1. `upstream/main` -- fork checkouts
2. `origin/main` -- direct clones
3. `main` -- local-only fallback

Rationale: This order matches the most common git workflows.
Fork checkouts name the canonical repo `upstream`; direct
clones use `origin`. The local `main` is always the least
current option. No configuration is needed because the
fallback is deterministic and correct for all known workflows.

**Security invariant**: The candidate ref list MUST remain a
closed set of string literals — never user-supplied values.
This prevents ref injection attacks where a malicious ref name
could alter diff output or resolve to an unintended commit.

**D2: Hard abort on total failure**

If none of the three refs resolve, the tool returns a failure
result with `retryable: false`. The consuming command MUST
abort the review -- proceeding with no valid base would
produce meaningless output.

**D3: Factory function pattern for testability**

The tool logic lives in a `createResolveBaseRefTool(exec)`
factory that accepts an executor function for subprocess
calls. Tests inject a mock executor; the plugin wires in a
real `child_process.execSync` wrapper. This matches the
existing `review-dispatch` pattern and satisfies
constitution principle IV (Testability).

**D4: Structured output with provenance**

The tool returns `{ ref, sha, source }` where `source` is
one of `upstream/main`, `origin/main`, or `main`. This
satisfies constitution principle III (Observable Quality) --
the consuming command can announce which ref was selected
and why before computing the diff.

**D5: Remove Key Files Changed from uf.unleash Step 10**

The `## Key Files Changed` section and the `git diff
--name-only main...HEAD` instruction that feeds it are
removed entirely. The section adds token cost without
actionable value in the demo output -- the user already has
the full diff context from the review.

## Risks / Trade-offs

**Risk: Remote-tracking refs may also be stale**

If the user hasn't fetched in a long time, `upstream/main`
could itself be behind the true upstream. This is strictly
better than using local `main` (which is even more stale)
but not perfect. Mitigation: the tool announces its selected
ref and SHA, making staleness visible. A future enhancement
could optionally fetch before resolving.

**Risk: Non-standard remote names**

Some users name their fork remote something other than
`origin` or their upstream something other than `upstream`.
The three-level fallback covers the standard naming
conventions. Non-standard setups fall through to local
`main`, which is the current behavior -- no regression.

**Trade-off: No user configuration**

Keeping the fallback hardcoded simplifies the implementation
and avoids adding config surface. Users with non-standard
remote names get current behavior (local `main`). If demand
arises, a configurable ref list can be added later without
breaking the existing contract.
