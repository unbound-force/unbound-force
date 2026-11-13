#!/usr/bin/env bash
# Snapshot this repo's working tree into $EVAL_WORKDIR for one eval run.
#
# Copies tracked and untracked (non-ignored) files, so uncommitted edits
# to commands, agents, and skills are what gets evaluated. .eval/ is left
# out: the agent under test must not be able to read its own assertions.
# The result is a fresh one-commit repo on a feature branch, so
# `files_unchanged` assertions start from a clean tree.
#
# Extra arguments are run as a command inside the workdir afterwards,
# letting a fixture adjust state (e.g. `git checkout --detach`).
set -euo pipefail

lib="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
src="${EVAL_REPO_ROOT:?}"
dst="${EVAL_WORKDIR:?}"

cd "$src"
git ls-files -z --cached --others --exclude-standard \
  | while IFS= read -r -d '' f; do
      case "$f" in .eval/*) continue ;; esac
      # Skip tracked files deleted in the working tree
      if [ -e "$f" ] || [ -L "$f" ]; then printf '%s\0' "$f"; fi
    done \
  | tar --null -T - -cf - \
  | tar -xf - -C "$dst"

# Plugin dependencies are gitignored; link rather than reinstall.
if [ -d "$src/.opencode/node_modules" ]; then
  ln -s "$src/.opencode/node_modules" "$dst/.opencode/node_modules"
fi

cd "$dst"
git init -q -b main
git add -A
git -c user.name=eval -c user.email=eval@example.invalid \
  commit -q --no-verify -m "eval fixture snapshot"

bash "$lib/write-run-config.sh"

if [ "$#" -gt 0 ]; then
  "$@"
fi
