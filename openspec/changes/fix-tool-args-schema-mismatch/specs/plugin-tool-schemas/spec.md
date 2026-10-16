## ADDED Requirements

### Requirement: FR-001 SDK-conformant argument declaration

Every parameterized tool factory in `uf-workflow` and
`uf-finale` MUST declare its parameters under the `args` key
holding a zod raw shape (`z.ZodRawShape`), the only argument
contract recognized by `@opencode-ai/plugin`'s `tool()` helper.
After this change, no tool factory in either plugin MUST declare
a `parameters` key, and `uf-finale` MUST import `z` from `zod`
(`uf-workflow` already does). The eleven affected factories are:

- uf-workflow: `createCheckpointTool`, `createGetProgressTool`,
  `createClearProgressTool`, `createNormalizeModelNameTool`,
  `createGitFetchTool`, `createGitPushTool`, `createGitMergeTool`,
  `createGitRebaseTool`
- uf-finale: `createUfFinaleCommitContextTool`,
  `createUfFinaleCreatePrTool`, `createUfFinaleCiPollTool`

#### Scenario: No legacy declaration keys remain

- **GIVEN** the change is implemented
- **WHEN** `.opencode/plugins/uf-workflow/index.ts` and
  `.opencode/plugins/uf-finale/index.ts` are searched for the
  string `parameters:`
- **THEN** no match is found in either file
- **AND** each parameterized factory's tool definition exposes
  its parameters via `args`

#### Scenario: Model-visible schema is populated

- **GIVEN** any affected factory, e.g. `createGitMergeTool`, is
  invoked with injectable dependencies
- **WHEN** the returned tool definition is inspected
- **THEN** its `args` object contains exactly the factory's
  declared parameter names as keys

### Requirement: FR-002 Parameter fidelity

The zod conversion MUST preserve, for every affected parameter,
the name, primitive type (`string`, `boolean`, `number`),
optionality, and description carried by the legacy
`parameters:` declaration. Required parameters MUST map to
required zod fields; parameters whose legacy `execute` type is
optional (`?`) MUST map to `.optional()` zod fields. The
conversion MUST NOT introduce `.default()` values, renaming,
or dropped parameters.

#### Scenario: Optional boolean parameter converts faithfully

- **GIVEN** `createGitPushTool`'s legacy declaration had
  `setUpstream` (boolean, optional) and `branch` (string,
  optional)
- **WHEN** the tool definition is inspected after conversion
- **THEN** `args.setUpstream` accepts booleans and is optional
- **AND** `args.branch` accepts strings and is optional

#### Scenario: Numbered parameters with optionality convert faithfully

- **GIVEN** `createUfFinaleCiPollTool`'s legacy declaration had
  `prNumber` (number, required) and `timeoutMinutes` (number,
  optional)
- **WHEN** the tool definition is inspected after conversion
- **THEN** `args.prNumber` accepts numbers and is required
- **AND** `args.timeoutMinutes` accepts numbers and is optional
- **AND** the `timeoutMinutes` default of 15 continues to be
  applied inside `execute`, not by the schema

### Requirement: FR-003 Execute behavior unchanged

The migration MUST NOT modify any `execute` body, runtime
validation, result envelope, or error code. In particular the
shell-injection guards (`INVALID_PR_NUMBER` in
`uf_finale_ci_poll`, the `targetRepo` format check in
`uf_finale_create_pr`) MUST remain in place as defense-in-depth
behind the schema-validation layer. Existing unit tests that
invoke `execute` directly MUST continue to pass unmodified.

#### Scenario: Direct execute callers are unaffected

- **GIVEN** the existing `invokeTool(createGitPushTool(...), {`
  `setUpstream: true, branch: "feat/my-branch" })` test in
  `.opencode/test/uf-workflow.test.ts`
- **WHEN** the test suite runs after the migration
- **THEN** the test passes without modification

#### Scenario: Runtime validation survives schema validation

- **GIVEN** `uf_finale_ci_poll`'s `execute` receives a
  non-finite or out-of-range `prNumber`
- **WHEN** the execute body runs
- **THEN** it still returns the `INVALID_PR_NUMBER` failure
  envelope

### Requirement: FR-004 Schema-exposure regression tests

Each of the eleven affected factories MUST be covered by a
regression test asserting that the factory's returned tool
definition has an `args` object whose keys are exactly the
expected parameter names (sorted comparison), following the
`invoke-agent.test.ts` "exposes ... as a provider-free tool
definition" precedent. The tests MUST run in vitest without
network access, the opencode runtime, or any external service,
and MUST fail if a factory regresses to a `parameters` key
(which leaves `args` undefined) or drops/renames a parameter.

#### Scenario: Regression to the legacy key fails the suite

- **GIVEN** the schema-exposure tests are in place
- **WHEN** a factory's declaration is reverted to the legacy
  `parameters:` style
- **THEN** the corresponding schema-exposure test fails

#### Scenario: All affected factories are enumerated

- **GIVEN** the schema-exposure tests run
- **WHEN** their coverage is reviewed
- **THEN** all eleven affected factories across
  `uf-workflow.test.ts` and `uf-finale.test.ts` are asserted,
  including `persist_unleash_summary`'s already-correct
  declaration as a guard against future regression

### Requirement: FR-005 Scaffold parity

The canonical edits to `.opencode/plugins/uf-workflow/index.ts`
and `.opencode/plugins/uf-finale/index.ts` MUST be mirrored
byte-identically into
`internal/scaffold/assets/opencode/plugins/uf-workflow/index.ts`
and
`internal/scaffold/assets/opencode/plugins/uf-finale/index.ts`.
The existing embedded-asset drift tests MUST pass after the
sync; no new asset files are introduced and no dependency
changes are made.

#### Scenario: Drift tests pass after mirror sync

- **GIVEN** canonical plugin files are updated and mirrored to
  the scaffold assets
- **WHEN** the `internal/scaffold` drift tests run
- **THEN** canonical and scaffold copies match byte-for-byte

### Requirement: FR-006 Change record and documentation gate

The change MUST add a `CHANGELOG.md` entry describing the fixed
parameter-schema exposure. The documentation gate MUST be
assessed before PR submission: a documentation issue MUST be
filed in this repository for any user-facing impact (the
restored model-facing tool parameters), or the exemption MUST
be recorded in the task check-off with its justification and,
when filed, the issue number.

#### Scenario: Change record exists

- **GIVEN** the change is implemented
- **WHEN** `CHANGELOG.md` is read
- **THEN** an entry describes the repair of the `parameters` →
  `args` schema exposure for uf-workflow and uf-finale tools
