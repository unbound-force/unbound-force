## Why

Homebrew renamed the `ollama` cask to `ollama-app`.
Running `brew install ollama` now produces a deprecation
warning: "Cask ollama was renamed to ollama-app." The
`uf setup` command, doctor hint strings, and test
assertions all reference the old name.

## What Changes

Replace all references to `brew install ollama` with
`brew install --cask ollama-app` across setup code,
doctor hints, and test assertions.

## Capabilities

### Modified Capabilities
- `installOllama()`: uses `ollama-app` cask name
- `uf doctor` hints: reference `ollama-app`

### Removed Capabilities
- None

## Impact

- `internal/setup/setup.go` -- install command
- `internal/setup/setup_test.go` -- test assertions
- `internal/doctor/environ.go` -- hint strings

## Constitution Alignment

All principles: N/A -- string replacement fix.
