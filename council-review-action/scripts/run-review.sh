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
set -euo pipefail

PROVIDER="${MODEL%%/*}"
MODEL_NAME="${MODEL#*/}"

if [[ "${PROVIDER}" == "google-vertex-anthropic" ]]; then
  export OPENCODE_CONFIG_CONTENT
  OPENCODE_CONFIG_CONTENT=$(jq -n \
    --arg model "${MODEL_NAME}" \
    '{
      "$schema": "https://opencode.ai/config.json",
      "provider": {
        "google-vertex-anthropic": {
          "models": {
            ($model): {}
          }
        }
      }
    }')
fi

# Unset sensitive credentials before invoking OpenCode to limit
# the blast radius if prompt injection triggers tool execution.
# WIF credentials for Vertex AI are inherited via the environment
# automatically; GH_TOKEN and explicit credential paths are not
# needed by the review invocation.
unset GH_TOKEN GOOGLE_APPLICATION_CREDENTIALS 2>/dev/null || true

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
