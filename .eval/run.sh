#!/usr/bin/env bash
# Run eval fixtures in the isolated environment.
#
# Usage: .eval/run.sh [run-eval args]
# Example: .eval/run.sh --only review-council-clean --output /tmp/out
#
# Needs ENMAAS_KEY. The driver comes from the eval-infra commit pinned in
# .eval/lib/pins.sh; set EVAL_DRIVER to a run-eval.ts to use another one.
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# shellcheck source=.eval/lib/pins.sh
. "$root/.eval/lib/pins.sh"
model="${EVAL_MODEL:-enmaas/rits/zai-org/glm-5-3}"
cache="${EVAL_CACHE:-$HOME/.cache/uf-eval}"

# Tools and API keys: chezmoi's env.sh locally, plain env vars in CI.
if [ -r "$HOME/.config/shell/env.sh" ]; then
  # shellcheck disable=SC1091
  . "$HOME/.config/shell/env.sh"
fi
: "${ENMAAS_KEY:?ENMAAS_KEY must be set}"

version="$(opencode --version)"
if [ "$version" != "$EVAL_OPENCODE_VERSION" ]; then
  echo "opencode $version at $(command -v opencode); evals are pinned to $EVAL_OPENCODE_VERSION" >&2
  exit 1
fi

xdg="$(bash "$root/.eval/lib/prepare-eval-env.sh")"
export XDG_CONFIG_HOME="$xdg"
driver="${EVAL_DRIVER:-$cache/plugins/eval-infra-$EVAL_DRIVER_SHA/driver/run-eval.ts}"

cd "$root"
echo "opencode $version ($(command -v opencode)), driver $driver"
exec bun run "$driver" --fixtures "$root/.eval/fixtures" --model "$model" \
  --answerer plugin --agent autonomy --require-tools compress "$@"
