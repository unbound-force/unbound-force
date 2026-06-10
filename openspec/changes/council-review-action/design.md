# Council Review Action — Design

## Context

The council review CI pipeline authenticates to Vertex AI via ITPC
Workload Identity Federation (established in complytime/org-infra
PR #313) and invokes the review via OpenCode. The orchestration
logic — auto-discovering Divisor personas, constructing prompts
from `.opencode/` methodology files, invoking `opencode run`,
parsing structured output — is generic and belongs in unbound-force
as a reusable composite GitHub Action.

See ADR-001 (`docs/decisions/001-opencode-over-claude-cli.md`) for
the decision to use OpenCode over Claude Code CLI.

The consuming org (complytime/org-infra) provides:
- Fork-safe two-workflow chain (collect + consumer + reusable)
- WIF authentication to Vertex AI
- Secret forwarding and concurrency control
- PR comment posting with bot marker deduplication

The action provides:
- Native agent discovery via OpenCode's `.opencode/` auto-discovery
- Pre-fetch of PR context (CI checks, reviews, linked issues)
- Prompt construction referencing review methodology
- Review invocation via `opencode run` with `--format json`
- Structured JSON output (summary + inline comments)

**Predecessor**: org-infra `council-review-orchestration` design.md
(D1-D8). This design supersedes D8 (prototype-then-extract) by
building directly in unbound-force. See ADR-001 for the OpenCode
decision.

## Goals / Non-Goals

### Goals

- Composite GitHub Action consumable by any org with Claude access
- Multi-agent review via `--agents` with dynamic persona discovery
- Review criteria sourced from the repo's `.opencode/` files
- Automatic adaptation when `uf init` updates agent definitions
- Pre-fetched PR context for methodology compliance without Shell
- Structured JSON output compatible with inline review posting
- Single-agent fallback for repos without `uf init` scaffolding
- Improved prompt injection posture over the prototype

### Non-Goals

- WIF or any authentication (consumer's responsibility)
- Fork-safe workflow chain (consumer's architecture)
- PR comment posting (consumer handles output formatting)
- Container image packaging (CLI install latency is acceptable)
- Full `/review-council` parity (fix loops, Gaze, interactive mode)

## Decisions

### D1: Composite action over reusable workflow

**Decision**: Package as a composite GitHub Action (`action.yml` with
`runs: using: composite`) rather than a reusable workflow.

**Alternative**: Reusable workflow in
`unbound-force/.github/workflows/`.

**Why rejected**: Reusable workflows require `workflow_call` trigger
and cannot be composed within a single job. A composite action can be
called as a step within the consumer's existing reusable workflow,
keeping the auth → review → post flow in one job without cross-job
artifact passing. Composite actions are also more portable — any
workflow can `uses:` them regardless of trigger type.

### D2: OpenCode native agent discovery over JSON construction

**Decision**: Use `opencode run` which auto-discovers agents from
`.opencode/agents/divisor-*.md` via its native context loading.
The prompt instructs the orchestrator to read and apply each
persona's review criteria.

**Alternative**: Build `--agents` JSON and use `claude -p --agents`
(the Claude Code CLI approach).

**Why rejected**: See ADR-001. OpenCode is the team standard, auto-
discovers agent definitions natively, and eliminates the Python
JSON construction pipeline. Agent frontmatter (permissions, mode)
is used directly without translation.

### D3: Diff in file, not in prompt string

**Decision**: The diff stays in `pr-diff-truncated.patch` on disk.
The prompt instructs Claude to `Read` the file. The diff is never
interpolated into the prompt string or shell argument.

**Alternative**: Pass diff content in the prompt via heredoc or
shell variable expansion.

**Why rejected**: Interpolating untrusted diff content into a prompt
string is the primary prompt injection vector. File-based access
creates a structural boundary — Claude's tool call to `Read` the
diff is distinct from the system/user prompt instructions.

### D4: Pre-fetch PR context in action steps

**Decision**: The action runs `gh` commands to pre-fetch CI check
results, existing reviews, inline comments, and linked issues into
JSON files. Claude reads these files via `Read` tool.

**Alternative**: Give Claude Shell access to run `gh` directly.

**Why rejected**: The diff contains untrusted content. Shell access
would allow prompt injection to escalate to command execution.
Pre-fetching in trusted bash steps eliminates this risk.

### D5: Read-only tool access

**Decision**: Parent agent gets `Read,Glob,Agent`. Subagents get
`Read,Glob`. No agent gets `Shell`, `Write`, or `Edit`.

**Alternative**: Scoped Shell access (e.g., `Bash(jq:*,wc:*)`).

**Why rejected**: Even scoped Shell access creates injection surface.
Claude can parse diff content and JSON files using text understanding
without Shell tools.

### D6: Single-agent fallback

**Decision**: When zero `divisor-*.md` files are found (repo has not
run `uf init`), fall back to single-agent mode: skip `--agents`,
use `--allowedTools "Read,Glob"`, and include a generic review
prompt. Log a `::notice::` about the fallback.

**Alternative**: Skip review entirely.

**Why rejected**: Repos consuming the synced workflows may not have
scaffolded unbound-force yet. A degraded review is better than no
review.

### D7: Action does not post comments

**Decision**: The action outputs the review JSON file path and
review mode. The consumer workflow handles posting (inline review
via GitHub API, fallback to PR comment, bot marker deduplication).

**Alternative**: Action posts comments directly using `gh`.

**Why rejected**: Comment posting involves org-specific concerns:
bot markers, deduplication strategy, review event type (COMMENT vs
APPROVE), formatting. These vary by consumer. The action stays
generic by outputting structured data.

## Risks / Trade-offs

**[R1] `--agents` flag stability** — The `--agents` CLI flag is
part of Claude Code SDK. The CLI is pinned to a specific version
(`@2.1.168`). Version upgrades are explicit and tested.

**[R2] Higher token cost** — Multi-agent uses more tokens than
single-agent. Each subagent has its own context window. Mitigation:
`max-budget-usd` input caps total spend per review.

**[R3] Python dependency** — Agent discovery uses Python for safe
JSON construction. GitHub-hosted runners include Python 3 by
default. Self-hosted runners may not. Mitigation: the script is
minimal (~15 lines) and could be rewritten in `jq` if needed.

**[R4] Action versioning** — Consumers pin to a SHA or tag. During
development, consumers use `@branch-name`. For production, tag
releases and pin to SHA.

**[R5] Cross-repo `uses:` reference** — Consumers reference
`unbound-force/unbound-force/council-review-action@SHA`. This
requires the unbound-force repo to be public (it is).
