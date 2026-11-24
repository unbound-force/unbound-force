#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
#
# Learnings-Only Change Detection (CI-FR-001)
#
# Decides whether a pull request changes anything besides Dewey
# learnings. Shared by the `changes` job in ci_local.yml and
# ci_checks.yml so both workflows classify a PR the same way.
#
# Usage: bash scripts/detect-learnings-only.sh
#   Run from the repository root with full history.
#   Reads GITHUB_EVENT_NAME, BASE_SHA, and HEAD_SHA.
#
# Prints `true` (code present, run full CI) or `false` (only
# regular .md files under .uf/dewey/learnings/ changed) on
# stdout. Every changed path is logged to stderr. Anything
# indeterminate prints `true`, so a bug costs CI minutes, never
# gate coverage.

set -uo pipefail

code='true'
event="${GITHUB_EVENT_NAME:-}"
base="${BASE_SHA:-}"
head="${HEAD_SHA:-}"

if [ "$event" = 'pull_request' ] && [ -n "$base" ] && [ -n "$head" ]; then
  diffout="$(mktemp)"
  # --no-renames: a rename is a deletion plus an addition, and
  # both paths are classified. -z: paths are NUL-delimited and
  # never quoted, so crafted filenames cannot break parsing.
  if git diff -z --no-renames --raw "$base...$head" >"$diffout" && [ -s "$diffout" ]; then
    code='false'
    # Each entry is ":oldmode newmode oldsha newsha status" NUL path NUL.
    while IFS= read -r -d '' meta && IFS= read -r -d '' path; do
      newmode="${meta#* }"
      newmode="${newmode%% *}"
      status="${meta##* }"
      printf 'changed: %s\n' "$path" >&2
      case "$path" in
        .uf/dewey/learnings/*.md) ;;
        *) code='true' ;;
      esac
      # Only regular files qualify. Symlinks (120000),
      # submodules (160000), and deletions (000000) count as code.
      case "$newmode" in
        100644|100755) ;;
        *) code='true' ;;
      esac
      # Added, modified, or type-changed entries only. With
      # --no-renames git never reports R or C, and D is already
      # code through its 000000 mode. Anything else fails safe.
      case "$status" in
        A|M|T) ;;
        *) code='true' ;;
      esac
    done <"$diffout"
  fi
  rm -f "$diffout"
fi

printf '%s\n' "$code"
