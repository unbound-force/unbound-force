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

# Validate model input against shell metacharacters
if ! [[ "${MODEL}" =~ ^[a-zA-Z0-9._/-]+$ ]]; then
  echo "::error::model contains invalid characters"
  exit 1
fi

PROVIDER="${MODEL%%/*}"
MODEL_NAME="${MODEL#*/}"

# Defense-in-depth sandbox (runtime permissions + plugin isolation).
# Permission denials are enforced by OpenCode's permission
# system at runtime, not by prompt instructions.
#
# Denied tools:
#   edit               — no file modifications (covers edit, write,
#                        and patch tools per OpenCode docs)
#   bash               — no shell command execution
#   webfetch           — no URL fetching
#   websearch          — no web search
#   skill              — no skill loading
#   external_directory — no access outside project (defaults to ask)
#   doom_loop          — no stuck-loop recovery (defaults to ask)
#
# NOT denied (intentionally):
#   read, glob, grep — allowed by OpenCode defaults
#   task — required for multi-agent Divisor subagent invocation;
#          denying it removes subagents from the Task tool
#          description, silently degrading to single-agent mode
#
# All permissions that default to "ask" (external_directory,
# doom_loop) are explicitly denied so no TTY prompt is needed.
# This avoids --dangerously-skip-permissions / --auto which
# would blanket-approve everything not denied.
# shellcheck disable=SC2016  # JSON string, no shell expansion intended
PERMISSION_CONFIG='{
  "$schema": "https://opencode.ai/config.json",
  "permission": {
    "edit": "deny",
    "bash": "deny",
    "webfetch": "deny",
    "websearch": "deny",
    "skill": "deny",
    "external_directory": "deny",
    "doom_loop": "deny"
  }
}'

if [[ "${PROVIDER}" == "google-vertex-anthropic" ]]; then
  # Merge Vertex provider config with permission config
  export OPENCODE_CONFIG_CONTENT
  OPENCODE_CONFIG_CONTENT=$(jq -n \
    --arg model "${MODEL_NAME}" \
    --argjson perms "${PERMISSION_CONFIG}" \
    '$perms + {
      "provider": {
        "google-vertex-anthropic": {
          "models": {
            ($model): {}
          }
        }
      }
    }')
else
  export OPENCODE_CONFIG_CONTENT="${PERMISSION_CONFIG}"
fi

# Unset GH_TOKEN before invoking OpenCode to limit blast radius
# if prompt injection triggers tool execution. GH_TOKEN is not
# needed by the review invocation (no gh CLI calls).
# NOTE: GOOGLE_APPLICATION_CREDENTIALS must NOT be unset — it
# points to the WIF credential file that the Vertex AI SDK
# reads for authentication. Unsetting it breaks auth.
unset GH_TOKEN 2>/dev/null || true

# --pure: skip external MCP plugins (closes plugin bypass vector)
OPENCODE_EXIT=0
timeout 300 opencode run \
  --model "${MODEL}" \
  --format json \
  --pure \
  --file review_prompt.txt \
  -- "Review this PR according to the attached prompt." \
  > review_raw.txt 2>review_err.txt || OPENCODE_EXIT=$?

# Exit code 124 is coreutils timeout's sentinel: it means the
# child process was killed after the time limit expired.
if [[ "${OPENCODE_EXIT}" -eq 124 ]]; then
  echo "::warning::OpenCode timed out after 300s"
  cat review_err.txt >&2
elif [[ "${OPENCODE_EXIT}" -ne 0 ]]; then
  echo "::warning::OpenCode invocation failed (exit ${OPENCODE_EXIT})"
  cat review_err.txt >&2
fi

if [[ ! -s review_raw.txt ]]; then
  echo "::warning::OpenCode produced no output"
  echo '{"summary": "No output.", "inline_comments": []}' \
    > review_raw.txt
fi
