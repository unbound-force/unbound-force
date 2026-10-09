## ADDED Requirements

### Requirement: resolve_base_ref tool

The `uf-workflow` plugin MUST register a `resolve_base_ref`
tool that determines the best available base ref for local
diff operations.

The tool MUST try refs in this order:
1. `upstream/main` via `git rev-parse --verify upstream/main`
2. `origin/main` via `git rev-parse --verify origin/main`
3. `main` via `git rev-parse --verify main`

The tool MUST return the first ref that resolves
successfully, along with its resolved SHA and a source
label indicating which ref was selected.

The tool MUST NOT perform any network access (no `git
fetch`). It SHALL use only cached remote-tracking refs.

If none of the three refs resolve, the tool MUST return a
failure result with `retryable: false`.

#### Scenario: Fork checkout with upstream remote

- **GIVEN** the local repository has an `upstream` remote
  and `upstream/main` resolves to a valid commit
- **WHEN** `resolve_base_ref` is called
- **THEN** it returns `upstream/main` as the ref, its
  resolved SHA, and source `upstream/main`

#### Scenario: Direct clone without upstream remote

- **GIVEN** the local repository has no `upstream` remote
  but `origin/main` resolves to a valid commit
- **WHEN** `resolve_base_ref` is called
- **THEN** it returns `origin/main` as the ref, its
  resolved SHA, and source `origin/main`

#### Scenario: Local-only repository

- **GIVEN** neither `upstream/main` nor `origin/main`
  resolve but local `main` exists
- **WHEN** `resolve_base_ref` is called
- **THEN** it returns `main` as the ref, its resolved SHA,
  and source `main`

#### Scenario: No resolvable ref

- **GIVEN** none of `upstream/main`, `origin/main`, or
  `main` resolve to valid commits
- **WHEN** `resolve_base_ref` is called
- **THEN** it returns a failure with `retryable: false`
  and a descriptive error message

## MODIFIED Requirements

### Requirement: uf.review-council local diff base

The local (no-PR) review path in `uf.review-council` MUST
call `resolve_base_ref` to determine the diff base ref
instead of hardcoding `main`.

Previously: The local review path used `main` as the base
ref unconditionally.

The command MUST announce the selected base ref and its
resolved SHA before computing the diff.

If `resolve_base_ref` fails, the command MUST abort with
an error message. It MUST NOT fall back to `main` or
proceed without a valid base.

#### Scenario: Stale local main in fork checkout

- **GIVEN** local `main` is 15 commits behind
  `upstream/main` and the user is on a feature branch
- **WHEN** `/uf.review-council` runs in local mode
- **THEN** it uses `upstream/main` as the diff base and
  the diff excludes the 15 already-merged commits

### Requirement: uf.unleash Step 10 demo output

The Step 10 demo output MUST NOT include a
`## Key Files Changed` section.

Previously: Step 10 ran `git diff --name-only main...HEAD`
and included the result in a `## Key Files Changed` section.

## REMOVED Requirements

### Requirement: Key Files Changed display

The `## Key Files Changed` section and the `git diff
--name-only main...HEAD` instruction in `uf.unleash`
Step 10 are removed to reduce token usage. The user
already has full diff context from the review step.
