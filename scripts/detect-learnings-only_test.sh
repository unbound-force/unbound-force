#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
#
# Tests for scripts/detect-learnings-only.sh (CI-FR-001).
#
# Builds a throwaway git repository per scenario and checks the
# detector's verdict. The `changes` job runs this before the
# detector, so a broken detector fails the job (and both
# aggregates) instead of silently skipping CI.
#
# Usage: bash scripts/detect-learnings-only_test.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DETECT="$SCRIPT_DIR/detect-learnings-only.sh"
L='.uf/dewey/learnings'

pass=0
fail=0
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

# new_repo: fresh repo with one code file and one learning on
# main; leaves BASE set to that commit and checks out a branch.
new_repo() {
  repo="$work/repo$((pass + fail))"
  git init -q -b main "$repo"
  cd "$repo"
  git config user.email ci@example.invalid
  git config user.name ci
  git config commit.gpgsign false
  mkdir -p "$L" src
  printf 'package main\n' >src/main.go
  printf '# existing\n' >"$L/existing.md"
  git add -A
  git commit -q -m base
  BASE="$(git rev-parse HEAD)"
  git checkout -q -b pr
}

commit_pr() {
  git add -A
  git commit -q -m change --allow-empty
  HEAD_SHA="$(git rev-parse HEAD)"
}

# expect NAME WANT [EVENT] [BASE] [HEAD]
expect() {
  local name="$1" want="$2" event="${3-pull_request}"
  local base="${4-$BASE}" head="${5-$HEAD_SHA}"
  local got
  got="$(GITHUB_EVENT_NAME="$event" BASE_SHA="$base" HEAD_SHA="$head" \
    bash "$DETECT" 2>/dev/null)"
  if [ "$got" = "$want" ]; then
    pass=$((pass + 1))
    printf 'ok   %s\n' "$name"
  else
    fail=$((fail + 1))
    printf 'FAIL %s: want %s, got %s\n' "$name" "$want" "$got"
  fi
  cd "$work"
}

new_repo; printf '# new\n' >"$L/new.md"; commit_pr
expect 'added learning' false

new_repo; printf '# more\n' >>"$L/existing.md"; commit_pr
expect 'modified learning' false

new_repo; mkdir -p "$L/topic"; printf '# nested\n' >"$L/topic/n.md"; commit_pr
expect 'nested learning' false

new_repo; printf '# x\n' >"$L/x.md"; chmod +x "$L/x.md"; commit_pr
expect 'executable learning' false

new_repo; printf '# new\n' >"$L/new.md"; printf '// x\n' >>src/main.go; commit_pr
expect 'learning plus code' true

new_repo; printf '// x\n' >>src/main.go; commit_pr
expect 'code only' true

new_repo; git mv src/main.go "$L/auth.md"; commit_pr
expect 'code renamed into learnings' true

new_repo; git mv "$L/existing.md" "$L/renamed.md"; commit_pr
expect 'rename within learnings' false

new_repo; git rm -q "$L/existing.md"; commit_pr
expect 'deleted learning' false

new_repo; git rm -q src/main.go; commit_pr
expect 'deleted code file' true

new_repo; git rm -q "$L/existing.md"; mkdir -p "$L"; ln -s ../../../src/main.go "$L/existing.md"; commit_pr
expect 'learning replaced by symlink' true

new_repo; git update-index --add --cacheinfo "160000,$BASE,$L/sub.md"; commit_pr
expect 'submodule under prefix' true

new_repo; printf '# x\n' >"$L/.md"; commit_pr
expect 'bare .md name' true

new_repo; ln -s ../../../src/main.go "$L/link.md"; commit_pr
expect 'symlink under prefix' true

new_repo; printf 'x\n' >"$L/notes.txt"; commit_pr
expect 'non-markdown under prefix' true

new_repo; printf 'x\n' >"$L/README.MD"; commit_pr
expect 'uppercase extension' true

new_repo; printf '# x\n' >".uf/dewey/learnings.md"; commit_pr
expect 'prefix lookalike' true

new_repo; printf '# x\n' >"$L/a
b.md"; commit_pr
expect 'newline in learning name' false

new_repo; mkdir -p "src/evil
$L"; printf 'x\n' >"src/evil
$L/x.md"; commit_pr
expect 'newline smuggling a learnings path' true

new_repo; printf '# x\n' >"$L/\$(touch pwned) \"q\" ;x.md"; commit_pr
expect 'shell metacharacters in learning name' false
[ ! -e "$repo/pwned" ] || { fail=$((fail + 1)); echo 'FAIL metacharacters ran a command'; }

new_repo; commit_pr
expect 'empty diff' true

new_repo; printf '# new\n' >"$L/new.md"; commit_pr
expect 'unknown base' true pull_request 0000000000000000000000000000000000000000

new_repo; printf '# new\n' >"$L/new.md"; commit_pr
expect 'missing base' true pull_request ''

new_repo; printf '# new\n' >"$L/new.md"; commit_pr
expect 'push event' true push

printf '\n%d passed, %d failed\n' "$pass" "$fail"
[ "$fail" -eq 0 ]
