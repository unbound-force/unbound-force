## Context

The `/finale` command step 6b (Conflict Recovery) currently
offers four options when a PR has merge conflicts:

1. Merge target branch
2. Rebase onto target branch
3. Stop and resolve manually
4. Continue anyway (CI will not run)

Options 1 and 2 attempt git-level resolution but abort if
file-level conflicts exist, directing the user to resolve
manually. This means any non-trivial conflict requires leaving
the AI-assisted workflow.

Per issue #325 (from PR #320 review), users running `/finale`
are already in an AI-assisted session and would benefit from
an option to spawn a sub-agent for conflict resolution rather
than switching to manual editing.

## Goals / Non-Goals

### Goals
- Add a fifth option to step 6b that spawns a
  `cobalt-crush-dev` sub-agent to attempt conflict resolution
- The sub-agent reads conflicting files, understands both
  sides, resolves conflicts, and stages resolved files
- If the sub-agent succeeds, `/finale` continues with push
  and CI checks
- If the sub-agent fails (partially or fully), fall back to
  manual resolution with a clear report of what was and was
  not resolved
- Show the user a diff of the resolution before committing

### Non-Goals
- Auto-resolving without user confirmation -- the user MUST
  see and approve the resolution before it is committed
- Supporting conflict resolution outside of `/finale` -- this
  is scoped to the step 6b flow only
- Resolving conflicts that span binary files -- only text
  file conflicts are in scope
- Modifying the merge/rebase options (1-2) -- they remain
  as-is for users who prefer git-level resolution

## Decisions

### D1: Sub-agent type is `cobalt-crush-dev`

The `cobalt-crush-dev` agent is the implementation engine
with coding capabilities. It has access to file read/write
tools needed to resolve conflicts. Using an existing agent
type avoids creating a new specialized agent.

### D2: Option placement as option 5

The new option is added after the existing four options
rather than replacing any of them. This preserves backward
compatibility for users with muscle memory for the existing
option numbers.

The option text:

> 5. Spawn sub-agent to resolve conflicts (AI-assisted)

### D3: Two-phase flow -- merge first, then resolve

The sub-agent option works with the merge path (not rebase):

1. `/finale` executes `git merge <target>` which creates
   conflict markers in files
2. The sub-agent reads each conflicting file, resolves the
   markers, and writes the resolved content
3. The sub-agent stages resolved files with `git add`
4. `/finale` shows the resolution diff to the user
5. If approved, `/finale` completes the merge commit and
   pushes

Rationale: Merge is safer than rebase (no force push needed)
and creates conflict markers that the sub-agent can parse.
Rebase conflict resolution would require iterating through
multiple commits, adding complexity for minimal benefit.

### D4: User approval gate before commit

After the sub-agent resolves conflicts, `/finale` MUST show
the user a diff of the changes and ask for confirmation
before completing the merge commit. This aligns with the
existing guardrail: "NEVER commit without user approval."

```bash
git diff --cached
```

The user sees exactly what the sub-agent changed and can
approve, request edits, or abort (falling back to manual).

### D5: Failure handling with partial resolution reporting

If the sub-agent cannot resolve all conflicts:

- Report which files were resolved and which remain
- Abort the merge: `git merge --abort`
- Present the user with the remaining manual options (1-4)

If the sub-agent resolves all conflicts but the user rejects
the resolution:

- Abort the merge: `git merge --abort`
- Return to the conflict recovery options menu

### D6: Sub-agent prompt construction

The Task tool prompt for the sub-agent includes:

- The list of conflicting files (from `git diff --name-only
  --diff-filter=U`)
- The target branch name and PR context
- Instructions to resolve conflict markers (`<<<<<<<`,
  `=======`, `>>>>>>>`) by understanding both sides
- Instructions to stage resolved files
- A directive to report success/failure per file

The sub-agent does NOT receive the full `/finale` context --
it gets a focused, scoped prompt for conflict resolution
only.

## Risks / Trade-offs

### R1: Sub-agent may produce incorrect resolutions

**Risk**: The sub-agent may misunderstand the intent of
conflicting changes and produce semantically incorrect
merges that compile but behave incorrectly.

**Mitigation**: The user approval gate (D4) ensures a human
reviews the resolution diff before it is committed. The CI
check step (step 6) provides automated verification after
push.

### R2: Large conflicts may exceed sub-agent context

**Risk**: Files with extensive conflicts or very large files
may exceed the sub-agent's context window or produce poor
results.

**Mitigation**: The sub-agent reports per-file success/failure.
If it cannot handle a file, that file is reported as
unresolved and the user falls back to manual resolution.

### R3: Merge abort leaves clean state

**Trade-off**: If the sub-agent fails or the user rejects
the resolution, `git merge --abort` returns the working tree
to the pre-merge state. No partial resolutions are left
behind. The downside is that any correct partial resolutions
are also discarded -- the user must redo them manually.

**Accepted**: Clean abort is safer than leaving partial
state. Users who want to preserve partial resolutions can
use option 3 (manual) from the start.
