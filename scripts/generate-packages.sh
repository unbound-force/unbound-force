#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
#
# Generate OpenPackage packages from the canonical .opencode/ source files.
# Single source of truth: .opencode/agents/, .opencode/command/, .opencode/uf/packs/
# Output: packages/review-council/ and packages/workflows/
#
# Usage: ./scripts/generate-packages.sh
#        make packages

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PACKAGES_DIR="$REPO_ROOT/packages"

# ---------- helpers ----------

strip_scaffold_marker() {
  sed '/<!-- scaffolded by/d'
}

# Transform OpenCode-native frontmatter to OpenPackage cross-platform format.
# Reads a .md file, rewrites the YAML frontmatter between --- markers,
# preserves all body content below.
transform_frontmatter() {
  local src="$1"
  local claude_name="$2"
  local claude_tools="$3"

  local in_frontmatter=false
  local frontmatter_done=false
  local frontmatter=""
  local body=""
  local line_num=0

  while IFS= read -r line || [[ -n "$line" ]]; do
    line_num=$((line_num + 1))
    if [[ $line_num -eq 1 && "$line" == "---" ]]; then
      in_frontmatter=true
      continue
    fi
    if $in_frontmatter && [[ "$line" == "---" ]]; then
      in_frontmatter=false
      frontmatter_done=true
      continue
    fi
    if $in_frontmatter; then
      frontmatter+="$line"$'\n'
    fi
    if $frontmatter_done; then
      body+="$line"$'\n'
    fi
  done < "$src"

  local description
  description=$(echo "$frontmatter" | grep '^description:' | sed 's/^description: *//')

  local mode temp tools_block
  mode=$(echo "$frontmatter" | grep '^mode:' | sed 's/^mode: *//')
  temp=$(echo "$frontmatter" | grep '^temperature:' | sed 's/^temperature: *//')

  # Extract tools block (indented lines after "tools:")
  tools_block=$(echo "$frontmatter" | sed -n '/^tools:/,/^[^ ]/p' | head -n -1)
  if [[ -z "$tools_block" ]]; then
    tools_block=$(echo "$frontmatter" | grep -A 20 '^tools:' || true)
  fi

  # Build the opencode tools section
  local oc_tools=""
  while IFS= read -r tline; do
    [[ "$tline" == "tools:" ]] && continue
    [[ -z "$tline" ]] && continue
    oc_tools+="      $tline"$'\n'
  done <<< "$tools_block"

  # Write transformed file
  {
    echo "---"
    echo "description: $description"
    echo "openpackage:"
    echo "  opencode:"
    [[ -n "$mode" ]] && echo "    mode: $mode"
    [[ -n "$temp" ]] && echo "    temperature: $temp"
    if [[ -n "$oc_tools" ]]; then
      echo "    tools:"
      printf '%s' "$oc_tools"
    fi
    echo "  claude:"
    echo "    name: $claude_name"
    echo "    tools: $claude_tools"
    echo "  cursor:"
    echo "    mode: agent"
    echo "---"
    printf '%s' "$body"
  }
}

copy_stripped() {
  strip_scaffold_marker < "$1" > "$2"
}

# ---------- clean ----------

rm -rf "$PACKAGES_DIR"

# ---------- review-council ----------

RC="$PACKAGES_DIR/review-council"
mkdir -p "$RC/agents/review-council" "$RC/commands/review-council" "$RC/rules/review-council"

# Agents: divisor-* (read-only personas)
readonly_agents=(divisor-guard divisor-architect divisor-sre divisor-testing)
for agent in "${readonly_agents[@]}"; do
  short="${agent#divisor-}"
  transform_frontmatter \
    "$REPO_ROOT/.opencode/agents/${agent}.md" \
    "$short" \
    "Read, Grep, Glob" \
    | strip_scaffold_marker > "$RC/agents/review-council/${agent}.md"
done

# Agents: divisor-adversary (read-only, extra tool fields)
transform_frontmatter \
  "$REPO_ROOT/.opencode/agents/divisor-adversary.md" \
  "adversary" \
  "Read, Grep, Glob" \
  | strip_scaffold_marker > "$RC/agents/review-council/divisor-adversary.md"

# Agents: write-capable content agents
declare -A write_agents=(
  [divisor-scribe]="scribe"
  [divisor-herald]="herald"
  [divisor-envoy]="envoy"
)
for agent in "${!write_agents[@]}"; do
  transform_frontmatter \
    "$REPO_ROOT/.opencode/agents/${agent}.md" \
    "${write_agents[$agent]}" \
    "Read, Edit, Write, Grep, Glob" \
    | strip_scaffold_marker > "$RC/agents/review-council/${agent}.md"
done

# Agents: curator (read + bash)
transform_frontmatter \
  "$REPO_ROOT/.opencode/agents/divisor-curator.md" \
  "curator" \
  "Read, Bash, Grep, Glob" \
  | strip_scaffold_marker > "$RC/agents/review-council/divisor-curator.md"

# Commands
copy_stripped "$REPO_ROOT/.opencode/command/review-council.md" "$RC/commands/review-council/review-council.md"
copy_stripped "$REPO_ROOT/.opencode/command/review-pr.md" "$RC/commands/review-council/review-pr.md"

# Rules
copy_stripped "$REPO_ROOT/.opencode/uf/packs/severity.md" "$RC/rules/review-council/severity.md"
copy_stripped "$REPO_ROOT/.opencode/uf/packs/default.md" "$RC/rules/review-council/default.md"
copy_stripped "$REPO_ROOT/.opencode/uf/packs/default-custom.md" "$RC/rules/review-council/default-custom.md"

# MCP config
cat > "$RC/mcp.jsonc" << 'MCPEOF'
{
  // Dewey semantic knowledge layer — optional.
  // All review agents gracefully degrade if unavailable.
  // Install: brew install unbound-force/tap/dewey
  "dewey": {
    "type": "local",
    "command": ["dewey", "serve", "--vault", "."],
    "enabled": true
  }
}
MCPEOF

# Manifest
cat > "$RC/openpackage.yml" << 'YMLEOF'
name: "@unbound-force/review-council"
version: 0.1.0
description: >
  AI code review council — 9 reviewer personas audit your
  code for security, architecture, testing, operations,
  intent drift, and documentation completeness.
keywords:
  - code-review
  - ai-review
  - security
  - architecture
  - testing
  - operations
author: unbound-force
license: Apache-2.0
YMLEOF

# README
cat > "$RC/README.md" << 'READMEEOF'
# @unbound-force/review-council

AI code review council -- 9 reviewer personas audit your
code in parallel for security, architecture, testing,
operations, intent drift, and documentation completeness.

## Install

```bash
opkg install @unbound-force/review-council
```

Or add to your project's `openpackage.yml`:

```yaml
dependencies:
- name: "@unbound-force/review-council"
  version: ^0.1.0
```

## What You Get

| Persona | Agent | Focus |
|:---|:---|:---|
| The Guard | `divisor-guard` | Intent drift, zero-waste, constitution alignment |
| The Architect | `divisor-architect` | Structure, conventions, DRY, patterns |
| The Adversary | `divisor-adversary` | Secrets, CVEs, error handling, injection safety |
| The Operator | `divisor-sre` | Deployment, performance, dependencies, observability |
| The Tester | `divisor-testing` | Test architecture, coverage, assertions, isolation |
| The Curator | `divisor-curator` | Documentation gaps, blog/tutorial opportunities |
| The Scribe | `divisor-scribe` | Technical documentation (READMEs, API docs) |
| The Herald | `divisor-herald` | Blog posts, release notes, announcements |
| The Envoy | `divisor-envoy` | Press releases, social media, community updates |

Plus 2 commands (`/review-council`, `/review-pr`) and
3 convention packs (severity, default, default-custom).

Auto-converts to OpenCode, Cursor, Claude Code, Gemini
CLI, and 30+ other platforms.

## License

Apache-2.0
READMEEOF

# ---------- workflows ----------

WF="$PACKAGES_DIR/workflows"
mkdir -p "$WF/agents/workflows" "$WF/commands/workflows"

# Agent: constitution-check
transform_frontmatter \
  "$REPO_ROOT/.opencode/agents/constitution-check.md" \
  "constitution-check" \
  "Read, Grep, Glob" \
  | strip_scaffold_marker > "$WF/agents/workflows/constitution-check.md"

# Commands: speckit + opsx + constitution-check
for f in "$REPO_ROOT"/.opencode/command/speckit.*.md \
         "$REPO_ROOT"/.opencode/command/opsx-*.md \
         "$REPO_ROOT/.opencode/command/constitution-check.md"; do
  name=$(basename "$f")
  copy_stripped "$f" "$WF/commands/workflows/$name"
done

# Manifest
cat > "$WF/openpackage.yml" << 'YMLEOF'
name: "@unbound-force/workflows"
version: 0.1.0
description: >
  Spec-driven development workflows — Speckit pipeline
  for features and OpenSpec workflow for bug fixes.
  Two tiers, one package.
keywords:
  - specification
  - workflow
  - speckit
  - openspec
  - planning
author: unbound-force
license: Apache-2.0
dependencies:
- name: "@unbound-force/review-council"
  version: ^0.1.0
YMLEOF

# README
cat > "$WF/README.md" << 'READMEEOF'
# @unbound-force/workflows

Spec-driven development workflows for AI-assisted
software engineering. Two tiers:

- **Speckit** (strategic): 9-phase pipeline for features
  with 3+ user stories
- **OpenSpec** (tactical): Lightweight workflow for bug
  fixes and small changes

## Install

```bash
opkg install @unbound-force/workflows
```

Or add to your project's `openpackage.yml`:

```yaml
dependencies:
- name: "@unbound-force/workflows"
  version: ^0.1.0
```

This also installs `@unbound-force/review-council` as a
dependency.

## Speckit Commands

| Command | Purpose |
|:---|:---|
| `/speckit.constitution` | Create/update project constitution |
| `/speckit.specify` | Create feature specification |
| `/speckit.clarify` | Reduce spec ambiguity |
| `/speckit.plan` | Generate implementation plan |
| `/speckit.tasks` | Break plan into ordered tasks |
| `/speckit.analyze` | Cross-artifact consistency check |
| `/speckit.checklist` | Quality validation |
| `/speckit.implement` | Execute tasks |
| `/speckit.taskstoissues` | Convert tasks to GitHub Issues |

## OpenSpec Commands

| Command | Purpose |
|:---|:---|
| `/opsx-propose` | Create change proposal with plan and tasks |
| `/opsx-explore` | Think through ideas (read-only) |
| `/opsx-apply` | Implement tasks from a change |
| `/opsx-archive` | Archive a completed change |

## License

Apache-2.0
READMEEOF

# ---------- summary ----------

rc_count=$(find "$RC" -type f | wc -l)
wf_count=$(find "$WF" -type f | wc -l)
echo "Generated packages/ from .opencode/ source files:"
echo "  review-council: $rc_count files"
echo "  workflows:      $wf_count files"
echo "  total:          $((rc_count + wf_count)) files"
