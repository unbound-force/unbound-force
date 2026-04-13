## Context

The `forwardedAPIKeys` slice in `config.go` lists env
vars forwarded from host to container via Podman's
`-e VAR` syntax. Currently includes 6 entries covering
Anthropic (direct), OpenAI, Gemini, OpenRouter, and
Google Cloud (project + location).

The Anthropic-via-Vertex integration path uses
different env vars from the native Google Vertex AI
path:

| Integration | Env Vars |
|-------------|----------|
| Google Vertex AI (native) | `GOOGLE_CLOUD_PROJECT`, `VERTEX_LOCATION` |
| Anthropic via Vertex | `ANTHROPIC_VERTEX_PROJECT_ID`, `VERTEX_LOCATION`, `CLAUDE_CODE_USE_VERTEX` |

Both `VERTEX_LOCATION` is already forwarded. The missing
vars are `ANTHROPIC_VERTEX_PROJECT_ID` (which GCP
project to use) and `CLAUDE_CODE_USE_VERTEX` (signals
to use the Vertex backend).

## Goals / Non-Goals

### Goals

- Forward `ANTHROPIC_VERTEX_PROJECT_ID` and
  `CLAUDE_CODE_USE_VERTEX` to the container
- Verify forwarding works in tests

### Non-Goals

- No changes to the Google Cloud credential mounting
  logic (already handles service account keys and ADC)
- No changes to CDE backend credential injection

## Decisions

### D1: Add both vars unconditionally

Both vars are added to `forwardedAPIKeys`. They are
only forwarded when set in the host environment
(the `if v := opts.Getenv(key); v != ""` check in
`forwardedEnvVars()` handles this). No conditional
logic needed.
