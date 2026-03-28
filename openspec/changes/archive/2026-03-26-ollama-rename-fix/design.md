## Context

Homebrew renamed `ollama` cask to `ollama-app`. All
references in setup, doctor, and tests need updating.

## Goals / Non-Goals

### Goals
- Replace `brew install ollama` with
  `brew install --cask ollama-app` in all Go code
- Update test assertions to match
- Update doctor hint strings

### Non-Goals
- Updating the dewey repo's GoReleaser config (separate
  repo, separate change)
- Changing how Ollama is detected (LookPath("ollama")
  is still correct -- the binary name didn't change)

## Decisions

The Ollama binary is still called `ollama` in PATH.
Only the Homebrew cask name changed from `ollama` to
`ollama-app`. So `LookPath("ollama")` remains correct.
Only the `brew install` command and hint strings change.

## Risks / Trade-offs

None. Pure string replacement.
