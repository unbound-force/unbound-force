## Context

The `/finale` command (`.opencode/commands/finale.md`) step 6
runs `gh pr checks <number> --watch` to monitor CI status. The
`gh` CLI returns "no checks reported" in multiple situations:

1. No CI workflows are configured for the repository
2. Checks exist but have not been queued yet (e.g., merge
   conflict blocks GitHub from starting runs)
3. A race condition where checks have not propagated yet

The current instructions treat all three cases identically:
accept the output and proceed. This leads to silent failures
when merge conflicts prevent CI from running.

## Goals / Non-Goals

### Goals
- Detect when "no checks reported" is caused by a merge
  conflict rather than absent CI configuration
- Surface the actual blocker to the user with actionable
  remediation steps (rebase, resolve, re-push)
- Cross-reference workflow files to validate whether CI
  is genuinely unconfigured vs. temporarily blocked
- Keep the fix contained to step 6 of `/finale` without
  restructuring other steps

### Non-Goals
- Handling all possible reasons checks might not start
  (e.g., GitHub outages, rate limits, workflow syntax
  errors) -- only merge conflicts are addressed
- Auto-resolving merge conflicts -- the agent offers to
  rebase but conflict resolution requires human judgment
  for non-trivial cases
- Modifying the `gh` CLI behavior or wrapping it in a
  script -- we work within the existing CLI capabilities
- Adding retry/polling for check propagation delays --
  `gh pr checks --watch` already handles this for the
  normal case

## Decisions

### D1: Mergeability pre-check before interpreting check results

Insert a mergeability query before acting on `gh pr checks`
output. When checks return "no checks reported":

```bash
gh pr view <number> --json mergeable,mergeStateStatus
```

The `mergeable` field has three values:
- `MERGEABLE` -- no conflicts, checks genuinely absent
- `CONFLICTING` -- merge conflict blocks check execution
- `UNKNOWN` -- GitHub is still computing (rare, transient)

This approach uses data already available via the GitHub
API without adding new tools or dependencies. Aligns with
the proposal's Constitution Alignment (Composability First:
no new dependencies).

### D2: Workflow file cross-reference

When mergeability is `MERGEABLE` but no checks are reported,
the agent SHOULD check for workflow files:

```bash
ls .github/workflows/*.yml .github/workflows/*.yaml \
  2>/dev/null
```

If workflow files with `pull_request` triggers exist, "no
checks reported" is anomalous and the agent SHOULD warn
the user rather than silently proceeding. This catches edge
cases beyond merge conflicts (e.g., workflow syntax errors,
disabled workflows).

### D3: Conflict recovery flow

When `mergeable` is `CONFLICTING`:

1. Report the conflict status
2. Show which files conflict (from merge attempt output)
3. Offer three options:
   a. Rebase onto target branch and re-push
   b. Stop and let the user resolve manually
   c. Continue anyway (with explicit warning)

The rebase option uses:
```bash
git fetch origin main
git rebase origin/main
```

If rebase succeeds cleanly, force-push and re-check:
```bash
git push --force-with-lease
gh pr checks <number> --watch
```

If rebase has conflicts, report them and stop -- automated
conflict resolution is out of scope.

### D4: Placement within step 6

The mergeability check is added as a sub-step within the
existing step 6, not as a new top-level step. This keeps
the overall `/finale` structure unchanged and limits the
blast radius of the change.

## Risks / Trade-offs

### R1: GitHub API latency for mergeability computation

GitHub sometimes returns `UNKNOWN` for `mergeable` while
computing the merge state. The design handles this by
treating `UNKNOWN` as "proceed with caution" and warning
the user. Acceptable because `UNKNOWN` is transient and
resolves within seconds.

### R2: Rebase may introduce new conflicts

Offering to rebase is a convenience, not a guarantee. If
the rebase itself fails, the agent stops and defers to
the user. This is the correct behavior -- automated
conflict resolution for non-trivial conflicts is
unreliable.

### R3: False positive on workflow detection

A repository might have workflow files that don't trigger
on `pull_request` events (e.g., `workflow_dispatch` only).
The cross-reference (D2) uses a heuristic (presence of
workflow files), not a deep parse of trigger conditions.
This is acceptable because the agent warns rather than
blocks -- a false positive results in an informational
message, not a failure.
<!-- scaffolded by uf vdev -->
