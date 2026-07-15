# Council Review Action

AI code review composite GitHub Action using [OpenCode](https://opencode.ai) with Divisor persona discovery. Runs structured reviews against PR diffs via Claude on Vertex AI and outputs JSON for inline GitHub comments.

## Architecture

### End-to-end flow

```
┌─────────────────────────────────────────────────────────┐
│  Downstream Repo (e.g., org-infra, gaze)                │
│                                                         │
│  ci_council_review_collect.yml  (pull_request trigger)  │
│  ├── Gate: skip bots, skip drafts                       │
│  ├── Capture diff: gh pr diff → pr-diff.patch           │
│  ├── Build metadata: pr-meta.json                       │
│  └── Upload artifact: council-review-diff               │
│                                                         │
│  ci_council_review.yml  (workflow_dispatch / manual)     │
│  └── calls → reusable_council_review.yml (org-infra)    │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────┐
│  org-infra: reusable_council_review.yml                 │
│                                                         │
│  ├── Download artifact (pr-diff.patch, pr-meta.json)    │
│  ├── WIF auth → Google Cloud (Vertex AI)                │
│  ├── Run council-review-action ─────────────────────┐   │
│  │                                                  │   │
│  │   ┌──────────────────────────────────────────┐   │   │
│  │   │  council-review-action (this repo)       │   │   │
│  │   │                                          │   │   │
│  │   │  1. Install OpenCode                     │   │   │
│  │   │  2. Filter noise → pr-diff-filtered.patch │   │   │
│  │   │  3. Annotate     → pr-diff-annotated.patch│   │   │
│  │   │  4. Pre-fetch PR context (CI, reviews)   │   │   │
│  │   │  5. Discover Divisor personas            │   │   │
│  │   │  6. Build prompt                         │   │   │
│  │   │  7. opencode run → review_raw.txt        │   │   │
│  │   │  8. Parse + filter → review_output.json  │   │   │
│  │   └──────────────────────────────────────┘   │   │
│  │                                              │   │
│  ├── Clean up previous bot comments ◄───────────┘   │
│  ├── Post review summary (issue comment)            │
│  └── Post inline comments (PR review comments)      │
└─────────────────────────────────────────────────────┘
```

### Three-workflow chain

The council review uses a three-file pattern for fork PR support:

| File | Location | Trigger | Purpose |
|---|---|---|---|
| `ci_council_review_collect.yml` | Synced to all repos | `pull_request` | Captures diff + metadata, no secrets needed |
| `ci_council_review.yml` | Synced to all repos | `workflow_run` / `workflow_dispatch` | Thin consumer — calls the reusable workflow |
| `reusable_council_review.yml` | org-infra only | `workflow_call` | Core logic — WIF auth, review, posting |

Fork PRs trigger `pull_request` on the fork (no secrets). The consumer workflow runs on the base repo where secrets are available. The reusable workflow stays in org-infra and is never synced downstream.

### Authentication

```
GitHub Actions runner
    │
    ▼  OIDC token exchange
GCP Workload Identity Federation (WIF)
    │
    ▼  Short-lived credentials
Vertex AI (Claude on Google Cloud)
    │
    ▼  opencode run --model google-vertex-anthropic/claude-sonnet-4-6
Review JSON output
```

### Persona discovery

The action auto-discovers Divisor reviewer personas in three tiers:

1. **Repo agents** — `.opencode/agents/divisor-*.md` in the PR's repo
2. **Bundled agents** — shipped with this action (fallback for repos without `uf init`)
3. **Single-agent mode** — general reviewer if no personas found

### Comment posting

| Type | API | Deletable? |
|---|---|---|
| Review summary | Issue comment (`POST /issues/{n}/comments`) | Yes — deleted on re-review |
| Inline findings | PR review comment (`POST /pulls/{n}/comments`) | Yes — deleted on re-review |
| Stale reviews | GraphQL `minimizeComment` | Collapsed as "outdated" |

All bot comments are tagged with `<!-- council-review-bot -->` for cleanup.

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
| `review-json` | Path to the review output JSON file |
| `review-mode` | `inline` (structured) or `comment` (fallback) |

## Directory structure

```
council-review-action/
├── action.yml              # Composite action definition
├── README.md               # This file
├── scripts/
│   ├── prepare-diff.sh     # Noise filter + line annotation
│   ├── build-prompt.sh     # Prompt construction
│   ├── run-review.sh       # OpenCode invocation
│   ├── prefetch.sh         # PR context pre-fetch (CI, reviews)
│   ├── extract-review-json.py  # JSON extraction from JSONL
│   └── filter-diff-lines.py    # Line number validation
├── test/
│   └── test-pipeline.sh    # Pipeline tests (39 assertions)
└── docs/
    ├── decisions.md         # Key technical decisions
    └── testing.md           # Test coverage and strategy
```

## Testing

```bash
cd council-review-action
bash test/test-pipeline.sh
```

See [docs/testing.md](docs/testing.md) for coverage details and what requires live credentials.
