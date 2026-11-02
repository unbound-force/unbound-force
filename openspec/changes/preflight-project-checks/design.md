## Context

Pre-flight builds its tool list only from tracked config
(`.pre-commit-config.yaml`, `Makefile`, `.github/workflows`).
Projects that need a check of their own, especially stealth
installs that cannot touch tracked files, have nowhere to declare
one. See the proposal for the motivating fullsend-ai bash
env-contract case and the constitution alignment (all six
principles PASS). This design records the decisions behind the
implementation as built.

## Goals / Non-Goals

### Goals
- A project-owned file where checks are declared, which survives
  `uf init`.
- Deterministic execution: no shell, literal argv, bounded
  output and time.
- The same file-scope and soft-gate semantics that detected
  tools get.
- A skill change limited to one tool call plus result handling
  (determinism pack DE-001/DE-003/DE-005/DE-006).

### Non-Goals
- Converting the rest of the pre-flight skill to tools.
- Mapping declared checks to CI checks, or skipping them in
  `ci-aware`.
- A new Divisor or any change to Divisor permissions.

## Decisions

### D1: Extend pre-flight rather than add a Divisor

Divisors have `bash: deny` and receive command evidence only
through `pre_flight_results`. A Divisor would therefore need
pre-flight anyway, and it would spend a review slot. Pre-flight
produces the evidence, and the existing Divisors interpret it.
(Constitution I: artifact-based handoff.)

### D2: A new project-owned file, not a new key in an existing file

`.uf/review-matrix.yaml` and `.uf/reviewer-capabilities.yaml`
configure reviewer selection. Execution policy is a separate
concern, and a separate file keeps each schema strict. The file
is classified exactly like the other review-config extension
points (seeded once, kept on re-run, reset by `--force`), and the
classification is pinned by the ownership table and a re-run
regression test.

### D3: `run` is a path, not a command line

A shell string would require escaping rules, would make
`requires` probing unreliable, and would open an injection
surface through YAML that a branch can modify. `run` is
therefore validated as a single repo-relative executable (no
absolute paths, no `..` segments, no metacharacters), `args` is
a literal list, and execution uses `execFile` with
`shell: false`. Projects that need pipelines wrap them in a
script. (Constitution V.)

### D4: One tool owns the whole chain

Validation, probing, scoping, execution, status mapping,
baseline classification and rendering are a fixed chain with no
judgment in it. Splitting the chain across prose steps would
make the model the integrator (DE-005). The tool returns
structured `checks` for reasoning and a `rendered` string for
the report, so the skill only inserts that string (DE-003). The
library takes every side effect as an injected dependency
(`readText`, `fileExists`, `isExecutable`, `parseYaml`,
`pathEnv`, `git`, `runProcess`, `makeTempDir`, `removeDir`), so
all routing is unit-testable. (Constitution IV and VI.)

### D5: The soft-gate baseline runs the branch's script against merge-base content

The prose draft re-ran checks in a merge-base worktree, which
fails for stealth installs: git-excluded `.uf/scripts/` do not
exist there, so every baseline would fail and every failure
would classify as `pre-existing` or `unknown` by accident. The
tool instead runs `<primary root>/<run>` with the worktree as
cwd and passes `UF_PREFLIGHT_PROJECT_ROOT=<primary root>`. That
holds the check logic constant and varies only the code under
test, which is the question causality asks.

The trade-off is accepted: if the branch changed the check
script itself, the baseline uses the new script. Scripts locate
their config through `UF_PREFLIGHT_PROJECT_ROOT`, and they
inspect the cwd (the worktree) when they need the code. A
missing merge-base or a failed `worktree add` yields `unknown`,
which is treated as branch-caused (fail closed). Cleanup runs in
a `finally` block: `git worktree remove --force` and then
directory removal.

### D6: Scope semantics match Phase 2a

`paths` uses the Phase 2a matching rules (a `dir/` prefix,
otherwise a suffix) and the same conservative fallbacks. Users
learn one model. A check with no `paths` is always in scope.
The diff is computed only when some check declares `paths`, or
lazily for the soft-gate merge-base.

### D7: Missing tools skip, invalid config fails

A missing binary or a non-executable `run` reports
`SKIP — not available`, the same as detected tools that are not
installed, so contributors without optional tools are not
blocked. A present but invalid `.uf/preflight.yaml` returns
`INVALID_PREFLIGHT_CONFIG` and pre-flight fails, because a typo
must not silently disable a gate. (Constitution V.)

### D8: Registration and staging

`preflight-declared.ts` lives in `.opencode/lib/` beside the
other shared modules. It is tool-owned, belongs to the Divisor
subset, and is staged with the review plugin sources, so the
plugin import resolves in every deploy mode. The YAML parser is
created eagerly when the tool registers so that the golden test
harness's Bun shim applies.

## Risks / Trade-offs

- **Script trust.** Declared checks run project code during
  review. This is the same trust level as the `Makefile` targets
  pre-flight already runs, and the file is project-owned and
  visible in review. A branch can change `.uf/preflight.yaml`
  and the scripts it names, just as it can change a `Makefile`.
- **Runtime.** Each check may take up to `timeout_seconds`
  (default 600, maximum 3600), and soft-gate re-runs failing
  checks once more. Projects control the cost through `paths`
  and `timeout_seconds`.
- **Skill growth.** +38 lines (551 -> 589). The rest of the
  pre-flight skill is still prose; migrating it is a separate
  change.
- **Stealth activation.** Existing stealth installs keep the old
  skill until they re-run `uf init` with a release that contains
  this change. Until then a seeded `.uf/preflight.yaml` has no
  effect.
