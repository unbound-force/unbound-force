#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
#
# Run council review via opencode run. OpenCode auto-discovers
# .opencode/ context (agents, commands, packs). In multi-agent
# mode, the prompt instructs the orchestrator to delegate to
# each Divisor persona. In single-agent mode, it reviews as a
# general reviewer.
#
# Required env: MODEL, AGENT_MODE
# Optional env: GOOGLE_CLOUD_PROJECT, VERTEX_LOCATION,
#               GOOGLE_APPLICATION_CREDENTIALS
# TODO: restore set -euo pipefail once OpenCode+Vertex is stable
set -uo pipefail

PROVIDER="${MODEL%%/*}"
MODEL_NAME="${MODEL#*/}"

if [[ "${PROVIDER}" == "google-vertex-anthropic" ]]; then
  export OPENCODE_CONFIG_CONTENT
  OPENCODE_CONFIG_CONTENT=$(cat <<OCEOF
{
  "\$schema": "https://opencode.ai/config.json",
  "provider": {
    "google-vertex-anthropic": {
      "models": {
        "${MODEL_NAME}": {}
      }
    }
  }
}
OCEOF
)
fi

PROMPT=$(cat review_prompt.txt)

if ! opencode run "${PROMPT}" \
  --model "${MODEL}" \
  --format json \
  --dangerously-skip-permissions \
  > review_raw.txt 2>review_err.txt; then
  echo "::warning::OpenCode invocation failed (exit $?)"
  cat review_err.txt >&2
fi

if [[ ! -s review_raw.txt ]]; then
  echo "::warning::OpenCode produced no output"
  echo '{"summary": "No output.", "inline_comments": []}' \
    > review_raw.txt
fi
