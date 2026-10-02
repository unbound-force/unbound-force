## Context

The `/uf.init` command (`.opencode/commands/uf.init.md`) is a 1008-line
slash command that applies project-specific customizations to third-party
tool files. It contains 12 numbered steps plus post-write verification,
covering branch enforcement, Dewey context injection, 3-tier degradation
patterns, speckit custom commands, guardrail injection, STOP HERE blocks,
scaffold comment deduplication, and legacy directory cleanup.

Other UF slash commands (`uf.review-council`, `uf.unleash`, `uf.finale`,
`uf.address-feedback`, `uf.review-pr`, `uf.triage-issue`) already use
`<protect>` blocks to shield their core instructions from context
compaction. The `/uf.init` command is a notable omission -- it is the
longest UF command and one of the most step-heavy, making it the most
vulnerable to compaction-induced instruction loss.

Additionally, several files that `/uf.init` writes into target repositories
(speckit command guardrails, STOP HERE blocks, branch enforcement blocks)
are themselves long instruction files that would benefit from compaction
protection in subsequent LLM sessions.

## Goals / Non-Goals

### Goals
- Wrap the core instruction body of `/uf.init` in a `<protect>` block,
  following the same pattern used by other UF commands.
- Ensure that insertion templates defined in `/uf.init` Steps 2-10 produce
  content that is either wrapped in its own `<protect>` block or inserted
  into an existing `<protect>` block in the target file.
- Maintain full idempotency: re-running `/uf.init` after this change MUST
  NOT produce duplicate `<protect>` blocks or duplicate content.
- Preserve backward compatibility: target repositories that do not use
  OpenCode MUST continue to function correctly (the `<protect>` tag is
  an HTML comment-equivalent that non-OpenCode tools ignore).

### Non-Goals
- Modifying the `uf` Go binary. This change is purely to slash-command
  markdown files and insertion templates.
- Adding `<protect>` blocks to files that `/uf.init` does NOT write to.
  Other commands are responsible for their own protection.
- Changing the functional behavior of `/uf.init`. The command's logic,
  step ordering, and idempotency markers remain unchanged.
- Protecting the OpenSpec skill files (`.opencode/skills/openspec-*/SKILL.md`)
  that `/uf.init` modifies. Those files are owned by the OpenSpec CLI and
  will be reset by `openspec init`; protection is the responsibility of
  the OpenSpec project.

## Decisions

### D1: Single `<protect>` block wrapping the entire instruction body

The `<protect>` block will be placed immediately after the `## Instructions`
heading (line 24 of the current file) and will extend to the end of the
last step (Step 12, before the `### Next Steps` section). This wraps all
12 steps plus the Post-Write Verification section in a single block.

**Rationale**: A single block is simpler to maintain than per-step blocks,
and matches the pattern used by `uf.review-council` and `uf.unleash`.
The `## Description` and `## Instructions` headings remain outside the
block so that the command's frontmatter and overview are still visible
after compaction.

### D2: Protection for files written by `/uf.init`

Files that `/uf.init` creates or modifies fall into two categories:

1. **Files created by `/uf.init`** (Step 5: speckit custom commands like
   `speckit.analyze.md`, `speckit.checklist.md`, `speckit.clarify.md`,
   `speckit.taskstoissues.md`): These files MUST include a `<protect>`
   block wrapping their core instruction body at creation time.

2. **Files modified by `/uf.init`** (Steps 2-4, 6, 8, 10: branch
   enforcement, Dewey context, guardrails, STOP HERE blocks inserted
   into existing OpenSpec/speckit files): The inserted content MUST be
   placed inside an existing `<protect>` block if one is present in the
   target file. If no `<protect>` block exists, the insertion MUST NOT
   create one -- the target file's owner is responsible for its own
   protection. This avoids `/uf.init` making structural changes to files
   it does not own.

**Rationale**: Category 1 files are fully owned by `/uf.init` (it creates
them from scratch), so they can include protection from the start.
Category 2 files are owned by upstream projects (OpenSpec, speckit);
`/uf.init` is a guest modifier and SHOULD NOT impose structural changes
like adding `<protect>` wrappers to files it does not own.

### D3: Idempotency marker for `<protect>` block in `/uf.init`

The idempotency check for the `<protect>` block in `/uf.init` itself
will look for the presence of `<protect>` between the `## Instructions`
heading and the first step heading (`### Step 0`). If found, the block
is considered present and the insertion is skipped.

For files created by Step 5, the existing file-existence check (Step 5.1)
already provides idempotency: if the file exists, it is skipped entirely,
so the `<protect>` block within it is never duplicated.

### D4: No changes to OpenSpec skill files

The OpenSpec skill files (`.opencode/skills/openspec-*/SKILL.md`) are
modified by `/uf.init` Steps 2-4 but are owned by the OpenSpec CLI.
They will be reset by `openspec init` or `npm update`. Adding `<protect>`
blocks to them would be futile since the blocks would be overwritten.
This change does NOT add protection to OpenSpec skill files.

## Risks / Trade-offs

- **Risk**: If the `<protect>` tag format changes in a future OpenCode
  version, all protected files would need to be updated. **Mitigation**:
  The `<protect>` tag is a simple HTML-like tag that has been stable
  across OpenCode versions. The risk is low.
- **Risk**: Target files with `<protect>` blocks may be rejected by
  strict markdown linters that do not recognize the tag. **Mitigation**:
  The tag is placed inside markdown files where HTML passthrough is
  standard. Existing UF commands already use this pattern without
  linter issues.
- **Trade-off**: Wrapping the entire instruction body in a single
  `<protect>` block means the entire 1000+ lines are protected from
  compaction, which consumes more of the context window. **Acceptance**:
  This is the desired behavior -- the entire instruction set is critical
  and must not be partially compacted.
