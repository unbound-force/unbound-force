## Why

When `uf init` runs `dewey init`, the default
`sources.yaml` has a single disk source pointing to
`.` (current repo only). The developer must manually
edit `.dewey/sources.yaml` to add sibling repos,
org-level files, and GitHub API sources. This is
tedious and error-prone -- the developer has to know
the source configuration format and discover which
repos exist.

The sibling repos are right there in `../` -- the
scaffold engine can detect them and generate a
comprehensive sources config automatically.

## What Changes

After `dewey init` creates the default `sources.yaml`,
`initSubTools()` in `scaffold.go` detects sibling repos
(directories with `.git/` in `../`), extracts the
GitHub org from the current repo's remote URL, and
generates a multi-repo `sources.yaml` with:

- Per-repo disk sources for each sibling (provenance)
- A `disk-org` source for `../` (org-level files)
- A GitHub API source with all detected repo names
  (issues, PRs, READMEs)

The generated config is user-owned after creation --
`uf init` never overwrites a customized `sources.yaml`.

## Capabilities

### New Capabilities
- `auto-detect-sources`: Detects sibling repos and
  generates multi-repo Dewey sources config
- `github-org-detection`: Extracts GitHub org name
  from `git remote get-url origin`

### Modified Capabilities
- `initSubTools()`: After `dewey init`, generates
  sources config before `dewey index`

### Removed Capabilities
- None

## Impact

- `internal/scaffold/scaffold.go` -- `generateDeweySources()`
- `internal/scaffold/scaffold_test.go` -- tests

## Constitution Alignment

### I. Autonomous Collaboration
**Assessment**: PASS -- generates a config file (artifact).

### II. Composability First
**Assessment**: PASS -- user-owned after creation. Dewey
handles empty sources gracefully (0 documents, no error).

### III. Observable Quality
**Assessment**: PASS -- sources.yaml is human-readable YAML.

### IV. Testability
**Assessment**: PASS -- uses injected `ExecCmd` for git
remote, `os.ReadDir` for sibling detection, `t.TempDir`
for test isolation.
