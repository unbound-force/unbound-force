#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
#
# Generate OpenPackage packages from canonical .opencode/ sources.
# Output: packages/review-council/ and packages/workflows/
#
# Usage: ./scripts/generate-packages.sh
#        make packages

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PACKAGES_DIR="$REPO_ROOT/packages"

strip_scaffold_lines() {
	sed '/<!-- scaffolded by/d'
}

python_transform_agent_frontmatter() {
	# Usage: python_transform_agent_frontmatter <src.md> <claude_short_name> <claude_tools_csv>
	local src="$1"
	local claude_name="$2"
	local claude_tools="$3"
	REPO_SRC="$src" AGENT_OPKG_CLAUDE_NAME="$claude_name" AGENT_OPKG_CLAUDE_TOOLS="$claude_tools" python3 <<'PY'
import os
import re
import sys

try:
	import yaml  # type: ignore
except ImportError:
	sys.stderr.write("generate-packages.sh: Python PyYAML is required (import yaml).\n")
	sys.exit(1)

path = os.environ["REPO_SRC"]
claude_name = os.environ["AGENT_OPKG_CLAUDE_NAME"]
claude_tools = os.environ["AGENT_OPKG_CLAUDE_TOOLS"]

text = open(path, encoding="utf-8").read()
if not text.startswith("---\n"):
	sys.stderr.write(f"{path}: expected YAML frontmatter starting with ---\n")
	sys.exit(1)
rest = text[4:]
end = rest.find("\n---\n")
if end == -1:
	sys.stderr.write(f"{path}: missing closing --- for frontmatter\n")
	sys.exit(1)
fm_inner = rest[:end]
body = rest[end + 5 :]

try:
	data = yaml.safe_load(fm_inner)
except yaml.YAMLError as e:
	sys.stderr.write(f"{path}: frontmatter YAML error: {e}\n")
	sys.exit(1)

if not isinstance(data, dict):
	sys.stderr.write(f"{path}: frontmatter must be a mapping\n")
	sys.exit(1)

out = {
	"description": data.get("description"),
	"openpackage": {
		"opencode": {},
		"claude": {"name": claude_name, "tools": claude_tools},
		"cursor": {"mode": "agent"},
	},
}
oc = out["openpackage"]["opencode"]
if "mode" in data:
	oc["mode"] = data["mode"]
if "temperature" in data:
	oc["temperature"] = data["temperature"]
if "tools" in data and isinstance(data["tools"], dict):
	oc["tools"] = data["tools"]

fm_out = yaml.dump(
	out,
	sort_keys=False,
	default_flow_style=False,
	allow_unicode=True,
	width=120,
).rstrip()
if not fm_out.endswith("\n"):
	fm_out += "\n"

sys.stdout.write("---\n")
sys.stdout.write(fm_out)
sys.stdout.write("---\n")
# strip scaffold markers from body
for line in body.splitlines():
	if "<!-- scaffolded by" in line:
		continue
	sys.stdout.write(line + "\n")
PY
}

copy_stripped() {
	strip_scaffold_lines <"$1" >"$2"
}

rm -rf "$PACKAGES_DIR"

# ---------- review-council ----------

RC="$PACKAGES_DIR/review-council"
mkdir -p "$RC/agents/review-council" "$RC/commands/review-council" "$RC/rules/review-council"

readonly RC_AGENTS=(
	divisor-guard
	divisor-architect
	divisor-adversary
	divisor-sre
	divisor-testing
)
for agent in "${RC_AGENTS[@]}"; do
	short="${agent#divisor-}"
	python_transform_agent_frontmatter \
		"$REPO_ROOT/.opencode/agents/${agent}.md" \
		"$short" \
		"Read, Grep, Glob" >"$RC/agents/review-council/${agent}.md"
done

# Write-capable personas (explicit loops for bash 3 portability)
python_transform_agent_frontmatter \
	"$REPO_ROOT/.opencode/agents/divisor-scribe.md" \
	scribe \
	"Read, Edit, Write, Grep, Glob" >"$RC/agents/review-council/divisor-scribe.md"
python_transform_agent_frontmatter \
	"$REPO_ROOT/.opencode/agents/divisor-herald.md" \
	herald \
	"Read, Edit, Write, Grep, Glob" >"$RC/agents/review-council/divisor-herald.md"
python_transform_agent_frontmatter \
	"$REPO_ROOT/.opencode/agents/divisor-envoy.md" \
	envoy \
	"Read, Edit, Write, Grep, Glob" >"$RC/agents/review-council/divisor-envoy.md"

python_transform_agent_frontmatter \
	"$REPO_ROOT/.opencode/agents/divisor-curator.md" \
	"curator" \
	"Read, Bash, Grep, Glob" >"$RC/agents/review-council/divisor-curator.md"

copy_stripped "$REPO_ROOT/.opencode/command/review-council.md" "$RC/commands/review-council/review-council.md"
copy_stripped "$REPO_ROOT/.opencode/command/review-pr.md" "$RC/commands/review-council/review-pr.md"

copy_stripped "$REPO_ROOT/.opencode/uf/packs/severity.md" "$RC/rules/review-council/severity.md"
copy_stripped "$REPO_ROOT/.opencode/uf/packs/default.md" "$RC/rules/review-council/default.md"
copy_stripped "$REPO_ROOT/.opencode/uf/packs/default-custom.md" "$RC/rules/review-council/default-custom.md"

cat >"$RC/mcp.jsonc" <<'MCPEOF'
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

cat >"$RC/openpackage.yml" <<'YMLEOF'
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

cat >"$RC/README.md" <<'READMEEOF'
# @unbound-force/review-council

AI code review council — 9 reviewer personas audit your code
in parallel for security, architecture, testing, operations,
intent drift, and documentation completeness.

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

Plus 2 commands (`/review-council`, `/review-pr`) and 3 convention
packs (`severity`, `default`, `default-custom`).

## License

Apache-2.0
READMEEOF

# ---------- workflows ----------

WF="$PACKAGES_DIR/workflows"
mkdir -p "$WF/agents/workflows" "$WF/commands/workflows"

python_transform_agent_frontmatter \
	"$REPO_ROOT/.opencode/agents/constitution-check.md" \
	"constitution-check" \
	"Read, Grep, Glob" >"$WF/agents/workflows/constitution-check.md"

for f in "$REPO_ROOT"/.opencode/command/speckit.*.md \
	"$REPO_ROOT"/.opencode/command/opsx-*.md \
	"$REPO_ROOT/.opencode/command/constitution-check.md"; do
	if [[ ! -f "$f" ]]; then
		echo "Missing expected command file: $f" >&2
		exit 1
	fi
	name=$(basename "$f")
	copy_stripped "$f" "$WF/commands/workflows/$name"
done

cat >"$WF/openpackage.yml" <<'YMLEOF'
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

cat >"$WF/README.md" <<'READMEEOF'
# @unbound-force/workflows

Spec-driven development workflows for AI-assisted software
engineering. Two tiers:

- **Speckit** (strategic): multi-phase pipeline for features with
  several user stories
- **OpenSpec** (tactical): lightweight workflow for bug fixes and
  small changes

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

This also pulls in `@unbound-force/review-council` as a dependency.

## Speckit Commands

| Command | Purpose |
|:---|:---|
| `/speckit.constitution` | Create or update project constitution |
| `/speckit.specify` | Create feature specification |
| `/speckit.clarify` | Reduce spec ambiguity |
| `/speckit.plan` | Generate implementation plan |
| `/speckit.tasks` | Break plan into ordered tasks |
| `/speckit.analyze` | Cross-artifact consistency check |
| `/speckit.checklist` | Quality validation |
| `/speckit.implement` | Execute tasks |
| `/speckit.taskstoissues` | Convert tasks to GitHub Issues |
| `/speckit.testreview` | Test review pass |

## OpenSpec Commands

| Command | Purpose |
|:---|:---|
| `/opsx-propose` | Create change proposal with plan and tasks |
| `/opsx-explore` | Think through ideas (read-only exploration) |
| `/opsx-apply` | Implement tasks from a change |
| `/opsx-archive` | Archive a completed change |

## Other

| Command | Purpose |
|:---|:---|
| `/constitution-check` | Hero vs org constitution alignment |

## License

Apache-2.0
READMEEOF

rc_count=$(find "$RC" -type f | wc -l)
wf_count=$(find "$WF" -type f | wc -l)
echo "Generated packages/ from .opencode/ source files:"
echo "  review-council: $rc_count files"
echo "  workflows:      $wf_count files"
echo "  total:          $((rc_count + wf_count)) files"
