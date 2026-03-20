# Quickstart: Cobalt-Crush Architecture

**Spec**: 006-cobalt-crush-architecture
**Date**: 2026-03-20

## Prerequisites

- `unbound` CLI installed: `brew install unbound-force/tap/unbound`
- OpenCode running in the target project

## Deploy Cobalt-Crush

```bash
unbound init
```

This deploys everything including `cobalt-crush-dev.md`
in `.opencode/agents/`, the Divisor review agents, and
the shared convention packs at `.opencode/unbound/packs/`.

## What Gets Deployed

```text
.opencode/
├── agents/
│   ├── cobalt-crush-dev.md      # Developer persona
│   ├── divisor-guard.md         # Review: intent drift
│   ├── divisor-architect.md     # Review: structure
│   ├── divisor-adversary.md     # Review: resilience
│   ├── divisor-sre.md           # Review: operations
│   └── divisor-testing.md       # Review: test quality
├── command/
│   └── review-council.md        # Review orchestration
└── unbound/
    └── packs/
        ├── go.md                # Go conventions (shared)
        ├── go-custom.md         # Project-specific Go rules
        ├── default.md           # Universal conventions
        ├── default-custom.md    # Project-specific rules
        ├── typescript.md        # TypeScript conventions
        └── typescript-custom.md # Project-specific TS rules
```

## How Cobalt-Crush Works

Cobalt-Crush is the coding persona — it defines *how* to
write code. When you or `/speckit.implement` write code,
the Cobalt-Crush agent provides:

1. **Engineering Philosophy**: Clean code, SOLID, TDD
   awareness, spec-driven development
2. **Convention Pack Adherence**: Reads the same convention
   packs as The Divisor, ensuring developer-reviewer
   alignment
3. **Test Hook Generation**: Produces code with interface
   abstractions, dependency injection, and exported test
   helpers for Gaze validation
4. **Gaze Feedback Loop**: Reads Gaze quality reports from
   `.unbound-force/artifacts/` and addresses issues
5. **Divisor Review Preparation**: Prepares code to pass
   The Divisor's multi-persona review
6. **Knowledge Graph** (optional): Uses graphthulhu MCP
   to search specs and past review patterns

## Feedback Loop Workflow

```text
1. Write code (Cobalt-Crush persona active)
      │
      ▼
2. Run Gaze validation
      │
      ▼
3. Read Gaze report, fix issues
      │
      ▼
4. Submit for Divisor review (/review-council)
      │
      ▼
5. Read review findings, fix issues
      │
      ▼
6. Re-run Gaze validation after fixes
      │
      ▼
7. Re-submit for review
      │
      ▼
8. APPROVED → merge
```

## Customize Cobalt-Crush

Edit `.opencode/agents/cobalt-crush-dev.md` to customize
the persona for your project. The file is user-owned —
`unbound init` will not overwrite it on subsequent runs.

### Add Project-Specific Coding Rules

Edit `.opencode/unbound/packs/{lang}-custom.md` to add
rules that both Cobalt-Crush and The Divisor will follow:

```markdown
## Custom Rules

- **CR-001** [MUST] All database queries MUST use the
  repository pattern defined in `internal/repo/`.
- **CR-002** [SHOULD] HTTP handlers SHOULD validate
  request bodies using the `validate` package.
```

## Without Gaze or The Divisor

Cobalt-Crush works standalone. If Gaze is not installed,
the agent notes that quality validation is unavailable
and recommends installing it. If The Divisor is not
installed, the agent skips review preparation steps.
Convention packs still guide coding standards regardless
of which heroes are deployed.
<!-- scaffolded by unbound vdev -->
