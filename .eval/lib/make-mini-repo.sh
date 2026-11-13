#!/usr/bin/env bash
# Turn a setup-repo-copy.sh snapshot into a small controlled project and
# put one feature on its own branch for /uf.review-council to review.
#
# Usage (from a fixture's setup.sh):
#   setup-repo-copy.sh bash make-mini-repo.sh <variant>
# where <variant> names a directory under lib/mini-repo/.
#
# A variant is either a flat directory overlaid and committed as one
# commit, or a set of numbered step directories (1-spec, 2-impl, ...)
# applied and committed in order. In any overlay directory:
#   .commit   commit message (default: a generic feature message)
#   .hook.sh  run from the work tree after the overlay is copied, for
#             edits to files the snapshot owns (AGENTS.md, coverage-gate)
# A variant's .branch file names the feature branch (default
# feat/ref-lookup).
#
# The snapshot keeps everything the council reads as policy (agents,
# commands, plugins, .uf review config, packs, constitution, AGENTS.md,
# OpenSpec config) and drops the Go tree, so pre-flight runs a tiny CI
# instead of the repository's full `make check`, and the reviewed diff
# is only the planted feature.
set -euo pipefail

variant="${1:?variant directory under lib/mini-repo}"
lib="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
overlay="$lib/mini-repo/$variant"
[ -d "$overlay" ] || { echo "unknown variant: $variant" >&2; exit 2; }

cd "${EVAL_WORKDIR:?}"

commit() {
  git add -A
  git -c user.name=eval -c user.email=eval@example.invalid \
    commit -q --no-verify "$@"
}

# Copy one overlay directory into the work tree and run its hook.
apply() {
  rsync -a --exclude .commit --exclude .hook.sh --exclude .branch "$1/" ./
  if [ -f "$1/.hook.sh" ]; then
    bash "$1/.hook.sh"
  fi
}

message() {
  if [ -f "$1/.commit" ]; then cat "$1/.commit"; else echo "feat(refs): add Lookup to resolve git refs"; fi
}

keep=" .git .opencode .uf .specify .github .gitignore AGENTS.md LICENSE README.md opencode.json coverage-gate.json openspec "
for entry in .[!.]* *; do
  [ -e "$entry" ] || [ -L "$entry" ] || continue
  case "$keep" in *" $entry "*) ;; *) rm -rf "$entry" ;; esac
done
rm -rf .github/workflows
# Keep OpenSpec's schema and canonical specs; the repository's own
# in-flight changes are not part of the mini project.
rm -rf openspec/changes
apply "$lib/mini-repo/base"
commit --amend -m "eval fixture base"

# Sibling evidence is fetched clone-on-review; seed the clones from the
# source checkout so a run fetches deltas instead of cloning from scratch.
if [ -d "$EVAL_REPO_ROOT/.uf/sibling-clones" ]; then
  cp -R "$EVAL_REPO_ROOT/.uf/sibling-clones" .uf/sibling-clones
fi

# Reviewers use the canonical review matrix, whose empty profiles make
# every run inherit the parent session's model (the eval model). No
# override is written, so the run exercises the same host-fallback path
# a fresh checkout does.

branch=feat/ref-lookup
[ -f "$overlay/.branch" ] && branch="$(cat "$overlay/.branch")"
git checkout -q -b "$branch"

steps=()
for d in "$overlay"/[0-9]*-*/; do
  [ -d "$d" ] && steps+=("${d%/}")
done
if [ ${#steps[@]} -eq 0 ]; then
  steps=("$overlay")
fi
for step in "${steps[@]}"; do
  apply "$step"
  commit -m "$(message "$step")"
done
