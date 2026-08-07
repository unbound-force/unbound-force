# Architecture

## End-to-end flow

```
┌──────────────────────────────────────────────────────────┐
│  Downstream Repo (e.g., org-infra, gaze)                 │
│                                                          │
│  ci_council_review_collect.yml  (pull_request trigger)   │
│  ├── Gate: skip bots, skip drafts                        │
│  ├── Capture diff: gh pr diff → pr-diff.patch            │
│  ├── Build metadata: pr-meta.json                        │
│  └── Upload artifact: council-review-diff                │
│                                                          │
│  ci_council_review.yml  (workflow_dispatch / manual)      │
│  └── calls → reusable_council_review.yml (org-infra)     │
└─────────────────────────┬────────────────────────────────┘
                          │
┌─────────────────────────▼────────────────────────────────┐
│  org-infra: reusable_council_review.yml                  │
│                                                          │
│  ├── Download artifact (pr-diff.patch, pr-meta.json)     │
│  ├── WIF auth → Google Cloud (Vertex AI)                 │
│  ├── Run council-review-action ────────────────────┐     │
│  │                                                 │     │
│  │   ┌─────────────────────────────────────────┐   │     │
│  │   │  council-review-action (this repo)      │   │     │
│  │   │                                         │   │     │
│  │   │  1. Install OpenCode CLI                │   │     │
│  │   │  2. Filter + annotate diff              │   │     │
│  │   │  3. Pre-fetch PR context (CI, reviews)  │   │     │
│  │   │  4. Discover Divisor personas           │   │     │
│  │   │  5. Build prompt                        │   │     │
│  │   │  6. opencode run (sandboxed)            │   │     │
│  │   │  7. Parse + filter → review_output.json │   │     │
│  │   └─────────────────────────────────────────┘   │     │
│  │                                                 │     │
│  ├── Clean up previous bot comments ◄──────────────┘     │
│  ├── Post review summary (issue comment)                 │
│  └── Post inline comments (PR review comments)           │
└──────────────────────────────────────────────────────────┘
```

## Three-workflow chain

The council review uses a three-file pattern for fork PR
support:

| File | Location | Trigger | Purpose |
|---|---|---|---|
| `ci_council_review_collect.yml` | Synced to all repos | `pull_request` | Captures diff + metadata, no secrets needed |
| `ci_council_review.yml` | Synced to all repos | `workflow_run` / `workflow_dispatch` | Thin consumer — calls the reusable workflow |
| `reusable_council_review.yml` | org-infra only | `workflow_call` | Core logic — WIF auth, review, posting |

Fork PRs trigger `pull_request` on the fork (no secrets). The
consumer workflow runs on the base repo where secrets are
available. The reusable workflow stays in org-infra and is
never synced downstream.

## Authentication

```
GitHub Actions runner
    │
    ▼  OIDC token exchange
GCP Workload Identity Federation (WIF)
    │
    ▼  Short-lived credentials
Vertex AI (Claude on Google Cloud)
    │
    ▼  opencode run --pure --model google-vertex-anthropic/claude-sonnet-4-6
Review JSON output
```

Authentication is the consumer's responsibility. The action
expects Vertex AI credentials to be available in the
environment (via `GOOGLE_APPLICATION_CREDENTIALS`).

## Persona discovery

The action auto-discovers Divisor reviewer personas in three
tiers:

1. **Repo agents** — `.opencode/agents/divisor-*.md` in the
   PR's repo
2. **Bundled agents** — shipped with this action (fallback for
   repos without `uf init`)
3. **Single-agent mode** — general reviewer if no personas
   found

## Runtime sandbox

The action processes untrusted diff content. Three layers
prevent tool misuse:

1. **Runtime permissions** — `OPENCODE_CONFIG_CONTENT` denies
   bash, edit, webfetch, websearch, and skill at the OpenCode
   runtime level
2. **Plugin isolation** — `--pure` flag prevents external MCP
   plugins from loading
3. **Agent frontmatter** — each Divisor agent's `permission:`
   config denies dangerous tools as defense-in-depth

`GH_TOKEN` is unset before `opencode run` to remove GitHub
API access. `GOOGLE_APPLICATION_CREDENTIALS` cannot be unset
(breaks Vertex AI auth) but bash is denied at the runtime
level and `external_directory: deny` confines file reads to
the project workspace. On GitHub-hosted runners the WIF
credential file resides in `RUNNER_TEMP` outside the
workspace. See [security-risks.md](security-risks.md) A3
for the full threat model.

The review runs with a 300-second timeout. OpenCode CLI is
pinned to `opencode-ai@1.15.13` (JSONL format compatibility
boundary — see [decisions.md](decisions.md),
"OpenCode version pinning").

See [security-risks.md](security-risks.md) for the full
risk register.

## Output handling

The action outputs structured JSON and a review mode. It does
**not** post comments — that is the consumer workflow's
responsibility (see [decisions.md](decisions.md),
"Comment cleanup: delete + minimize").

| Output | Description |
|---|---|
| `review-json` | Path to the review JSON file |
| `review-mode` | `inline` (structured) or `comment` (fallback) |

The consumer workflow uses these outputs to post:

| Type | API | Cleanup |
|---|---|---|
| Review summary | Issue comment (`POST /issues/{n}/comments`) | Deleted on re-review |
| Inline findings | PR review comment (`POST /pulls/{n}/comments`) | Deleted on re-review |
| Stale reviews | GraphQL `minimizeComment` | Collapsed as "outdated" |

The consumer tags all bot comments with
`<!-- council-review-bot -->` for cleanup on subsequent runs.
