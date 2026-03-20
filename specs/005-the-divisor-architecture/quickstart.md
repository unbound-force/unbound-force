# Quickstart: The Divisor Architecture

**Spec**: 005-the-divisor-architecture
**Date**: 2026-03-19

## Prerequisites

- `unbound` CLI installed: `brew install unbound-force/tap/unbound`
- OpenCode running in the target project
- A project with source code to review

## Deploy The Divisor

### Option 1: Full scaffold (recommended for new projects)

```bash
unbound init
```

This deploys everything: Speckit templates, OpenSpec
schema, Divisor agents, convention packs, and all
supporting files. Language is auto-detected from
project markers (`go.mod`, `package.json`, etc.).

### Option 2: Divisor only

```bash
unbound init --divisor
```

Deploys only The Divisor's review agents, the
`/review-council` command, and the convention pack
for the detected language.

### Option 3: Specific language

```bash
unbound init --divisor --lang typescript
```

Deploys Divisor with the TypeScript convention pack
regardless of auto-detection.

## What Gets Deployed

```text
.opencode/
├── agents/
│   ├── divisor-guard.md         # Intent drift detection
│   ├── divisor-architect.md     # Structural review
│   ├── divisor-adversary.md     # Resilience/security
│   ├── divisor-sre.md           # Operational readiness
│   └── divisor-testing.md       # Test quality
├── command/
│   └── review-council.md        # Orchestration command
└── divisor/
    └── packs/
        ├── go.md                # Go convention pack
        ├── go-custom.md         # Project-specific extensions
        ├── default.md           # Language-agnostic fallback
        └── default-custom.md    # Fallback extensions
```

## Run a Code Review

```
/review-council
```

This will:
1. Discover all `divisor-*.md` agents in `.opencode/agents/`
2. Replicate CI checks locally
3. Delegate to all discovered personas in parallel
4. Collect verdicts and compute the council decision
5. If any persona requests changes, iterate (up to 3 times)
6. Produce a structured review report

## Run a Spec Review

```
/review-council specs
```

Reviews specification artifacts instead of code changes.

## Customize Convention Packs

### Add project-specific rules

Edit `.opencode/unbound/packs/{lang}-custom.md` and add
rules using the `CR-NNN` prefix:

```markdown
## Custom Rules

- **CR-001** [MUST] All HTTP handlers MUST validate
  the request body before processing.
- **CR-002** [SHOULD] Database queries SHOULD use the
  repository pattern defined in `internal/repo/`.
```

### Update canonical pack rules

Canonical packs (`go.md`, `typescript.md`, `default.md`)
are tool-owned and auto-updated by `unbound init`. Your
customizations in `{lang}-custom.md` are never
overwritten.

To force-update all files including custom packs:

```bash
unbound init --divisor --force
```

## Add a Custom Persona

Create a new `divisor-{function}.md` file in
`.opencode/agents/`:

```markdown
---
description: Accessibility reviewer
mode: subagent
model: google-vertex-anthropic/claude-sonnet-4-6@default
temperature: 0.1
tools:
  write: false
  edit: false
  bash: false
---

# Role: The Accessibility Reviewer

You review code changes for accessibility compliance...
```

The `/review-council` command will automatically discover
and invoke your custom persona on the next run.

## Remove a Persona

Delete the `divisor-*.md` file from `.opencode/agents/`.
The `/review-council` command will note it as absent in
the discovery summary but will not block the review.

## Migration from `reviewer-*` Files

If your project has existing `reviewer-*.md` files from
a previous `unbound init`:

1. Run `unbound init --divisor` to deploy the new
   `divisor-*` files alongside the old ones.
2. The `/review-council` command now scans for
   `divisor-*.md` only. Old `reviewer-*` files are
   inert.
3. Verify the new agents work: `/review-council`
4. Manually remove old files:
   ```bash
   rm .opencode/agents/reviewer-*.md
   ```
<!-- scaffolded by unbound vdev -->
