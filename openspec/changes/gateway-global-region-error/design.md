## Context

`newVertexProvider()` in `internal/gateway/provider.go`
currently returns `*VertexProvider` (no error). When the
resolved region is `"global"`, it silently replaces it
with `"us-east5"` (line 182-184). This causes downstream
401 errors from Vertex AI because `rawPredict` requires a
specific regional endpoint and the user's credentials or
model access may not cover the fallback region.

The function is called from two places:
- `DetectProvider()` (line 66) -- auto-detection path
- `NewProviderByName()` (line 99) -- explicit `--provider
  vertex` path

Both callers already return `(Provider, error)`, so
propagating an error from `newVertexProvider()` is a
mechanical change with no architectural impact.

## Goals / Non-Goals

### Goals
- Return a clear, actionable error when the resolved
  Vertex region is `"global"`, explaining why it fails
  and how to fix it (set `ANTHROPIC_VERTEX_REGION`)
- Change `newVertexProvider()` signature to return
  `(*VertexProvider, error)` and propagate through callers
- Preserve the existing `"us-east5"` default for the
  empty-region case (no region env vars set at all)
- Add test coverage for the `"global"` rejection path

### Non-Goals
- Supporting `"global"` via a different API path
  (e.g., `generateContent` instead of `rawPredict`) --
  this would require a fundamentally different translation
  layer and is out of scope
- Changing the region resolution priority order (already
  correct per Spec 033/034)
- Adding region validation beyond `"global"` (other
  invalid regions will produce clear errors from Vertex
  API itself)
- Modifying sandbox behavior -- the sandbox calls
  `uf gateway start` as a subprocess; the error will
  surface naturally through the existing error path

## Decisions

### D1: Error instead of fallback

**Decision**: Return an error when region resolves to
`"global"` instead of silently falling back.

**Rationale**: The silent fallback violates Observable
Quality (constitution principle III) -- the gateway
produces a misleading 401 error minutes later instead of
an immediate, actionable error at startup. The error
message includes:
- The rejected region value
- Why it's rejected (rawPredict requires a specific
  region)
- The fix (`ANTHROPIC_VERTEX_REGION`)
- The env var priority chain for context

### D2: Validate in constructor, not in PrepareRequest

**Decision**: Validate the region in `newVertexProvider()`
(constructor) rather than in `PrepareRequest()` (per-
request).

**Rationale**: Fail-fast at gateway startup is better than
failing on the first request. The user sees the error
immediately when running `uf gateway start` rather than
after the gateway appears to start successfully. This
aligns with `VertexProvider.Start()` which already fails
fast on token acquisition errors.

### D3: Preserve empty-region default

**Decision**: Keep the `"us-east5"` default when no region
env vars are set (empty string case).

**Rationale**: This is the existing behavior for users who
rely on the default. Changing it would be a separate
concern. Only the `"global"` case is problematic because
it indicates the user has explicitly configured a region
that cannot work with rawPredict.

## Risks / Trade-offs

### R1: Breaking change for users with `global` set

**Risk**: Users who previously had `VERTEX_LOCATION=global`
and the gateway "worked" (because the fallback to us-east5
happened to be correct for their project) will now see an
error.

**Mitigation**: The error message clearly states the fix
(set `ANTHROPIC_VERTEX_REGION`). These users were already
in a fragile state -- the silent fallback only worked by
coincidence if their project happened to have Claude enabled
in us-east5.

### R2: Signature change in unexported function

**Risk**: Changing `newVertexProvider()` from returning
`*VertexProvider` to `(*VertexProvider, error)` requires
updating all call sites.

**Mitigation**: There are exactly two call sites
(`DetectProvider` line 66, `NewProviderByName` line 99),
both already return `(Provider, error)`. The change is
mechanical.
