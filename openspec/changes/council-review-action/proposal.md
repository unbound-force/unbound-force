# Council Review Action

## Why

The council review CI workflow validates PRs using Claude on Vertex AI
via WIF authentication. The current prototype (complytime/org-infra
PR #313) embeds the orchestration logic — agent discovery, prompt
construction, Claude invocation, output parsing — directly in the
reusable workflow YAML. This creates three problems:

1. **Wrong ownership**: The orchestration logic is generic. It reads
   Divisor persona definitions, review methodology, and convention
   packs that are all defined and maintained by unbound-force. Hosting
   the runner in org-infra forces org-infra to maintain code that
   belongs to unbound-force.

2. **Single-agent limitation**: The prototype simulates all five
   Divisor personas in a single Claude invocation. The interactive
   `/review-council` command spawns each persona as a parallel
   subagent with its own context window. The `--agents` CLI flag
   makes real multi-agent orchestration feasible in `claude -p`.

3. **Prompt injection risk**: The prototype interpolates diff content
   into the prompt string and uses a soft preamble as the only
   mitigation. A composite action can isolate the diff in a file
   read by Claude's tools, reducing the injection surface.

Tracked as unbound-force/unbound-force#253.

## What Changes

- **New**: `council-review-action/` directory at repo root containing
  a composite GitHub Action (`action.yml`) and supporting scripts
- **New**: Scripts for pre-fetching PR context, discovering Divisor
  agents, building the review prompt, running Claude, and parsing
  structured output
- **Downstream**: org-infra's `reusable_council_review.yml` becomes a
  thin consumer — WIF auth, fork-safe chain, secret forwarding, and
  PR comment posting — that calls this action for the review itself

## Non-goals

- Full `/review-council` parity in CI (fix loops, local tool
  execution, Gaze analysis, interactive confirmation)
- Multi-provider support (Bedrock, direct Anthropic API)
- Changes to the fork-safe two-workflow chain (stays in org-infra)
- PR comment posting logic (stays in org-infra)
- Container image packaging (CLI install is fast enough)

## Capabilities

### New Capabilities

- `council-review-action`: Composite GitHub Action that discovers
  Divisor personas from `.opencode/agents/divisor-*.md`, builds
  `--agents` JSON dynamically, pre-fetches PR context (CI checks,
  existing reviews, linked issues), invokes `claude -p --agents`
  with read-only tool restrictions, and outputs structured review
  JSON (summary + inline comments)

### Modified Capabilities

(none — this is a new action; org-infra consumption is tracked
separately in complytime/org-infra)

## Impact

- **New directory**: `council-review-action/` with `action.yml` and
  `scripts/` subdirectory
- **Dependencies**: Claude Code CLI (npm, pinned version), Python 3
  (pre-installed on GitHub runners), `jq` (pre-installed), `gh` CLI
  (pre-installed)
- **No Go code changes**: Action is shell scripts + Python, not part
  of the Go module
- **No workflow changes**: Existing CI workflows in unbound-force are
  not modified
