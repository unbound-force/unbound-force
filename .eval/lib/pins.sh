# shellcheck shell=bash disable=SC2034
# Pinned versions of everything the eval runtime loads.
#
# Sourced by prepare-eval-env.sh and run.sh, and by the nightly workflow
# (ci_eval.yml), so local runs and CI use the same pins. Every commit is
# a full SHA. Changing a pin is a reviewed change: it changes what the
# evals test.

# opencode itself. run.sh refuses to run any other version.
EVAL_OPENCODE_VERSION=1.19.0

# Plugins, built once per commit into $EVAL_CACHE/plugins.
EVAL_DCP_URL=https://github.com/jflowers/opencode-dynamic-context-pruning.git
EVAL_DCP_SHA=79174fc74418be390536b8de887c5327d6ba01e4
EVAL_AQ_URL=https://github.com/jflowers/opencode-auto-question.git
EVAL_AQ_SHA=6eb86836928b421a6863d8f2f1ef1f303141e3f6

# Eval driver (unbound-force/eval-infra).
EVAL_DRIVER_URL=https://github.com/unbound-force/eval-infra.git
EVAL_DRIVER_SHA=8a31ef09d0dcce900f556103b3f412ef2f91f81f

# MCP servers declared in opencode.json. CI installs these Linux
# release builds; locally the installed binaries are used.
EVAL_DEWEY_VERSION=3.5.3
EVAL_REPLICATOR_VERSION=0.7.0
