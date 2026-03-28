## MODIFIED Requirements

### Requirement: ollama-cask-name

All Homebrew install commands for Ollama MUST use
`ollama-app` (the current cask name), not `ollama`
(deprecated).

#### Scenario: uf setup installs ollama

- **GIVEN** Ollama is not installed
- **WHEN** `uf setup` runs
- **THEN** it executes `brew install --cask ollama-app`

#### Scenario: doctor hint

- **GIVEN** Ollama is not installed
- **WHEN** `uf doctor` runs
- **THEN** the hint says `brew install --cask ollama-app`
