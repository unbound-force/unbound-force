# ADR-003: Instruction-lint CI threshold and scaffold adoption

**Status:** Proposed
**Date:** 2026-11-08
**Context:** Issue #971 (follow-ups from PR #958). The instruction
linter (`.opencode/lib/instruction-lint.ts`, wrapper
`.uf/scripts/instruction-lint`) runs locally via
`make lint-instructions` and as an advisory pre-flight check
(`.uf/preflight.yaml`, `gate: advisory`). It is not a CI job, and
`uf init` does not scaffold it. Both open questions change a
quality gate or the scaffold contract, so under the AGENTS.md
Gatekeeping rule they need a maintainer decision. This ADR only
records the options and a recommendation. It changes nothing.

## Baseline (after the #971 remediation PR)

`node .opencode/lib/instruction-lint.ts --root .`: 0 errors,
16 warnings, 24 info (before: 1 error, 31 warnings, 24 info).

The 16 remaining warnings fall into two groups:

- 13 token-budget warnings (IL-030 file budget, IL-031 loaded
  budget) on the large orchestration commands (`uf.unleash`,
  `uf.review-council`, `uf.review-pr`, `uf.triage-issue`,
  `uf.fix-ci`, `uf.init`, `uf.address-feedback`). Clearing them
  means moving procedure out of these commands into tools or
  skills, which is its own refactor.
- 3 marker/body mismatches (IL-021) in `go.md` SC-006,
  `meta-prompt.md` MP-011 and `python.md` TC-010. Each rule mixes
  a [SHOULD]/[MAY] marker with an inner MUST clause. Fixing one
  changes either the rule's review severity or the force of its
  clause. Severity definitions are a Gatekeeping item.

## Question A: advisory or required in CI

| Option | Description | Cost | Risk |
|--------|-------------|------|------|
| A1 | Keep advisory (status quo): local make target plus the pre-flight advisory | None | Errors such as a dead `/command` reference (IL-010) can merge unnoticed |
| A2 | Required CI job, errors only: run the wrapper with no flags in the existing Node 22 `Plugin Tests` job of `ci_local.yml`. Exit 1 on errors, warnings stay informational | One CI step | Low. The baseline has 0 errors, and every error-severity rule is a mechanical fact (dead reference, broken frontmatter, unclosed fence) |
| A3 | Required CI job, `--strict` (warnings fail) | One CI step, plus clearing or suppressing the 16 warnings first | High churn. The token budgets are heuristics, so a hard gate would push contributors toward suppressions instead of real fixes |
| A4 | Required on changed files only (`--changed --base origin/main`), errors only | One CI step | Medium. A clean baseline already removes the problem this option addresses, and diff-scoped lint misses breakage in a file that did not change (deleting a command breaks references to it in other files) |

**Recommendation: A2.** Make errors a required CI step and keep
warnings advisory. Reconsider A3 for the IL-02x rules once the
IL-021 severity calls are made. Keep IL-030/031 advisory until the
large commands are split up. The pre-flight entry can then move
from `gate: advisory` to a blocking gate on the same terms
(errors only).

## Question B: scaffold the linter, or drop the mirror declaration

Current state: the scaffold mirror
`internal/scaffold/assets/opencode/package.json` ships no
`devDependencies`. It cannot ship them because of the npm 10.9.9
crash workaround (D2a). `yaml` 2.9.1 is declared only as a
devDependency in the canonical `.opencode/package.json`, and its
provenance is recorded in
`internal/scaffold/testdata/plugin-package-provenance.json`.
Neither the linter, its wrapper, nor the divisor-linguist persona
is scaffolded. So "drop the mirror yaml dependency" today only
means removing a provenance record.

| Option | Description | Cost | Risk |
|--------|-------------|------|------|
| B1 | Scaffold the linter, wrapper, make target and divisor-linguist | Embed 3 or more assets plus drift tests. `yaml` becomes a runtime dependency of scaffolded repos | `yaml` would have to move to `dependencies` in the mirror, which widens every downstream install. The linter's rules (pack ID formats, the RFC 2119 marker grammar, the builtin command list) are tuned to this repo's conventions and would need the configurable extension points that #971 also defers (BUILTIN_COMMANDS) |
| B2 | Keep repo-local and drop the provenance and mirror references to `yaml` that imply scaffolding | Small test and testdata edit | Low. Downstream repos get no linter |
| B3 | Keep repo-local now, and record a scaffold-readiness criterion: the linter runs dependency-free (frontmatter parsed without `yaml`, or behind a vendored parser) and its project conventions are configurable | Documentation only | Low. Adoption is deferred, not refused |

**Recommendation: B3, with B2's cleanup applied only if the
provenance record is found to cause churn.** Scaffold the linter
once it no longer needs a runtime dependency that scaffolded repos
would have to install, and once its repo-specific conventions are
configuration rather than code. Until then, shipping it would add
supply-chain surface to every downstream install and produce false
positives in repos that do not follow these pack conventions.

## Consequences

- If A2 is accepted, a follow-up PR adds one step to
  `ci_local.yml` and updates the `.uf/preflight.yaml` comment.
  That PR is a governance change and goes through spec workflow
  review.
- If B3 is accepted, #971's scaffold item closes and a tracking
  issue for the readiness criterion replaces it.
