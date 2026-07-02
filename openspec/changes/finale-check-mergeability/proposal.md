## Why

When `/finale` runs `gh pr checks --watch`, the command may
return "no checks reported" even though CI workflows exist
and are configured for the repository. This happens when a
merge conflict blocks GitHub from queuing check runs. The
agent currently accepts this at face value, fabricates a
plausible but incorrect explanation, and proceeds to the
summary step -- leaving the user to discover the merge
conflict manually on the GitHub PR page.

This was observed on PR unbound-force/gaze#169 where a
conflict in `CLAUDE.md` silently blocked all 4 required
checks. The `gh pr checks` CLI conflates "no runs started
yet" with "no checks configured," and the `/finale` command
has no guidance to distinguish the two.

Fixes: https://github.com/unbound-force/unbound-force/issues/291

## What Changes

Add a mergeability check to `/finale` step 6 (Watch CI
Checks) that detects when merge conflicts or other
conditions prevent checks from starting. When `gh pr checks`
returns "no checks reported," the agent will investigate
rather than silently proceeding.

## Capabilities

### New Capabilities
- `mergeability-gate`: Before concluding that no CI checks
  exist, `/finale` queries PR mergeability status and
  cross-references against known workflow files to
  distinguish "checks not configured" from "checks blocked
  by conflict."
- `conflict-recovery`: When a merge conflict is detected,
  the agent reports which files conflict and offers to
  rebase the branch onto the target, re-push, and re-check.

### Modified Capabilities
- `ci-check-watch`: Step 6 of `/finale` gains a pre-check
  for PR mergeability and a fallback investigation path when
  no checks are reported.

### Removed Capabilities
- None

## Impact

- **File**: `.opencode/commands/finale.md` -- step 6 (Watch
  CI Checks) is extended with mergeability detection and
  conflict recovery logic.
- **Behavioral**: Agents running `/finale` will no longer
  silently skip CI when checks are blocked by merge
  conflicts. Instead, they will report the blocker and
  offer remediation.
- **Cross-repo**: This change applies to any repo using the
  `/finale` command from the unbound-force scaffold. All
  hero repos benefit without needing repo-specific changes.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

This change modifies agent instructions (a slash command
definition), not inter-hero artifact communication. No
artifact formats or hero interfaces are affected.

### II. Composability First

**Assessment**: PASS

The `/finale` command remains self-contained. The
mergeability check uses `gh` CLI capabilities already
available in the environment. No new hero dependencies
are introduced.

### III. Observable Quality

**Assessment**: PASS

The change improves observability by surfacing the actual
blocker (merge conflict) instead of allowing the agent to
fabricate an incorrect explanation. The agent will report
concrete mergeability status rather than silently accepting
ambiguous output.

### IV. Testability

**Assessment**: N/A

This change modifies agent instructions (Markdown), not
executable code. No Go source files are affected. The
behavior is verified by the agent following the updated
instructions during `/finale` execution.

### V. Security by Default

**Assessment**: N/A

No new dependencies, inputs, or privilege escalations are
introduced. The change uses existing `gh` CLI commands with
the same permissions already required by `/finale`.
<!-- scaffolded by uf vdev -->
