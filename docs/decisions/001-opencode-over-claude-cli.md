# ADR-001: OpenCode over Claude Code CLI for CI council review

**Status:** Accepted
**Date:** 2026-06-10
**Context:** The council-review-action needs a CLI for non-interactive
AI-assisted code review in GitHub Actions. The team is standardizing
on OpenCode as the agent framework across all repos.

## Decision

Use `opencode run` instead of `claude -p` for the council review
action's Claude invocation layer. OpenCode is the team's agent
framework and auto-discovers `.opencode/` context (agents, commands,
convention packs, severity definitions) that the review depends on.

## Rationale

| Factor | OpenCode | Claude Code CLI |
|--------|----------|-----------------|
| Context discovery | Auto-discovers `.opencode/` | Manual `--agents` JSON |
| Agent definitions | Native (frontmatter) | Requires JSON blob |
| Model routing | `--model vertex/claude-sonnet-4-6` | `--model` + `CLAUDE_CODE_USE_VERTEX` env var |
| Team standard | Yes (`uf init`, all repos) | No (CI-only) |
| Slash commands | `--command review-council` | Not available |
| Install | `curl opencode.ai/install` | `npm install` |

OpenCode eliminates the `discover-agents.py` → `--agents` JSON
construction pipeline. Agent definitions in `.opencode/agents/`
are the single source of truth for both interactive and CI reviews.

## Consequences

**Positive:**
- Single tool for interactive dev and CI review
- Agent definitions used natively (no JSON translation layer)
- Slash commands (`/review-council`, `/review-pr`) reusable in CI
  via `--command`
- Model configuration follows OpenCode hierarchy (project >
  user > default)

**Trade-offs:**
- No `--max-turns` or `--max-budget-usd` flags (Claude-specific)
- No `--allowedTools` flag; tool restrictions come from agent
  frontmatter permissions
- No `--agents` multi-agent blob; OpenCode uses `--agent` (single)
  or relies on the orchestrator prompt to delegate via Agent tool
- `--dangerously-skip-permissions` required in CI (no TTY for
  approval prompts)

**Risks:**
- Vertex AI auth: Claude Code CLI uses `CLAUDE_CODE_USE_VERTEX=1`.
  OpenCode uses `--model vertex/claude-sonnet-4-6` to route via
  Vertex AI. WIF sets up GCP ADC; `ANTHROPIC_VERTEX_PROJECT_ID`
  and `CLOUD_ML_REGION` env vars stay in the consumer workflow.
  Needs validation during end-to-end testing.
- OpenCode `run` in headless CI is less battle-tested than
  `claude -p` for this use case.

## Deferred Work

| Item | Blocker | Action needed |
|------|---------|---------------|
| Vertex AI provider config for OpenCode | Needs testing | Validate WIF auth + `--model vertex/...` works with `opencode run` |
| Cost/turn guardrails | OpenCode lacks flags | Investigate OpenCode config options or accept agent-level controls |
| `--command review-council` in CI | CI constraints (no Shell) | Test if command respects CI constraints from prompt |

## Alternatives Considered

| Option | Verdict | Why |
|--------|---------|-----|
| Claude Code CLI (`claude -p`) | Rejected | Not the team standard; requires JSON agent construction; adds a separate tool to maintain |
| `anthropics/claude-code-action` | Deferred | First-party but lacks fork-safety, requires GitHub App token, doesn't use OpenCode context |
| OpenCode GitHub agent (`opencode github run`) | Investigate | May be the intended CI path but not documented enough yet |
