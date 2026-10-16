## Context

`@opencode-ai/plugin` 1.4.10's `tool()` helper accepts exactly one
argument contract: an `args` key holding a zod raw shape, with
`execute` receiving `z.infer<z.ZodObject<Args>>`. Eleven tool
factories in `uf-workflow` (8) and `uf-finale` (3) instead pass a
`parameters` key holding plain `{ type, description }` objects.
TypeScript cannot flag this — the SDK generic accepts the extra
key structurally — and at runtime the SDK ignores it, leaving
`args` undefined. OpenCode then exposes an empty JSON schema to
the model and strips all arguments before `execute` runs.

The defect was confirmed live during the PR #733 finale
(`uf_finale_ci_poll` returning `INVALID_PR_NUMBER` despite an
explicit `prNumber`, `gh`/bash fallbacks for PR creation and
git push across two consecutive finales). Unit tests never
caught it because `invokeTool` calls `execute` directly with
native JS types, bypassing the SDK schema path.

The correct pattern is already proven in this repo:
`review-dispatch` (7 tool sites), `invoke-agent`, and
`createPersistUnleashSummaryTool` in uf-workflow all declare
zod `args` and expose full schemas at runtime.

Constitution alignment (from the proposal): all six principles
PASS — see `proposal.md` "Constitution Alignment". Decisions D4
and D5 below carry the Security by Default and Testability
weight; D1 carries Deterministic Execution.

## Goals / Non-Goals

### Goals

- Every one of the eleven affected factories exposes its
  parameters through the SDK-recognized `args` zod contract,
  so the model-visible schema is populated and runtime
  arguments reach `execute`
- The conversion is mechanical and behavior-preserving:
  names, types, optionality, and descriptions survive exactly;
  no `execute` body, envelope, or error code changes
- A schema-exposure regression test guards every affected
  factory so the declaration contract cannot silently regress
  again
- Scaffold asset mirrors stay byte-identical with the
  canonical plugin files

### Non-Goals

- No upgrade of `@opencode-ai/plugin` or `zod` (pinned
  dependency versions are gate-protected; `args` is the SDK's
  intended interface, not a workaround)
- No refactor of `execute` bodies, validation, or error
  envelopes (including the shell-injection guards, which stay
  as defense-in-depth)
- No changes to `review-dispatch`, `invoke-agent`, or
  `persist_unleash_summary` declarations (control group,
  already correct)
- No abstraction layer (shared helper/wrapper) over the zod
  declarations
- No coverage-threshold, CI-flag, or agent-setting changes

## Decisions

### D1: Mechanical in-place conversion, not an abstraction

Rename `parameters:` to `args:` at each of the eleven sites and
convert each plain object to the matching zod expression
(`z.string().describe(...)`, `z.boolean()`, `z.number()`),
matching the style already used by
`createPersistUnleashSummaryTool` and review-dispatch. With 11
small sites and no shared semantics beyond the pattern itself,
a shared helper would add indirection without reducing real
duplication (Deterministic Execution: the conversion itself is
executed as exact edits, not re-derived per site).

### D2: Optionality via `.optional()`, never `.default()`

Parameters whose legacy `execute` type is optional (`data`,
`key`, `branch` on fetch/push, `setUpstream`, `targetRepo`,
`timeoutMinutes`) map to `.optional()` zod fields. Required
parameters (`command`, `step`, `rawModelId`, `remote`, `ref`,
`title`, `body`, `prNumber`) map to required fields. Defaults
stay in `execute` (e.g. `timeoutMinutes ?? 15`) so the runtime
semantics of direct `execute` callers — including the existing
test suite — are bit-for-bit unchanged.

### D3: `execute` bodies and inline annotations untouched

The existing inline `execute: async (args: { ... })`
annotations remain valid: the zod-inferred types are assignable
to them, and any drift would surface as a `tsc`/vitest failure
rather than a silent break. Keeping the bodies untouched
minimizes the diff and keeps the migration reviewable as pure
declaration repair.

### D4: Runtime validation retained as defense-in-depth

Schema validation only guards the live SDK invocation path;
direct `execute` callers (tests, internal consumers) bypass it.
The `INVALID_PR_NUMBER` guard, the `targetRepo` regex, and the
`INVALID_UNLEASH_SUMMARY` content checks therefore stay exactly
where they are (Security by Default: input validation before
security-sensitive operations remains structural, not
review-time).

### D5: Schema-exposure tests assert `Object.keys(tool.args)`

Each affected factory gets an assertion in its existing test
file (`uf-workflow.test.ts`, `uf-finale.test.ts`) that the
factory's returned definition has `args` keys equal to the
sorted expected parameter list — the pattern proven in
`invoke-agent.test.ts` ("exposes invoke_agent as a
provider-free tool definition") and `dispatch-agent-run.test.ts`.
The assertion fails both when `args` is left undefined
(regression to `parameters`) and when a parameter is dropped or
renamed. The already-correct `persist_unleash_summary`
declaration is included as a future-regression guard. No
opencode runtime or network is required (Testability).

### D6: Scaffold parity via byte-identical mirror

The canonical plugin edits are copied verbatim into
`internal/scaffold/assets/opencode/plugins/uf-workflow/index.ts`
and `.../uf-finale/index.ts`. The existing drift tests
(`internal/scaffold/scaffold_test.go`) enforce byte equality;
`uf-finale`'s new zod import changes no package manifests
because `zod` 4.1.8 is already a runtime dependency declared in
`.opencode/package.json` (Composability First: no new
dependency, plugins remain standalone).

### D7: Documentation gate satisfied via CHANGELOG plus issue

`CHANGELOG.md` gains an entry; because the fix restores
user-visible workflow behavior (model-invoked git push, CI
polling, PR creation, attribution), a documentation issue is
filed in this repository and its number recorded at task
check-off rather than claimed without provenance (Observable
Quality).

## Risks / Trade-offs

- **Type mismatch between zod inference and inline annotations**
  (`.optional()` yields `T | undefined`): if any annotation
  disagrees, `tsc`/vitest fails loudly. Acceptable — that is
  the desired failure mode. Verified compatible for all
  eleven sites' current annotations.
- **Live-only behavior change**: the model can now pass
  parameters, and invalid types are rejected by schema
  validation before `execute`. There is no test harness that
  exercises the live SDK path; the schema-exposure tests
  assert the declaration contract the SDK consumes, which is
  the deepest contract reachable without embedding opencode.
  Accepted.
- **Mirror drift risk during implementation**: canonical and
  scaffold copies must be updated in the same task step, or
  the Go drift tests fail. This is an existing, enforced
  invariant — failing fast is the designed behavior.
- **Test-count growth**: eleven-plus assertions add vitest
  weight, negligible against the existing suite.
