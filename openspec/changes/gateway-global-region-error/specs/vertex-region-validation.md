## ADDED Requirements

### Requirement: Global region rejection

`newVertexProvider()` MUST return an error when the
resolved Vertex region equals `"global"` (case-sensitive).
The error message MUST include:
1. The rejected region value
2. The reason (`rawPredict` requires a specific region)
3. The recommended fix (`ANTHROPIC_VERTEX_REGION`)
4. Examples of valid regions

#### Scenario: All region env vars set to global

- **GIVEN** `VERTEX_LOCATION=global` and
  `CLOUD_ML_REGION=global` and
  `ANTHROPIC_VERTEX_REGION` is unset
- **WHEN** `newVertexProvider()` is called
- **THEN** it MUST return a non-nil error containing
  the string `"global"` and
  `"ANTHROPIC_VERTEX_REGION"`

#### Scenario: ANTHROPIC_VERTEX_REGION overrides global

- **GIVEN** `VERTEX_LOCATION=global` and
  `ANTHROPIC_VERTEX_REGION=us-east5`
- **WHEN** `newVertexProvider()` is called
- **THEN** it MUST return a `*VertexProvider` with
  region `"us-east5"` and a nil error

#### Scenario: CLOUD_ML_REGION=global alone

- **GIVEN** `VERTEX_LOCATION` is unset and
  `CLOUD_ML_REGION=global` and
  `ANTHROPIC_VERTEX_REGION` is unset
- **WHEN** `newVertexProvider()` is called
- **THEN** it MUST return a non-nil error containing
  the string `"global"` and
  `"ANTHROPIC_VERTEX_REGION"`

#### Scenario: No region env vars set

- **GIVEN** `VERTEX_LOCATION`, `CLOUD_ML_REGION`, and
  `ANTHROPIC_VERTEX_REGION` are all unset
- **WHEN** `newVertexProvider()` is called
- **THEN** it MUST return a `*VertexProvider` with
  region `"us-east5"` and a nil error (default
  behavior preserved)

### Requirement: Error propagation to callers

`DetectProvider()` and `NewProviderByName()` MUST
propagate errors returned by `newVertexProvider()`.

#### Scenario: DetectProvider with global region

- **GIVEN** `CLAUDE_CODE_USE_VERTEX=1` and
  `ANTHROPIC_VERTEX_PROJECT_ID=my-project` and
  `VERTEX_LOCATION=global`
- **WHEN** `DetectProvider()` is called
- **THEN** it MUST return `(nil, error)` where the
  error contains `"global"`

#### Scenario: NewProviderByName with global region

- **GIVEN** `VERTEX_LOCATION=global`
- **WHEN** `NewProviderByName("vertex", ...)` is called
- **THEN** it MUST return `(nil, error)` where the
  error contains `"global"`

## MODIFIED Requirements

### Requirement: newVertexProvider function signature

`newVertexProvider()` MUST return
`(*VertexProvider, error)` instead of `*VertexProvider`.

Previously: `newVertexProvider()` returned
`*VertexProvider` with no error channel; invalid regions
were silently replaced.

## REMOVED Requirements

### Requirement: Silent global-to-us-east5 fallback

The behavior where `region == "global"` is silently
replaced with `"us-east5"` (line 182-184 of
`provider.go`) is removed. This silent fallback caused
confusing downstream 401 errors from Vertex AI because
the user's credentials or model access did not cover
the fallback region.
