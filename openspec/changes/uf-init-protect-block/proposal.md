## Why

The `/uf.init` slash command (`.opencode/commands/uf.init.md`) is a 1000+ line
instruction file that guides an LLM through applying project-specific
customizations to third-party tool files. During long execution runs, the
LLM's context window may undergo compaction, which truncates or summarizes
earlier instructions. Because `/uf.init` contains critical step-by-step
logic (prerequisite checks, idempotent insertions, post-write verification),
losing any portion of its instructions to compaction can cause the command
to skip steps, produce incorrect insertions, or fail to detect errors.

GitHub issue #540 identifies this vulnerability: the command has no
mechanism to protect its core instructions from being compacted away.
Additionally, files that `/uf.init` writes into target repositories
(scaffolded commands, config files) may themselves be vulnerable to
compaction in subsequent LLM sessions, creating a cascading risk.

## What Changes

- Wrap the core instruction body of `.opencode/commands/uf.init.md` in a
  single `<protect>` block so that OpenCode's context compaction engine
  preserves the full instruction set during execution.
- Ensure that files `/uf.init` writes into a target repository (scaffolded
  commands, config files, guardrail blocks) either include their own
  `<protect>` blocks or are inserted into existing `<protect>` blocks, so
  that downstream files are similarly resilient to compaction.

## Capabilities

### New Capabilities
- `uf-init-compaction-protection`: The `/uf.init` command's core
  instructions are wrapped in a `<protect>` block, ensuring they survive
  context compaction during execution. Files written by `/uf.init` into
  target repositories also carry protection blocks or are placed within
  existing protection blocks.

### Modified Capabilities
- `/uf.init`: Gains a `<protect>` block around its core instructions
  (Steps 0-12). The command's behavior is
  unchanged; only the wrapping metadata is added.

### Removed Capabilities
- None.

## Impact

- **`.opencode/commands/uf.init.md`**: The command file gains a `<protect>`
  block wrapper around its instruction body. No functional changes.
- **Scaffold files written by `/uf.init`**: Files that `/uf.init` creates
  or modifies in target repositories (e.g., OpenSpec skill files, speckit
  command files, guardrail blocks) will include or be placed within
  `<protect>` blocks. This affects the insertion templates defined in
  Steps 2-8 of the command.
- **Existing target repositories**: Re-running `/uf.init` after this change
  will add `<protect>` blocks to previously scaffolded files. The
  idempotency checks in each step ensure no duplicate content is inserted.
- **No Go binary changes**: This is purely a slash-command and template
  change. The `uf` CLI binary is unaffected.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

This change affects only the instruction metadata of a slash command file.
It does not alter hero-to-hero artifact interfaces, communication formats,
or synchronization patterns. The `/uf.init` command continues to produce
the same self-describing output files with scaffold comments and version
markers.

### II. Composability First

**Assessment**: PASS

The `<protect>` block is a metadata wrapper that does not introduce any
new runtime dependencies. Target repositories that do not use OpenCode
will simply ignore the `<protect>` tags. The change maintains standalone
functionality of every affected file.

### III. Observable Quality

**Assessment**: PASS

The `<protect>` block is a declarative metadata tag that is
machine-parseable. It does not alter the observable output or provenance
metadata of any artifact. Scaffold comments (`<!-- scaffolded by uf ... -->`)
remain intact.

### IV. Testability

**Assessment**: PASS

The change is testable in isolation: verification consists of checking
that `<protect>` tags are present in the expected locations in
`.opencode/commands/uf.init.md` and in the insertion templates defined
by the command. No external services are required.
