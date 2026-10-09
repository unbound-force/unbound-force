## ADDED Requirements

### Requirement: Re-exports from invoke-agent

The `invoke-agent/index.ts` module MUST re-export all
types and the `sanitizeInvocationError` function that move
to `agent-executor.ts`. Existing imports from
`./invoke-agent/index.js` MUST continue to resolve without
code changes.

Re-exported types: `ModelIdentity`, `InvocationTokens`,
`InvocationUsage`, `InvocationError`,
`InvocationProvenance`, `InvokeAgentResult`.

Re-exported functions: `sanitizeInvocationError`.

#### Scenario: Existing test imports remain stable

- **GIVEN** a test file importing `InvokeAgentResult` from
  `./invoke-agent/index.js`
- **WHEN** the library extraction is complete
- **THEN** the import MUST resolve without changes to the
  test file

#### Scenario: Re-export type identity

- **GIVEN** `InvokeAgentResult` is re-exported from
  `invoke-agent/index.ts` and also exported from
  `lib/agent-executor.ts`
- **WHEN** both are imported into the same module
- **THEN** the types MUST be structurally identical (same
  type — re-export, not duplication)

### Requirement: Dual-copy Sync

All file changes MUST be applied to both locations:
- `.opencode/lib/` and `.opencode/plugins/` (scaffolded)
- `internal/scaffold/assets/opencode/lib/` and
  `internal/scaffold/assets/opencode/plugins/` (canonical)

Existing drift detection tests MUST pass after all changes
are applied.

#### Scenario: Drift detection catches desync

- **GIVEN** `agent-executor.ts` is modified in `.opencode/lib/`
  but not in `internal/scaffold/assets/opencode/lib/`
- **WHEN** drift detection tests run
- **THEN** the tests MUST fail, identifying the desynchronized
  file

## MODIFIED Requirements

### Requirement: invoke-agent Module Exports

Previously: `invoke-agent/index.ts` defined and exported
all types and functions directly.

Now: `invoke-agent/index.ts` re-exports moved types from
`../lib/agent-executor.js` and retains only
`invoke_agent`-specific logic: input schema, input
validation, prompt resolution (`prompt` vs `promptFile`),
manifest validation, host model resolution, and the tool
factory.

The `InvokeAgentDependencies` type and
`createInvokeAgentTool()` function remain defined in
`invoke-agent/index.ts` (they are plugin-specific, not
shared).

## REMOVED Requirements

(None)
