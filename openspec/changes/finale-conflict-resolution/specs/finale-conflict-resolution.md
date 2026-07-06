## ADDED Requirements

### FR-001: Sub-agent conflict resolution option

Step 6b of `/finale` MUST present a fifth option when a PR
has merge conflicts:

> 5. Spawn sub-agent to resolve conflicts (AI-assisted)

The option MUST appear after the existing four options. The
existing options (1-4) MUST NOT be modified.

#### Scenario: User selects sub-agent resolution

- **GIVEN** `/finale` detects `mergeable: CONFLICTING` on
  the PR
- **WHEN** the user selects option 5 (spawn sub-agent)
- **THEN** `/finale` executes `git fetch` and `git merge`
  to create conflict markers, identifies the conflicting
  files, and spawns a `cobalt-crush-dev` sub-agent with a
  prompt scoped to resolving those conflicts

### FR-002: Sub-agent receives scoped conflict context

The sub-agent prompt MUST include:

- The list of conflicting files (from
  `git diff --name-only --diff-filter=U`)
- The target branch name
- Instructions to resolve conflict markers in each file
- Instructions to stage resolved files with `git add`
- A directive to report per-file success or failure

The sub-agent MUST NOT receive the full `/finale` flow
context. It SHALL receive only the information needed for
conflict resolution.

#### Scenario: Sub-agent receives correct file list

- **GIVEN** `git merge <target>` produces conflicts in
  `fileA.md` and `fileB.go`
- **WHEN** the sub-agent is spawned
- **THEN** the prompt includes both `fileA.md` and
  `fileB.go` as files to resolve

### FR-003: User approval gate after resolution

After the sub-agent completes, `/finale` MUST show the user
the staged diff of all resolved files:

```bash
git diff --cached
```

The user MUST be asked to approve, request edits, or abort
before any commit is made.

#### Scenario: User approves sub-agent resolution

- **GIVEN** the sub-agent has resolved all conflicts and
  staged the files
- **WHEN** `/finale` shows the cached diff to the user
- **THEN** the user is presented with options to approve,
  edit, or abort
- **AND** if the user approves, `/finale` completes the
  merge commit and pushes

#### Scenario: User rejects sub-agent resolution

- **GIVEN** the sub-agent has resolved conflicts
- **WHEN** the user rejects the resolution (selects abort)
- **THEN** `/finale` runs `git merge --abort` to restore
  the pre-merge state
- **AND** `/finale` returns to the conflict recovery
  options menu (options 1-5)

### FR-004: Sub-agent failure handling

If the sub-agent cannot resolve all conflicts, `/finale`
MUST:

1. Report which files were resolved and which remain
   conflicted
2. Run `git merge --abort` to restore clean state
3. Present the conflict recovery options menu again
   (options 1-5)

#### Scenario: Sub-agent partially resolves conflicts

- **GIVEN** conflicts exist in `fileA.md` and `fileB.go`
- **WHEN** the sub-agent resolves `fileA.md` but fails on
  `fileB.go`
- **THEN** `/finale` reports: "Sub-agent resolved 1 of 2
  files. Unresolved: fileB.go"
- **AND** `/finale` runs `git merge --abort`
- **AND** the conflict recovery options are shown again

#### Scenario: Sub-agent fails completely

- **GIVEN** conflicts exist in multiple files
- **WHEN** the sub-agent cannot resolve any conflicts
- **THEN** `/finale` reports: "Sub-agent could not resolve
  any conflicts"
- **AND** `/finale` runs `git merge --abort`
- **AND** the conflict recovery options are shown again

### FR-005: Merge-based conflict resolution

The sub-agent option MUST use the merge strategy (not
rebase). The flow SHALL be:

1. `git fetch <target-remote> <base-branch>`
2. `git merge <target-remote>/<base-branch>` (creates
   conflict markers)
3. Sub-agent resolves conflict markers in each file
4. Sub-agent stages resolved files with `git add <file>`
5. User reviews cached diff
6. If approved: complete merge commit and push
7. If rejected or failed: `git merge --abort`

#### Scenario: Successful end-to-end resolution

- **GIVEN** PR #42 has conflicts with `origin/main`
- **WHEN** the user selects option 5
- **THEN** `/finale` fetches and merges `origin/main`
- **AND** spawns a sub-agent to resolve the conflicts
- **AND** shows the resolution diff to the user
- **AND** on approval, completes the merge commit
- **AND** pushes and resumes CI check watching

### FR-006: Post-resolution CI check polling

After a successful sub-agent resolution, push, and merge
commit, `/finale` MUST poll for CI checks using the same
bash loop pattern as options 1 and 2:

```bash
while true; do
  STATUS=$(gh pr checks <number> 2>&1)
  if echo "$STATUS" | grep -qE 'pass|fail'; then
    echo "$STATUS"
    break
  fi
  sleep 10
done
```

#### Scenario: CI checks run after resolution

- **GIVEN** the sub-agent resolved conflicts and the merge
  commit was pushed
- **WHEN** CI checks complete
- **THEN** `/finale` reports check results and continues
  to step 7 (pass) or stops with failure details (fail)

## MODIFIED Requirements

### Requirement: Conflict recovery options menu

Previously: Step 6b presented 4 options (merge, rebase,
stop, continue).

Now: Step 6b presents 5 options. Option 5 is "Spawn
sub-agent to resolve conflicts (AI-assisted)." The text
of options 1-4 is unchanged. The numbering of options 1-4
is unchanged.

## REMOVED Requirements

None.
