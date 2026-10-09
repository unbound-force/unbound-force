## Context

OpenSpec changes are created via `openspec new change`
which scaffolds a directory containing `.openspec.yaml`
and placeholder artifact files. The `.openspec.yaml`
currently holds only `schema` and `created` fields.

Speckit specs have a frontmatter block in their
`proposal.md` (or equivalent) that similarly lacks an
originating-issue field.

`/uf.finale` Step 5d generates the PR body from commit
history, diff, and spec artifacts. It has no mechanism
to discover an originating issue number.

GitHub's `Closes #N` / `Fixes #N` trailer in a PR body
automatically closes the referenced issue when the PR
is merged. Without this trailer, issue closure is a
manual step that is frequently forgotten.

## Goals / Non-Goals

### Goals
- Add an optional `originating_issue` field to OpenSpec
  `.openspec.yaml` and Speckit spec frontmatter.
- Provide an `--issue <N>` flag on `openspec new change`
  to populate the field at creation time.
- Update `/uf.finale` Step 5d to read the field from
  the change directory and emit `Closes #<N>` in the
  PR body when present.
- Maintain full backward compatibility: changes without
  the field behave identically to today.

### Non-Goals
- Automatically discovering the issue number from
  branch names, commit messages, or `gh` API queries.
  The developer supplies it explicitly at creation time.
- Supporting multiple originating issues per change.
  A single integer field covers the common case; a
  list can be added later without breaking the scalar.
- Modifying the Speckit `specify` or `clarify` phases.
  Only the spec frontmatter is touched.
- Changing the PR template or adding a `Closes` field
  to it. The trailer is emitted programmatically by
  `/uf.finale`.

## Decisions

### D1: Field name and location

The field is named `originating_issue` and lives in
`.openspec.yaml` for OpenSpec changes. For Speckit,
it lives in the spec's `proposal.md` frontmatter as
`originating_issue: <integer>`.

**Rationale**: A single, well-known key in the
change's root metadata is the simplest location that
`/uf.finale` can read without parsing arbitrary
artifact files. YAML frontmatter is already the
convention for Speckit metadata.

### D2: Optional, not required

The field is optional. Its absence means "no
originating issue" and `/uf.finale` emits no `Closes`
trailer.

**Rationale**: Enforcing the field would break all
existing changes and create friction for ad-hoc
changes that do not originate from an issue.
Constitution Principle II (Composability First)
demands that the change remain usable without the
new field.

### D3: Scalar integer, not a list

The field holds a single integer (the issue number).
If a change addresses multiple issues, the developer
chooses the primary one. Additional references can
be added manually to the PR body.

**Rationale**: The common case is a 1:1 mapping.
A list adds complexity to the schema, the CLI flag,
and the PR body generation for marginal benefit.
The scalar can be extended to a list later as a
non-breaking change.

### D4: `Closes` keyword, not `Fixes`

`/uf.finale` emits `Closes #<N>`, not `Fixes #<N>`.

**Rationale**: Both keywords are functionally
equivalent on GitHub. `Closes` is the more neutral
term and is used elsewhere in the codebase (e.g.
commit messages in `/uf.finale` examples).

### D5: Placement in PR body

The `Closes #<N>` line is appended immediately after
the `## Summary` section, before `## How to Test`.

**Rationale**: GitHub's keyword detection scans the
entire PR body, so placement is flexible. Placing it
near the top ensures visibility to reviewers and
keeps it separate from the attribution footer.

### D6: Reading the field in /uf.finale

`/uf.finale` detects the change directory by
inspecting the branch name:
- `opsx/<name>` → read `openspec/changes/<name>/.openspec.yaml`
- `NNN-*` → read the Speckit spec's `proposal.md` frontmatter

The YAML is parsed with the existing YAML library
(`gopkg.in/yaml.v3` or `github.com/goccy/go-yaml`).

**Rationale**: Branch-name detection is already used
by `/uf.finale` to locate spec artifacts for the
"How to Test" section. Reusing the same mechanism
avoids introducing a new discovery path.

## Risks / Trade-offs

- **Risk**: Developer forgets to pass `--issue <N>`
  at creation time. **Mitigation**: `/uf.finale` can
  warn when on an `opsx/*` branch with no
  `originating_issue` set, suggesting the developer
  add it manually to `.openspec.yaml`. This is a
  soft warning, not a hard gate.

- **Risk**: The field becomes stale if the issue is
  closed or reassigned before the PR merges.
  **Mitigation**: GitHub ignores `Closes` references
  to already-closed issues. No action needed.

- **Trade-off**: Scalar vs. list. Accepting a single
  issue number is simpler but forces the developer
  to choose a primary issue when multiple apply.
  Accepted for now; can be extended later.

- **Risk**: Speckit changes use a different metadata
  location (frontmatter in `proposal.md`) than
  OpenSpec (`.openspec.yaml`). `/uf.finale` must
  handle both. **Mitigation**: The detection logic
  is branch-prefix-based and straightforward.
