## Why

The `uf gateway` Vertex AI provider silently replaces the
`"global"` region with `us-east5` when any of
`ANTHROPIC_VERTEX_REGION`, `VERTEX_LOCATION`, or
`CLOUD_ML_REGION` is set to `"global"`
(`internal/gateway/provider.go:182-184`). This causes a
confusing `401 UNAUTHENTICATED` /
`ACCESS_TOKEN_TYPE_UNSUPPORTED` error because the gateway
sends requests to a region the user didn't configure and
where their credentials or model access may not apply.

The `"global"` value is increasingly common because Google
Cloud admins set `VERTEX_LOCATION=global` or
`CLOUD_ML_REGION=global` for non-rawPredict workloads
(Gemini, embeddings, etc.). Vertex AI's `rawPredict` and
`streamRawPredict` endpoints -- which the gateway uses for
Claude -- require a specific regional endpoint
(e.g., `us-east5-aiplatform.googleapis.com`), not the
`global` endpoint.

The silent fallback violates the principle of least surprise
and makes the root cause nearly impossible to diagnose from
the error message alone.

## What Changes

Replace the silent `"global"` → `"us-east5"` fallback in
`newVertexProvider()` with a clear, actionable error that
tells the user exactly what's wrong and how to fix it.

The function signature changes from returning
`*VertexProvider` to returning `(*VertexProvider, error)`,
and all callers (`DetectProvider`, `NewProviderByName`)
propagate the error.

## Capabilities

### New Capabilities
- **Clear error on `global` region**: When the resolved
  Vertex region is `"global"`, the gateway returns an
  actionable error message explaining that `rawPredict`
  requires a specific region and suggesting
  `ANTHROPIC_VERTEX_REGION` as the override.

### Modified Capabilities
- **`newVertexProvider()` signature**: Returns
  `(*VertexProvider, error)` instead of `*VertexProvider`.
- **`DetectProvider()`**: Propagates region validation
  errors from `newVertexProvider()`.
- **`NewProviderByName()`**: Propagates region validation
  errors from `newVertexProvider()`.

### Removed Capabilities
- **Silent `"global"` fallback**: The implicit
  `"global"` → `"us-east5"` replacement is removed.
  Users who previously relied on this (unknowingly)
  will now get a clear error with a fix instruction.

## Impact

- **Files**: `internal/gateway/provider.go`,
  `internal/gateway/gateway_test.go`
- **Behavioral**: Users with `VERTEX_LOCATION=global` or
  `CLOUD_ML_REGION=global` will see a clear error on
  `uf gateway start` instead of a confusing 401 at
  request time. The fix is to set
  `ANTHROPIC_VERTEX_REGION` to a specific region.
- **Backward compatibility**: Users who already set
  `ANTHROPIC_VERTEX_REGION` to a specific region are
  unaffected (it takes priority 1). Users with no region
  set still get the `us-east5` default. Only users
  explicitly setting `"global"` are affected -- they were
  previously getting silent misrouting and auth failures.
- **Sandbox**: `autoStartGateway()` in
  `internal/sandbox/sandbox.go` calls `DetectProvider()`
  indirectly via `uf gateway start`. The error will surface
  at sandbox startup, preventing the confusing in-session
  401.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

This change affects the gateway's internal provider
construction. No inter-hero artifact formats, communication
protocols, or artifact-based interfaces are modified.

### II. Composability First

**Assessment**: PASS

The gateway remains independently usable. The change
improves standalone usability by providing clear error
messages instead of silent misrouting. No new dependencies
are introduced.

### III. Observable Quality

**Assessment**: PASS

The change improves observability by replacing a silent
fallback (which produced a misleading 401 at request time)
with an immediate, machine-parseable error at startup. The
error message includes the rejected value, the reason, and
the recommended fix.

### IV. Testability

**Assessment**: PASS

The `newVertexProvider()` function remains testable in
isolation via injected `getenv` and `execCmd` functions.
The new error return is directly assertable in unit tests
without requiring external services.
