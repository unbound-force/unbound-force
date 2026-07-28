#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
#
# Parse OpenCode review output into structured JSON. Handles three
# cases: direct JSON, JSONL streaming format, and plain text
# fallback. Filters inline comments to diff-visible lines.
#
# Required env: DIFF_PATH, SCRIPT_DIR
# Writes: review_output.json
# Outputs to GITHUB_OUTPUT: review_json, review_mode
set -euo pipefail

# Fallback GITHUB_OUTPUT for local testing
: "${GITHUB_OUTPUT:=/dev/null}"

if [[ ! -f review_raw.txt ]] || [[ ! -s review_raw.txt ]]; then
  jq -n '{summary: "Review skipped.", inline_comments: []}' \
    > review_output.json
  echo "review_json=review_output.json" >> "$GITHUB_OUTPUT"
  echo "review_mode=comment" >> "$GITHUB_OUTPUT"
  echo "::notice::Review skipped — no output file"
  exit 0
fi

KEYS='.summary and .inline_comments'

# Case 1: raw output is already structured JSON
if jq -e "${KEYS}" review_raw.txt > /dev/null 2>&1; then
  cp review_raw.txt review_parsed.json
else
  # Case 2: streaming JSONL from opencode run --format json.
  grep '"type"\s*:\s*"text"' review_raw.txt \
    | tail -1 \
    | jq -r '.part.text // empty' \
    > review_text.txt 2>/dev/null || true

  if [[ -s review_text.txt ]]; then
    if python3 "${SCRIPT_DIR}/extract-review-json.py" \
      > review_extracted.json 2>/dev/null && \
      jq -e "${KEYS}" review_extracted.json > /dev/null 2>&1; then
      mv review_extracted.json review_parsed.json
      echo "::notice::Extracted structured JSON from JSONL"
    else
      jq -n --rawfile s review_text.txt \
        '{summary: $s, inline_comments: []}' \
        > review_output.json
      echo "review_json=review_output.json" >> "$GITHUB_OUTPUT"
      echo "review_mode=comment" >> "$GITHUB_OUTPUT"
      echo "::notice::Text output — comment fallback"
      exit 0
    fi
  else
    # Case 3: unrecognized format — dump raw as summary
    jq -n --rawfile s review_raw.txt \
      '{summary: $s, inline_comments: []}' \
      > review_output.json
    echo "review_json=review_output.json" >> "$GITHUB_OUTPUT"
    echo "review_mode=comment" >> "$GITHUB_OUTPUT"
    echo "::notice::Non-JSON output — comment fallback"
    exit 0
  fi
fi

# Filter inline comments to only diff-visible lines
if python3 "${SCRIPT_DIR}/filter-diff-lines.py" \
  "${DIFF_PATH}" review_parsed.json \
  > review_output.json 2>filter_stderr.txt; then
  cat filter_stderr.txt >&2
  COUNT=$(jq '.inline_comments | length' review_output.json)
  if [[ "${COUNT}" -gt 0 ]]; then
    echo "review_mode=inline" >> "$GITHUB_OUTPUT"
  else
    echo "review_mode=comment" >> "$GITHUB_OUTPUT"
    echo "::notice::All comments filtered (not in diff)"
  fi
else
  echo "::warning::Diff line filter failed, using unfiltered"
  cp review_parsed.json review_output.json
  echo "review_mode=inline" >> "$GITHUB_OUTPUT"
fi
echo "review_json=review_output.json" >> "$GITHUB_OUTPUT"
