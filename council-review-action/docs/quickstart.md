# Quick Start

Add AI code review to your repo in three steps.

## Prerequisites

- A GitHub org with [Vertex AI](https://cloud.google.com/vertex-ai)
  access via Workload Identity Federation (WIF)
- A reusable workflow that handles WIF auth and comment posting
  (see [architecture.md](architecture.md) for the
  three-workflow chain)

## 1. Scaffold Divisor personas

Run `uf init` in your repo to scaffold the `.opencode/agents/`
directory with Divisor reviewer personas. This is optional — the
action falls back to bundled agents if none are found.

```bash
uf init
```

## 2. Add the collect workflow

Create `.github/workflows/ci_council_review_collect.yml` in your
repo. This workflow runs on every PR, captures the diff and
metadata, and uploads them as an artifact:

```yaml
name: Council Review Collect

on:
  pull_request:
    branches: [main]

permissions:
  contents: read

jobs:
  collect:
    if: >-
      github.actor != 'dependabot[bot]'
      && !github.event.pull_request.draft
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1  # v7.0.1

      - name: Capture diff
        env:
          GH_TOKEN: ${{ github.token }}
        run: |
          gh pr diff ${{ github.event.pull_request.number }} \
            > pr-diff.patch
          jq -n \
            --arg number "${{ github.event.pull_request.number }}" \
            --arg repo "${{ github.repository }}" \
            --arg title "${{ github.event.pull_request.title }}" \
            '{number: ($number | tonumber), repo: $repo, title: $title}' \
            > pr-meta.json

      - uses: actions/upload-artifact@v4
        with:
          name: council-review-diff
          path: |
            pr-diff.patch
            pr-meta.json
```

## 3. Call the action from your reusable workflow

In your org's reusable workflow (after WIF auth), add a step
that calls the council-review-action:

```yaml
- name: Run council review
  id: review
  uses: unbound-force/unbound-force/council-review-action@dc82f77d590a0fa68dd6e2344f6a9569087dedf5  # main
  with:
    diff-path: pr-diff.patch
    meta-path: pr-meta.json
    github-token: ${{ secrets.GITHUB_TOKEN }}
```

The action outputs `review-json` (path to the review JSON file)
and `review-mode` (`inline` or `comment`). Use these to post
the review as inline PR comments or a summary comment.

## What happens

1. The action installs OpenCode CLI (`opencode-ai@1.15.13`)
2. Filters noise from the diff (lock files, vendor, generated)
3. Annotates diff lines with `[L<N>]` prefixes for accurate
   inline comment placement
4. Pre-fetches PR context (CI checks, existing reviews, linked
   issues) via `gh` CLI
5. Discovers Divisor reviewer personas from
   `.opencode/agents/divisor-*.md`
6. Builds a review prompt referencing the repo's methodology
   files
7. Invokes `opencode run` with a runtime permission sandbox
   (bash, edit, webfetch denied)
8. Parses structured JSON output (summary + inline comments)

## Customization

### Model

Override the default model via the `model` input:

```yaml
with:
  model: "anthropic/claude-sonnet-4-6"
```

### Agent pattern

Override which agents are discovered for multi/single mode
detection. Note: the review prompt still references
`divisor-*.md` in its methodology instructions, so custom
patterns only affect whether multi-agent mode is activated:

```yaml
with:
  agents-pattern: ".opencode/agents/review-*.md"
```

## Next steps

- [README](../README.md) — inputs/outputs, security overview,
  directory structure
- [Architecture](architecture.md) — end-to-end flow,
  three-workflow chain, runtime sandbox
- [Security](security-risks.md) — risk register and
  defense-in-depth layers
- [Decisions](decisions.md) — key technical decisions and
  rationale
- [Testing](testing.md) — test coverage and strategy
