## Why

Speckit enforces feature branches at every pipeline step
via shared bash scripts (`check_feature_branch()` hard
gate). OpenSpec has zero git awareness -- changes are
created, implemented, and archived on whatever branch
the developer happens to be on. This means OpenSpec
changes can accidentally be committed to `main`,
mixed with other work, or lose traceability.

## What Changes

Add branch creation, validation, and lifecycle
management to all OpenSpec command and skill files.

**Convention**: `opsx/<change-name>` branches
(e.g., `opsx/doctor-ux-improvement`).

**Enforcement**: Hard gate -- error and stop if not on
the correct branch. Same rigor as speckit.

## Capabilities

### New Capabilities
- `opsx-branch-creation`: `/opsx-propose` creates and
  checks out an `opsx/<name>` branch after creating the
  change directory
- `opsx-branch-validation`: `/opsx-apply` validates the
  current branch matches the change before implementing
- `opsx-branch-cleanup`: `/opsx-archive` returns to
  main after archiving

### Modified Capabilities
- `cobalt-crush-opsx-detection`: The OpenSpec detection
  path in `/cobalt-crush` validates the branch matches
  the detected change

### Removed Capabilities
- None

## Impact

- 4 command files under `.opencode/command/`
- 3 skill files under `.opencode/skills/`
- AGENTS.md documentation update
- No Go code changes
- No openspec CLI changes (third-party)

## Constitution Alignment

### I. Autonomous Collaboration

**Assessment**: N/A

No artifact communication changes. Branch management
is a developer workflow concern.

### II. Composability First

**Assessment**: PASS

Branch enforcement is instruction-level only (Markdown
command files). No runtime dependencies introduced.
OpenSpec CLI works independently of branch state.

### III. Observable Quality

**Assessment**: N/A

No output format changes.

### IV. Testability

**Assessment**: N/A

No testable code changes -- these are Markdown
instruction files for AI agents.
