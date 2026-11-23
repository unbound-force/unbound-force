---
tag: command-evals
author: jay-flowers
category: gotcha
created_at: 2026-11-14T22:56:40Z
identity: command-evals-20261114T225640-jay-flowers
tier: draft
---

In unbound-force, the .eval/ command-evals assertion evaluator is NOT in this repo: the assertions.yaml vocabulary (tool_called, tool_output_matches, tool_output_not_matches, stub_not_called, no_permission_asks, no_unanswered_questions, sessions_completed, files_unchanged) is implemented by the externally pinned eval-infra driver (unbound-force/eval-infra at EVAL_DRIVER_SHA in .eval/lib/pins.sh, exec'd by .eval/run.sh). .eval/lib/ holds only setup shell scripts (make-mini-repo.sh, pins.sh, prepare-eval-env.sh, setup-repo-copy.sh, write-run-config.sh, stubs/gh, guidance/, mini-repo/). A spec task that extends eval assertions must state whether the work is fixture data (existing assertion vocabulary, in-repo, e.g. tool_output_not_matches on consolidate_dispatch output) or driver code (cross-repo change plus a pin update, which pins.sh marks as a reviewed change); naming a new assertion form without an existing owner strands the implementer at .eval/lib/ looking for a harness that is not there.
