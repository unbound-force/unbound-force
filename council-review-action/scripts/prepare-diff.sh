#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
#
# Prepare the PR diff for review by filtering out noise files
# (lock files, vendored deps, generated code). The filtered diff
# is written to pr-diff-filtered.patch.
#
# NOTE: Spec/doc files (openspec/, .specify/, docs/) are NOT excluded.
# Divisor personas specifically review specs for intent drift,
# constitution alignment, and completeness.
#
# Required env: DIFF_PATH
set -euo pipefail

EXCLUDE_PATTERNS=(
  # Lock files
  'go\.sum'
  'package-lock\.json'
  'yarn\.lock'
  'pnpm-lock\.yaml'
  'Cargo\.lock'
  'Gemfile\.lock'
  'poetry\.lock'
  'composer\.lock'
  # Vendored / third-party
  'vendor/'
  'node_modules/'
  'third_party/'
  # Generated code
  '\.pb\.go'
  '\.gen\.go'
  '_generated\.'
  '\.min\.js'
  '\.min\.css'
  '\.snap$'
  # Test fixtures
  'testdata/'
  'fixtures/'
)

build_regex() {
  local regex=""
  for pat in "${EXCLUDE_PATTERNS[@]}"; do
    if [[ -n "${regex}" ]]; then
      regex="${regex}|"
    fi
    regex="${regex}${pat}"
  done
  echo "${regex}"
}

EXCLUDE_REGEX=$(build_regex)

RAW_LINES=$(wc -l < "${DIFF_PATH}" | tr -d ' ')

filtered_diff() {
  local skip=false
  while IFS= read -r line; do
    if [[ "${line}" == "diff --git"* ]]; then
      skip=false
      if echo "${line}" | grep -qE "${EXCLUDE_REGEX}"; then
        skip=true
      fi
    fi
    if [[ "${skip}" == false ]]; then
      echo "${line}"
    fi
  done < "${DIFF_PATH}"
}

FILTERED=$(filtered_diff)
FILTERED_LINES=$(echo "${FILTERED}" | wc -l | tr -d ' ')
NOISE_LINES=$((RAW_LINES - FILTERED_LINES))

if [[ "${NOISE_LINES}" -gt 0 ]]; then
  echo "::notice::Excluded ${NOISE_LINES} noise lines (lock, vendor, generated). Raw: ${RAW_LINES}, filtered: ${FILTERED_LINES}"
fi

echo "${FILTERED}" > pr-diff-filtered.patch

# Produce a line-annotated version for the LLM prompt.
# Each '+' or context line gets a [L<N>] prefix showing its
# real source-file line number, so the model reads it directly
# rather than trying to count lines across the whole patch.
annotate_lines() {
  local new_line=0
  while IFS= read -r line; do
    if [[ "${line}" == "diff --git"* ]] || \
       [[ "${line}" == "---"* ]] || \
       [[ "${line}" == "+++"* ]] || \
       [[ "${line}" == "index "* ]] || \
       [[ "${line}" == "new file"* ]] || \
       [[ "${line}" == "deleted file"* ]] || \
       [[ "${line}" == "old mode"* ]] || \
       [[ "${line}" == "new mode"* ]] || \
       [[ "${line}" == "similarity index"* ]] || \
       [[ "${line}" == "rename from"* ]] || \
       [[ "${line}" == "rename to"* ]] || \
       [[ "${line}" == "Binary files"* ]]; then
      echo "${line}"
      continue
    fi

    # Hunk header: extract new-file start line
    if [[ "${line}" == @@* ]]; then
      new_line=$(echo "${line}" \
        | sed -n 's/^@@ -[0-9,]* +\([0-9]*\).*/\1/p')
      echo "${line}"
      continue
    fi

    if [[ "${line}" == +* ]] && [[ "${line}" != "+++"* ]]; then
      printf '[L%d] %s\n' "${new_line}" "${line}"
      new_line=$((new_line + 1))
    elif [[ "${line}" == -* ]] && [[ "${line}" != "---"* ]]; then
      echo "${line}"
    elif [[ "${line}" == \\* ]]; then
      echo "${line}"
    else
      # Context line
      printf '[L%d] %s\n' "${new_line}" "${line}"
      new_line=$((new_line + 1))
    fi
  done < pr-diff-filtered.patch
}

annotate_lines > pr-diff-annotated.patch
