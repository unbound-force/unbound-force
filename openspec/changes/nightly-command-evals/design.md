## Context

The prototype harness on `eval/review-council-arg-rejection` runs
five `/uf.review-council` fixtures through opencode with uf's real
plugins and checks deterministic assertions on the trajectory
(proposal.md, Why). It runs only on a maintainer's machine: it
expects `EVAL_DRIVER` to point at a local eval-infra checkout and
reads `ENMAAS_KEY` from chezmoi's `env.sh`. Constitution alignment
is assessed in proposal.md; Observable Quality (III) and
Deterministic Execution (VI) shape the reporting and pinning below.

Measured on GLM 5.3 through EnMaaS, one run per fixture plus three
repeats of each council fixture:

| Fixture | Time | Input tokens |
|---|---|---|
| Argument rejection (each) | 40 to 79 s | about 25k |
| Clean change | 23 to 35 min | 542k to 723k |
| Planted defects | 43 to 89 min | 0.98M to 1.56M |

## Goals / Non-Goals

### Goals

- Every fixture runs every night on main, unattended.
- A maintainer runs the same fixtures locally with one command and
  the same pins.
- Each run leaves enough evidence (summary, trajectory, run log,
  `runs.db`) to tell a regression from model variance.

### Non-Goals

- Gating PRs or releases on eval results. That needs a measured
  pass rate and #1052 fixed first.
- Deploying `.eval/` to downstream repos through `uf init`.
- Using eval-infra's reusable workflow, which still targets Claude
  on Vertex and an old opencode version.
- Judge-model grading.

## Decisions

### D1. Nightly and on demand only

The workflow triggers on `schedule` and `workflow_dispatch`. A
council fixture takes up to 90 minutes, too slow for every PR, and
`pull_request` from forks would have no secret anyway. Running on
upstream main only also keeps the key out of fork contexts
(EV-FR-005, EV-FR-007).

### D2. One job per fixture

The driver always serves opencode on port 4096, so fixtures can't
share a runner in parallel. A matrix job per fixture directory runs
them in parallel on separate runners. The longest job (about 90
minutes) sets the wall time instead of the sum of all five. Each job
gets a 120-minute timeout.

### D3. `ENMAAS_KEY`, not `OPENCODE_CONFIG_CONTENT`

The org's `OPENCODE_CONFIG_CONTENT` secret carries a whole opencode
config. The eval runtime needs the key itself for two reasons:
auto-question calls EnMaaS directly using `{env:ENMAAS_KEY}`, and an
inline opencode config takes precedence over project config, so any
`permission` block in that secret could loosen the eval's
`external_directory` rules (EV-FR-002). A plain `ENMAAS_KEY` secret
keeps CI identical to local runs.

### D4. The driver is pinned like the plugins

`prepare-eval-env.sh` already clones DCP and auto-question by commit
into `~/.cache/uf-eval/plugins` and builds them once. The eval-infra
driver joins that list (with `bun install`), and `run.sh` uses it
when `EVAL_DRIVER` is unset. One file then holds every pin, and CI
caches the directory keyed on its contents. Every build step returns
an error explicitly, because callers run it in a command
substitution where `set -e` doesn't apply (EV-FR-001).

### D5. Failures stay red

Until #1052 is fixed, `sessions_completed` fails about half of
council runs. Those jobs stay failing rather than
`continue-on-error`: nothing gates on the workflow (D1), and a
silently green job would hide exactly what the evals exist to find.

### D6. Runtime toolchain in CI

Each job installs what the local runs relied on: Node and Bun (DCP
build, driver, `.opencode` plugin dependencies through `npm ci`),
Go (reviewers run `go test` in the mini repo), and opencode-ai at
the version pinned in `.eval/` (1.19.0). Python 3 and git come with
the runner image.

## Open Questions

### OQ1. Dewey and Replicator MCP servers

The tracked `opencode.json` declares `dewey serve` and
`replicator serve` as MCP servers. Both ran in every local eval, and
`/uf.review-council` stores its lessons through Dewey. A bare runner
has neither binary, and Dewey's semantic search also expects
Ollama. Options:

- **Install both** from pinned releases, accepting degraded Dewey
  search without Ollama. This is closest to how uf really runs.
  (Recommended)
- **Disable both** in the per-run config. This is simpler but
  changes what's tested.

Task 1.1 settles this before the workflow is written. Either way,
the first nightly run is checked by hand before its results are
trusted (task 5.2).

## Coverage Strategy

- **Harness scripts**: `make eval-fast` exercises
  `prepare-eval-env.sh`, `setup-repo-copy.sh`, `write-run-config.sh`
  and `run.sh` end to end in about three minutes.
- **Startup checks**: a deliberate run with
  `--require-tools no_such_tool` confirms the driver fails at
  startup (EV-FR-001), as it did in the prototype.
- **Assertions**: unit-tested in eval-infra (66 tests, including
  `sessions_completed`). Replaying the new assertion over the ten
  recorded council runs flagged exactly the five with an empty
  reviewer.
- **Workflow**: a manual `workflow_dispatch` run on the branch
  before merge, plus the first two nightly runs reviewed by hand.
- **Coverage gates**: no Go or plugin code changes, so neither
  coverage floor moves.

## Risks / Trade-offs

- **Model variance**: the fix loop makes time and tokens vary by
  about 2x. Trends need several nights.
- **EnMaaS availability**: an outage fails every job that night.
  The summary must make that distinguishable from a regression (the
  run log shows the provider error).
- **Fixture upkeep**: changes to repo rules (constitution
  principles, MP numbering, coverage-gate format) can break the
  mini-repo fixtures. The fixtures fail loudly, and AGENTS.md says
  where to fix them.
