## Why

A developer who wants the full Unbound Force experience
must currently run `uf setup`, then manually run
`dewey init`, `dewey index`, install the OpenSpec CLI,
install Mx F, and install the GitHub CLI. There are 2
tools that `uf doctor` checks for but `uf setup` does
not install (`mxf`, `gh`), 1 tool that neither checks
nor installs (`openspec` CLI), and 2 initialization
steps missing (`dewey init`, `dewey index`).

Additionally, `uf init` scaffolds 49 files but does not
initialize Dewey's workspace -- so a developer who runs
`uf init` in a new project still needs to know to run
`dewey init` and `dewey index` separately. The post-init
guidance (`printSummary`) only mentions speckit and
openspec commands, missing crucial next steps like
creating a constitution and running `uf doctor`.

The goal is: **one command (`uf setup`) installs
everything, and `uf init` initializes everything** --
a developer wanting all features should not need to
know about individual tool init commands.

## What Changes

### `uf setup` additions (internal/setup/setup.go)

Install 3 new tools following existing patterns:
- **Mx F** (`mxf`): `brew install unbound-force/tap/mxf`
  -- follows `installGaze()` pattern
- **GitHub CLI** (`gh`): `brew install gh` -- follows
  `installGaze()` pattern
- **OpenSpec CLI** (`openspec`):
  `npm install -g @fission-ai/openspec@latest` --
  follows `installSwarmPlugin()` pattern (requires
  Node.js)

Add 2 new initialization steps:
- **`dewey init`**: Create `.dewey/` workspace after
  Dewey binary + model are installed
- **`dewey index`**: Build initial index from local
  files after workspace is created

### `uf init` additions (internal/scaffold/scaffold.go)

After scaffolding files, initialize sub-tools:
- Add `LookPath` and `ExecCmd` fields to
  `scaffold.Options` for testability
- If Dewey is available and `.dewey/` doesn't exist,
  run `dewey init` + `dewey index`
- Update `printSummary` with next-step guidance:
  constitution creation, `uf doctor`, and workflow
  commands

## Capabilities

### New Capabilities
- `setup-install-mxf`: Installs Mx F Manager hero
- `setup-install-gh`: Installs GitHub CLI
- `setup-install-openspec`: Installs OpenSpec CLI
- `setup-dewey-init`: Creates Dewey workspace
- `setup-dewey-index`: Builds initial Dewey index
- `init-sub-tools`: `uf init` initializes Dewey when
  available
- `init-next-steps`: `printSummary` shows actionable
  next steps based on tool availability

### Modified Capabilities
- `setup-step-order`: Revised from 11 to 16 steps
- `scaffold-options`: Adds LookPath/ExecCmd fields

### Removed Capabilities
- None

## Impact

- `internal/setup/setup.go` -- 3 new install functions,
  2 new init steps, updated step ordering
- `internal/setup/setup_test.go` -- new tests for each
  new step
- `internal/scaffold/scaffold.go` -- Options struct
  expanded, initSubTools(), updated printSummary
- `internal/scaffold/scaffold_test.go` -- tests for
  initSubTools and printSummary

## Constitution Alignment

### I. Autonomous Collaboration
**Assessment**: N/A -- tool installation, not
artifact communication.

### II. Composability First
**Assessment**: PASS -- all new installations produce
warnings on failure, not hard failures. Missing tools
degrade gracefully. No tool becomes a hard dependency.

### III. Observable Quality
**Assessment**: PASS -- each step produces a
machine-parseable `stepResult` with name, action, and
detail. The `printSummary` shows sub-tool init status.

### IV. Testability
**Assessment**: PASS -- all new functions use injected
`LookPath`/`ExecCmd`. Scaffold Options expanded to
support injection. No `exec.LookPath` or
`exec.Command` direct calls.
