## Why

`uf sandbox start` forwards LLM API keys to the
container but is missing the environment variables
needed for OpenCode's Anthropic-via-Vertex integration.
Without `ANTHROPIC_VERTEX_PROJECT_ID` and
`CLAUDE_CODE_USE_VERTEX`, OpenCode inside the container
cannot connect to Claude models hosted on Google
Vertex AI.

The current forwarded list covers direct Anthropic,
OpenAI, Gemini, OpenRouter, and Google Cloud project/
location — but not the Anthropic SDK's Vertex-specific
variables that OpenCode uses when the model string is
`google-vertex-anthropic/claude-*`.

## What Changes

### Modified Capabilities

- `forwardedAPIKeys` in `internal/sandbox/config.go`:
  Add `ANTHROPIC_VERTEX_PROJECT_ID` and
  `CLAUDE_CODE_USE_VERTEX` to the forwarded list.

### New Capabilities

None.

### Removed Capabilities

None.

## Impact

- 1 file modified: `internal/sandbox/config.go`
  (2 lines added to `forwardedAPIKeys`)
- 1 test file updated: `internal/sandbox/sandbox_test.go`
  (verify new vars are forwarded)
- No new Go logic — adding string entries to an
  existing slice

## Constitution Alignment

### I–IV: All N/A

This is a 2-line config change to a string slice.
No architectural, composability, quality, or
testability implications.
