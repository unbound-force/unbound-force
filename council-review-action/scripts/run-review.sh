#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
#
# Run council review via opencode run. OpenCode auto-discovers
# .opencode/ context (agents, commands, packs) and delegates to
# Divisor personas when available.
#
# Required env: MODEL
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

OPENCODE_EXIT=0
opencode run \
  --model "${MODEL}" \
  --format json \
  --file review_prompt.txt \
  -- "Review this PR according to the attached prompt." \
  > review_raw.txt 2>review_err.txt || OPENCODE_EXIT=$?

if [[ "${OPENCODE_EXIT}" -ne 0 ]]; then
  echo "::warning::OpenCode invocation failed (exit ${OPENCODE_EXIT})"
  cat review_err.txt >&2
fi

if [[ ! -s review_raw.txt ]]; then
  echo "::warning::OpenCode produced no output"
  echo '{"summary": "No output.", "inline_comments": []}' \
    > review_raw.txt
fi
