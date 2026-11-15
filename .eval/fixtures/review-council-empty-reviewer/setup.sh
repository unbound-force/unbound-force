#!/usr/bin/env bash
# Issue #1052 reproduction: a council whose reviewers return no final
# text.
#
# Builds the same clean mini-repo as review-council-clean-change, then
# pins every review-capable persona's run to a provider entry whose
# 64-token output limit forces finish reason "length" before any text —
# the empty-reviewer shape from issue #1052 (reasoning exhausts the
# budget, tokens.output stays 0). Pre-fix, each empty reviewer recorded
# success with an APPROVE vote and the council approved a change no
# reviewer reviewed; post-fix each is a retryable child_empty_response
# failure and the council lands on the zero-success non-approving path.
set -euo pipefail

fixture="${EVAL_FIXTURE_DIR:?}"

bash "$fixture/../../lib/setup-repo-copy.sh" \
  bash "$fixture/../../lib/make-mini-repo.sh" feature-clean

cd "${EVAL_WORKDIR:?}"

# The config patches below are committed onto the base branch (main)
# and the feature branch is rebased onto them. capture_review_diff
# diffs the working tree, uncommitted changes included, against the
# base ref — leaving the patches uncommitted would leak the reviewer
# pinning into the reviewed diff, so the council would no longer be
# reviewing the same clean change review-council-clean-change does.
# On the base they shape the run (the plugins read them from the
# working tree) without entering the diff, and the tree stays clean.
branch="$(git branch --show-current)"
git checkout -q main

# Patch the project the command runs in. Anchors fail loudly when
# upstream moves them, rather than silently producing a run that no
# longer reproduces the empty-reviewer shape (the same rule the
# mini-repo base hook follows). python3 -I keeps the edits stdlib-only.
python3 -I - <<'PY'
import json
import pathlib
import sys

# 1. Declare the tiny-output provider. The model id must be the real
#    run model (the provider answers 404 for invented ids); a separate
#    provider entry carries the 64-token output cap that ends every
#    generation in reasoning with no text. The eval's XDG config keeps
#    declaring the normal provider the host session uses; project
#    config adds this one, and only the pinned reviewer runs select it.
config = pathlib.Path("opencode.json")
data = json.loads(config.read_text())
data.setdefault("provider", {})["enmaas-empty"] = {
    "npm": "@ai-sdk/openai-compatible",
    "name": "EnMaaS empty",
    "options": {
        "baseURL": "https://api.enmaas.devshift.net/v1",
        "apiKey": "{env:ENMAAS_KEY}",
    },
    "models": {
        "rits/zai-org/glm-5-3": {
            "name": "GLM 5.3 (64-token output cap)",
            "limit": {"context": 262144, "output": 64},
        }
    },
}
config.write_text(json.dumps(data, indent=2) + "\n")

# 2. Pin every review-capable persona's code-mode run to that model.
#    The reviewer-capabilities manifest is lock-validated against the
#    canonical enrollment, so the panel cannot be reduced; pinning all
#    review personas makes the empty votes decisive whichever subset
#    the change's categories enroll.
matrix = pathlib.Path(".uf/review-matrix.yaml")
text = matrix.read_text()
anchor = "runs: {}"
if anchor not in text:
    sys.exit("review matrix no longer has the empty runs anchor")
tiny = "enmaas-empty/rits/zai-org/glm-5-3"
agents = [
    "divisor-adversary",
    "divisor-architect",
    "divisor-curator",
    "divisor-guard",
    "divisor-sre",
    "divisor-testing",
    "divisor-linguist",
    "divisor-determinism",
]
runs = "".join(f"    {agent}:\n      - model: {tiny}\n" for agent in agents)
matrix.write_text(text.replace(anchor, "runs:\n  code:\n" + runs, 1))
PY

# Land the matrix patch on the base, then replay the feature commits
# on top. The files are disjoint from the feature (config vs code), so
# the rebase cannot conflict; the -c identity matches the one the
# mini-repo scripts and the eval CI workflow already use. opencode.json
# is gitignored at the repo root, so the provider entry stays an
# untracked workdir file — read by the run, invisible to every diff,
# like the .opencode/opencode.json scaffolding.
git add .uf/review-matrix.yaml
git -c user.name=eval -c user.email=eval@example.invalid \
  commit -q -m "chore(eval): pin council reviewers to a 64-token provider"
git checkout -q "$branch"
git -c user.name=eval -c user.email=eval@example.invalid rebase -q main
