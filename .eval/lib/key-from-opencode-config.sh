#!/usr/bin/env bash
# Print the EnMaaS API key held in the opencode config in
# $OPENCODE_CONFIG_CONTENT (the org secret CI already uses).
#
# Only the key is taken: the eval runtime builds its own opencode config,
# and passing this one through would let its permissions, plugins, or
# model override the pinned runtime (design.md D3, EV-FR-007).
#
# Errors name the problem but never print anything from the config.
set -euo pipefail

: "${OPENCODE_CONFIG_CONTENT:?OPENCODE_CONFIG_CONTENT is not set}"

enmaas='[(.provider // {}) | to_entries[] | .value
         | select((.options.baseURL // "") | test("enmaas"; "i"))]'

if ! count="$(printf '%s' "$OPENCODE_CONFIG_CONTENT" | jq "$enmaas | length" 2>/dev/null)"; then
  echo "OPENCODE_CONFIG_CONTENT is not an opencode JSON config" >&2
  exit 1
fi
key="$(printf '%s' "$OPENCODE_CONFIG_CONTENT" | jq -r "$enmaas | .[0].options.apiKey // \"\"")"

if [ "$count" -eq 0 ]; then
  echo "OPENCODE_CONFIG_CONTENT has no provider whose baseURL is the EnMaaS endpoint" >&2
  exit 1
fi
if [ -z "$key" ]; then
  echo "the EnMaaS provider in OPENCODE_CONFIG_CONTENT has no options.apiKey" >&2
  exit 1
fi
case "$key" in
  '{env:'* | '{file:'*)
    echo "the EnMaaS provider's apiKey in OPENCODE_CONFIG_CONTENT is a reference, not a literal key" >&2
    exit 1
    ;;
esac

printf '%s\n' "$key"
