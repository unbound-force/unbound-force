#!/usr/bin/env bash
# Write the per-run opencode project config into $EVAL_WORKDIR and install
# the fixture's auto-question brief.
#
# Every agent is limited to the run's own directory and the temp areas.
# The rule is denied rather than asked so nothing waits on a human, and it
# sits at the root as well as on autonomy because subagents do not inherit
# autonomy's permissions. The real checkout is deliberately absent: a
# reviewer that could read it would be reviewing the wrong tree.
set -euo pipefail

work="${EVAL_WORKDIR:?}"
real="$(cd "$work" && pwd -P)"
data="${XDG_DATA_HOME:-$HOME/.local/share}/opencode"

allow=("$work/**" "$real/**" "$data/tool-output/**" "/tmp/**" "/private/tmp/**" "/var/tmp/**" "/private/var/tmp/**")
# .opencode/node_modules is a symlink into the source checkout.
if [ -n "${EVAL_REPO_ROOT:-}" ]; then
  allow+=("$EVAL_REPO_ROOT/.opencode/node_modules/**")
fi

rules='        "*": "deny"'
for path in "${allow[@]}"; do
  rules+=$',\n'"        \"$path\": \"allow\""
done

cat > "$work/.opencode/opencode.json" <<JSON
{
  "\$schema": "https://opencode.ai/config.json",
  "permission": {
    "doom_loop": "allow",
    "external_directory": {
$rules
    }
  },
  "agent": {
    "autonomy": {
      "permission": {
        "external_directory": {
$rules
        }
      }
    }
  }
}
JSON

if [ -f "$EVAL_FIXTURE_DIR/guidance.md" ]; then
  mkdir -p "$work/.opencode/auto-question/guidance"
  cp -L "$EVAL_FIXTURE_DIR/guidance.md" "$work/.opencode/auto-question/guidance/brief.md"
fi

# Run scaffolding must not appear in the reviewed diff or dirty the tree.
cat >> "$work/.git/info/exclude" <<'IGNORE'
/.opencode/opencode.json
/.opencode/auto-question/
IGNORE
