# Council Review Action

Automated multi-persona code review for GitHub PRs. Discovers
reviewer personas from your repo's `.opencode/agents/` directory,
pre-fetches PR context, and outputs structured JSON for inline
review comments — no custom orchestration code needed.

**[Quick Start](docs/quickstart.md)** — add AI code review to
your repo in three steps.

## Inputs

| Input | Required | Default | Description |
|---|---|---|---|
| `model` | No | `google-vertex-anthropic/claude-sonnet-4-6` | Model in provider/model format |
| `diff-path` | Yes | — | Path to the PR diff file |
| `meta-path` | Yes | — | Path to the PR metadata JSON |
| `github-token` | Yes | — | GitHub token for `gh` CLI |
| `agents-pattern` | No | `.opencode/agents/divisor-*.md` | Glob for Divisor agent files |

## Outputs

| Output | Description |
|---|---|
| `review-json` | Path to the review JSON file (summary + inline comments) |
| `review-mode` | `inline` (structured JSON) or `comment` (plain text fallback) |

The action outputs structured data only. Comment posting is the
consumer workflow's responsibility — see
[docs/architecture.md](docs/architecture.md) for the full
three-workflow chain.

## How it works

1. Installs [OpenCode](https://opencode.ai) CLI
2. Filters noise from the diff (lock files, vendor, generated code)
3. Annotates diff lines with `[L<N>]` prefixes for accurate
   inline comment placement
4. Pre-fetches PR context (CI checks, existing reviews, linked
   issues) via `gh` CLI
5. Discovers Divisor reviewer personas in three tiers:
   repo agents, bundled agents, or single-agent fallback
6. Builds a review prompt referencing the repo's methodology files
7. Invokes `opencode run` with a runtime permission sandbox
8. Parses structured JSON output

## Security

The action processes untrusted PR diff content through an LLM.
Three layers of defense prevent tool misuse:

1. **Runtime permissions** — `OPENCODE_CONFIG_CONTENT` denies
   bash, edit, webfetch, websearch, and skill at the OpenCode
   runtime level (hard enforcement, not prompt instruction)
2. **Plugin isolation** — `--pure` flag prevents external MCP
   plugins from loading
3. **Agent frontmatter** — each Divisor agent's `permission:`
   config denies dangerous tools as defense-in-depth

The diff stays in a file read by OpenCode's Read tool — it is
never interpolated into the prompt string.

See [docs/security-risks.md](docs/security-risks.md) for the
full risk register and
[docs/decisions.md](docs/decisions.md) for design rationale.

## Directory structure

```
council-review-action/
├── action.yml              # Composite action definition
├── README.md
├── scripts/
│   ├── prepare-diff.sh     # Noise filter + line annotation
│   ├── build-prompt.sh     # Prompt construction
│   ├── run-review.sh       # OpenCode invocation + sandbox
│   ├── prefetch.sh         # PR context pre-fetch (CI, reviews)
│   ├── parse-output.sh     # Review output parsing + filtering
│   ├── extract-review-json.py  # JSON extraction from JSONL
│   └── filter-diff-lines.py    # Line number validation
├── test/
│   └── test-pipeline.sh    # Pipeline tests (91 assertions)
└── docs/
    ├── quickstart.md        # Quick start guide
    ├── architecture.md      # End-to-end flow, workflow chain
    ├── decisions.md         # Key technical decisions
    ├── security-risks.md    # Security risk register
    └── testing.md           # Test coverage and strategy
```

## Testing

```bash
bash council-review-action/test/test-pipeline.sh
```

35 scenarios with 91 assertions covering diff processing, JSON
extraction, prompt construction, output parsing, sandbox config
generation, and agent frontmatter validation. See
[docs/testing.md](docs/testing.md) for the coverage matrix and
live integration testing instructions.
