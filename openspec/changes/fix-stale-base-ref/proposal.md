## Why

Local `/uf.review-council` runs use `git diff main...HEAD` with the
local `main` ref as the diff base. When local `main` is stale (behind
`upstream/main` or `origin/main`), the diff includes already-merged
commits as phantom changes -- inflating the change profile, diluting
agent attention, and producing misleading findings.

**Observed incident** (fullsend-ai/agents#1611): local `main` was 15
commits behind `upstream/main`, producing 7 phantom files (18 vs 11
actual) and 37% inflated additions (+1,608 vs +1,015).

## What Changes

Replace the hardcoded `main` base ref in the local (no-PR) review
path with a three-level fallback that uses cached remote-tracking
refs (no network access):

1. `upstream/main` -- fork checkouts where upstream is the canonical
   repo
2. `origin/main` -- direct clones or forks without an `upstream`
   remote
3. `main` -- fallback when no remote-tracking ref exists

Additionally, remove the token-wasting `## Key Files Changed` display
block from the `/uf.unleash` Step 10 demo output.

## Capabilities

### New Capabilities
- `resolve_base_ref`: uf-workflow plugin tool that resolves the best
  available base ref for local diff operations. Returns the ref name,
  its resolved SHA, and the source label.

### Modified Capabilities
- `uf.review-council`: local review path calls `resolve_base_ref`
  instead of hardcoding `main` as the diff base.
- `uf.unleash`: Step 10 demo output drops the `## Key Files Changed`
  section to save tokens.

### Removed Capabilities
- None

## Impact

- **`.opencode/plugins/uf-workflow/index.ts`**: New `resolve_base_ref`
  tool registered in the existing scaffold.
- **`.opencode/test/uf-workflow.test.ts`**: Unit tests for
  `resolve_base_ref`.
- **`internal/scaffold/assets/opencode/commands/uf.review-council.md`**:
  Local review path updated to call `resolve_base_ref`.
- **`internal/scaffold/assets/opencode/commands/uf.unleash.md`**:
  Step 10 demo output trimmed (remove Key Files Changed block).

No Go code changes. No CI changes. No schema changes.

**Note**: The `review-dispatch` plugin (`index.ts`) may receive
minor cleanup or type import updates as a side-effect of shared
type alignment, but no behavioral changes.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

The `resolve_base_ref` tool produces a self-describing result
(ref, SHA, source label) that the consuming command can interpret
without consulting the tool again. The tool reads only local git
state -- no synchronous inter-hero communication required.

### II. Composability First

**Assessment**: PASS

The tool is an optional plugin registered in the uf-workflow
plugin. Commands that already use PR-provided base refs
(e.g. `uf.review-pr`) are unaffected. The tool introduces no
mandatory dependencies.

### III. Observable Quality

**Assessment**: PASS

The tool returns structured output with provenance (source label
indicating which ref was selected and why). The selected base ref
and SHA are announced before the diff is computed, making the
decision auditable.

### IV. Testability

**Assessment**: PASS

The tool's logic is a pure fallback chain over `git rev-parse
--verify` results. Unit tests can mock the git subprocess to
cover all three fallback branches and the error case.

### V. Security by Default

**Assessment**: PASS

The tool reads only local git state via `git rev-parse --verify`
— no network access, no user-supplied input to the ref candidates.
The candidate ref list (`upstream/main`, `origin/main`, `main`) is
a closed set of string literals, never derived from user input.
The resolved SHA is validated as a 40-character hex string before
returning.
