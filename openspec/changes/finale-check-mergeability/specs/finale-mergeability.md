## ADDED Requirements

### FR-001: Mergeability gate on empty check results

When `gh pr checks` returns "no checks reported," the
agent MUST query PR mergeability before concluding that
no CI checks are configured.

```bash
gh pr view <number> --json mergeable,mergeStateStatus
```

The agent MUST interpret the `mergeable` field as follows:

- `CONFLICTING`: Report the conflict and offer recovery
  options (FR-002). MUST NOT proceed to the summary step.
- `UNKNOWN`: Warn the user that GitHub is still computing
  mergeability. SHOULD retry after a brief delay (5-10
  seconds). If still `UNKNOWN` after one retry, warn and
  proceed.
- `MERGEABLE`: Proceed to workflow cross-reference
  (FR-003).

If the `gh pr view` command itself fails (network error,
authentication failure, or PR not found), the agent MUST
report the error and **STOP**, consistent with the
existing `/finale` error handling pattern (e.g., step 4:
"If push fails: report error and STOP").

#### Scenario: Mergeability query fails

- **GIVEN** a PR has been created
- **WHEN** `gh pr checks` returns "no checks reported"
- **AND** `gh pr view --json mergeable,mergeStateStatus`
  fails with a non-zero exit code
- **THEN** the agent reports the error output to the user
- **AND** stops with instructions: "Could not query PR
  mergeability. Check network connectivity and `gh` CLI
  authentication."

#### Scenario: Merge conflict blocks CI checks

- **GIVEN** a PR has been created with `gh pr create`
- **AND** the PR has a merge conflict with the target
  branch
- **WHEN** `gh pr checks` returns "no checks reported"
- **THEN** the agent queries `gh pr view --json
  mergeable,mergeStateStatus`
- **AND** detects `mergeable: "CONFLICTING"`
- **AND** reports the merge conflict to the user
- **AND** offers recovery options instead of proceeding
  to the summary step

#### Scenario: Mergeability state unknown

- **GIVEN** a PR has been created
- **WHEN** `gh pr checks` returns "no checks reported"
- **AND** `gh pr view` returns `mergeable: "UNKNOWN"`
- **THEN** the agent warns the user that mergeability is
  being computed
- **AND** retries the mergeability check once after a
  brief delay
- **AND** if still `UNKNOWN`, warns and proceeds with
  caution

### FR-002: Conflict recovery options

When a merge conflict is detected (FR-001, `CONFLICTING`),
the agent MUST present the user with recovery options:

1. **Rebase onto target branch**: Run `git fetch origin
   main && git rebase origin/main`. If rebase succeeds
   cleanly, force-push with `git push --force-with-lease`
   and re-run `gh pr checks --watch`. If rebase has
   conflicts, report them and stop.
2. **Stop and resolve manually**: Report the conflict and
   let the user handle it outside the agent session.
3. **Continue anyway**: Proceed to the summary step with
   an explicit warning that CI has not run.

The agent MUST NOT automatically rebase without user
confirmation. The agent MUST NOT use `git push --force`
(MUST use `--force-with-lease` for safety).

#### Scenario: Clean rebase resolves conflict

- **GIVEN** a PR has a merge conflict
- **AND** the user selects "Rebase onto target branch"
- **WHEN** `git rebase origin/main` succeeds without
  conflicts
- **THEN** the agent runs `git push --force-with-lease`
- **AND** re-runs `gh pr checks <number> --watch`
- **AND** continues with normal step 6 check-watching
  behavior

#### Scenario: Rebase has conflicts

- **GIVEN** a PR has a merge conflict
- **AND** the user selects "Rebase onto target branch"
- **WHEN** `git rebase origin/main` encounters conflicts
- **THEN** the agent runs `git rebase --abort`
- **AND** reports which files have conflicts
- **AND** stops with instructions for manual resolution

#### Scenario: User chooses to continue without CI

- **GIVEN** a PR has a merge conflict
- **AND** the user selects "Continue anyway"
- **THEN** the agent proceeds to step 7
- **AND** the summary (step 8) includes a warning:
  "CI checks did not run due to merge conflict"

### FR-003: Workflow file cross-reference

When `mergeable` is `MERGEABLE` and `gh pr checks` returns
"no checks reported," the agent SHOULD check for workflow
files:

```bash
ls .github/workflows/*.yml .github/workflows/*.yaml \
  2>/dev/null
```

- If workflow files exist: warn the user that CI workflows
  are present but no checks ran. This MAY indicate a
  workflow syntax error, disabled workflows, or other
  configuration issue.
- If no workflow files exist: accept that no CI is
  configured and proceed normally.

The agent MUST NOT parse workflow file contents to
determine trigger conditions. The presence of workflow
files is sufficient for a warning.

#### Scenario: Workflows exist but no checks ran

- **GIVEN** a PR with `mergeable: "MERGEABLE"`
- **AND** `gh pr checks` returns "no checks reported"
- **AND** `.github/workflows/` contains workflow files
- **WHEN** the agent checks for workflow files
- **THEN** the agent warns: "CI workflow files exist but
  no checks were reported. This may indicate disabled
  workflows or a configuration issue."
- **AND** asks the user whether to proceed or investigate

#### Scenario: No workflow files configured

- **GIVEN** a PR with `mergeable: "MERGEABLE"`
- **AND** `gh pr checks` returns "no checks reported"
- **AND** `.github/workflows/` contains no workflow files
- **WHEN** the agent checks for workflow files
- **THEN** the agent reports "No CI workflows configured"
- **AND** proceeds to step 7 normally

## MODIFIED Requirements

### Requirement: Step 6 Watch CI Checks

The existing step 6 instruction to run
`gh pr checks <number> --watch` and interpret results
is modified to add the mergeability gate (FR-001) as a
sub-step when "no checks reported" is returned.

Previously: Step 6 ran `gh pr checks --watch` and
interpreted the result as either "checks pass" (proceed)
or "checks fail" (report and stop). "No checks reported"
was treated as "checks pass."

Now: "No checks reported" triggers the mergeability gate
(FR-001) before any conclusion is drawn. The existing
"checks pass" and "checks fail" paths are unchanged.

## REMOVED Requirements

None.
<!-- scaffolded by uf vdev -->
