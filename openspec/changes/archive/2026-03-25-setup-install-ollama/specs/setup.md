## ADDED Requirements

### Requirement: setup-install-ollama

`uf setup` MUST install Ollama via Homebrew when it is
not present in the PATH.

#### Scenario: Ollama not installed, Homebrew available

- **GIVEN** Ollama is not in the PATH and Homebrew is
  available
- **WHEN** the developer runs `uf setup`
- **THEN** Ollama is installed via `brew install ollama`
  and the embedding model is subsequently pulled

#### Scenario: Ollama already installed

- **GIVEN** Ollama is already in the PATH
- **WHEN** the developer runs `uf setup`
- **THEN** the Ollama installation step reports
  "already installed" and proceeds to the Dewey step

#### Scenario: Homebrew not available

- **GIVEN** Ollama is not in the PATH and Homebrew is
  not available
- **WHEN** the developer runs `uf setup`
- **THEN** the Ollama step is skipped with a hint to
  download from https://ollama.com/download

## MODIFIED Requirements

### Requirement: setup-tip-removal

The post-setup Ollama installation tip MUST be removed
since Ollama is now installed automatically.

## REMOVED Requirements

None.
