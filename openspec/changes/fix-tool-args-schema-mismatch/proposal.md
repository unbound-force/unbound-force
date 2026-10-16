## Why

Eleven tool factories across the `uf-workflow` and `uf-finale`
plugins declare their parameters under a `parameters:` key using
plain `{ type, description }` objects. The `tool()` helper from
`@opencode-ai/plugin` 1.4.10 only recognizes an `args:` key
holding a zod raw shape (`z.ZodRawShape`). Because neither the
key nor the format matches, the SDK silently ignores the
declaration — no type error, no runtime warning — and opencode
exposes an empty JSON schema to the model. Worse, runtime
arguments are stripped before `execute` receives them, so this
is a hard break for model-side invocation, not just
undiscoverability (confirmed live during the PR #733 finale:
`uf_finale_ci_poll` returns `INVALID_PR_NUMBER` even when
`prNumber` is explicitly emitted in the call).

Observed consequences (two consecutive finales):

- `git_push` cannot receive `{ setUpstream, branch }` — plain
  `git push` fails with "no upstream branch", forcing manual
  bash workarounds in the `opsx/unleash-direct-review-dispatch`
  and `opsx/address-feedback-model-normalization` finales
- `uf_finale_create_pr` cannot receive `title`/`body` — PR #733
  was created via direct `gh pr create` with a temp body file
- `uf_finale_ci_poll` cannot be pointed at a PR by the model
- `uf_finale_commit_context` and `normalize_model_name` never
  receive `rawModelId` — commit attribution degrades to
  `unknown-model`, and the migrated `/uf.address-feedback`
  command (#709) instructs the agent to call
  `normalize_model_name`, whose parameter this bug hides

Unit tests stay green because they invoke `execute` directly
with native JS types (`.opencode/test/uf-workflow.test.ts`),
bypassing the SDK schema path entirely — parameter exposure is
never asserted. In-repo precedent proves the correct pattern:
`review-dispatch` (7 tool sites), `invoke-agent`, and
`persist_unleash_summary` all declare zod `args:` and expose
their full schemas to the model at runtime.

## What Changes

- Rename `parameters:` to `args:` in all 11 affected factories
  and convert every plain `{ type, description }` object to a
  zod shape (`z.string().describe(...)`, `z.boolean()`,
  `z.number()`), preserving each parameter's name, type,
  optionality, and description exactly
- Add the `zod` import to `uf-finale` (uf-workflow already
  imports it)
- Leave every `execute` body unchanged: runtime validation
  (e.g. the `INVALID_PR_NUMBER` shell-injection guard,
  `targetRepo` format check) stays as defense-in-depth behind
  the new schema-validation layer
- Add schema-exposure regression tests asserting each affected
  factory's `args` object contains exactly the expected keys
  (following the `invoke-agent.test.ts` "exposes ... tool
  definition" precedent)
- Mirror both edited plugin files byte-identically into the
  scaffold assets (`internal/scaffold/assets/opencode/plugins/`)
  so the embedded-asset drift tests pass
- Add a CHANGELOG entry and assess the documentation gate

## Capabilities

### New Capabilities

- `plugin-tool-schemas`: uf-workflow and uf-finale plugin tools
  MUST declare parameters via the SDK-recognized zod `args`
  contract so the model-visible schema and the runtime argument
  path both work

### Modified Capabilities

- None. Tool behavior (execute bodies, envelopes, error codes)
  is unchanged; only the declaration contract is repaired.

### Removed Capabilities

- None.

## Impact

**Affected files**:

- `.opencode/plugins/uf-workflow/index.ts` — 8 factories:
  `checkpoint`, `get_progress`, `clear_progress`,
  `normalize_model_name`, `git_fetch`, `git_push`, `git_merge`,
  `git_rebase`
- `.opencode/plugins/uf-finale/index.ts` — 3 factories:
  `uf_finale_commit_context`, `uf_finale_create_pr`,
  `uf_finale_ci_poll` (which declares both `prNumber` and
  `timeoutMinutes`)
- `internal/scaffold/assets/opencode/plugins/uf-workflow/index.ts`
  and `.../uf-finale/index.ts` — byte-identical mirrors
- `.opencode/test/uf-workflow.test.ts` and
  `.opencode/test/uf-finale.test.ts` — schema-exposure tests
- `CHANGELOG.md`

**Behavior change**: model-side invocations can now pass
parameters; live tool schemas are non-empty. Direct `execute`
callers (existing unit tests, `invokeTool` helper) are
unaffected because the zod-inferred argument types match the
existing inline annotations.

**Not changed**: no SDK upgrade (pinned dependency versions are
gate-protected), no execute logic, no envelope format, no
review-dispatch/invoke-agent tools (control group, already
correct), no coverage thresholds or CI flags.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

The model-tool contract becomes self-describing again: the
schema exposed to the model advertises each tool's parameters,
so the agent can collaborate with the tool through its declared
artifact (the JSON schema) instead of falling back to bash
workarounds that bypass the artifact contract. Tool result
envelopes (`success`/`failure` with codes) are unchanged.

### II. Composability First

**Assessment**: PASS

Both plugins remain independently installable and usable. No
new dependency is introduced — `zod` 4.1.8 is already a runtime
dependency of the `.opencode` plugin package, and uf-workflow
already imports it. The change touches only the two broken
plugins; review-dispatch and invoke-agent are untouched.

### III. Observable Quality

**Assessment**: PASS

The model-visible schema is the machine-parseable declaration
of each tool's input contract; after this change it is backed by
an automated, re-runnable regression test (schema-exposure
assertions) rather than trust. All existing output envelopes and
provenance metadata are unchanged.

### IV. Testability

**Assessment**: PASS

Schema-exposure tests assert the factory output's `args` keys
directly in vitest, with no external services, no network, and
no opencode runtime — closing the exact gap that let unit tests
stay green while the model-facing contract was broken. Existing
DI-based factory tests continue to verify execute behavior.

### V. Security by Default

**Assessment**: PASS

zod shapes add schema validation in front of every `execute`
call in live sessions. Existing runtime validation — the
`INVALID_PR_NUMBER` shell-injection guard, `targetRepo` format
regex, `INVALID_UNLEASH_SUMMARY` content checks — is retained
untouched as defense-in-depth. No dependencies are added or
repinned.

### VI. Deterministic Execution

**Assessment**: PASS

The deterministic tool paths (git push with upstream, CI poll by
PR number, PR creation with title/body, model-name attribution)
become actually invocable by the model, eliminating the manual
bash fallbacks that re-implemented tool behavior outside the
tested, tool-rendered path. The schema contract itself is now
guarded by tests.
