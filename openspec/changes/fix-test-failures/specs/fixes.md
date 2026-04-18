## MODIFIED Requirements

### Requirement: Scaffold asset sync

17 stale scaffold assets MUST be synced from their
canonical sources.

### Requirement: Ollama test GOOS

TestSetupRun_OllamaInstall MUST set opts.GOOS to
runtime.GOOS so the expected brew command matches
the actual runtime behavior.
