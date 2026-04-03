# Contract: Doctor Embedding Capability Check

**Package**: `internal/doctor`  
**File**: `checks.go`

## Function: `checkEmbeddingCapability`

```go
func checkEmbeddingCapability(opts *Options) CheckResult
```

### Preconditions

- `opts.EmbedCheck` is non-nil (set by `defaults()`)
- Called only when Dewey binary is found (after
  `LookPath("dewey")` succeeds in `checkDewey()`)

### Postconditions

- Returns a `CheckResult` with `Name: "embedding capability"`
- On success: `Severity: Pass`, `Message` contains model name
- On failure: `Severity: Warn`, `InstallHint` contains
  actionable fix command
- Never panics — all errors are caught and returned as
  `Warn` severity results

### Behavior Matrix

| Condition | Severity | Message | InstallHint |
|-----------|----------|---------|-------------|
| EmbedCheck returns nil | Pass | "granite-embedding:30m generating embeddings" | (empty) |
| EmbedCheck returns error (connection refused) | Warn | "cannot generate embeddings (Ollama not running)" | "Start Ollama: ollama serve" |
| EmbedCheck returns error (model not found) | Warn | "cannot generate embeddings (model not loaded)" | "ollama pull granite-embedding:30m" |
| EmbedCheck returns error (other) | Warn | "cannot generate embeddings" | "Start Ollama: ollama serve, then: ollama pull granite-embedding:30m" |
| Dewey not installed | (skipped) | "skipped: dewey not installed" | (empty) |

## Function: `defaultEmbedCheck`

```go
func defaultEmbedCheck(getenv func(string) string) func(model string) error
```

### Preconditions

- `getenv` is non-nil

### Postconditions

- Returns a function that tests embedding generation
- The returned function sends a POST to Ollama's
  `/api/embed` endpoint
- Uses `OLLAMA_HOST` env var (default `http://localhost:11434`)
- Timeout: 5 seconds (embedding should be fast for a
  3-word test input)
- Returns `nil` on success, descriptive `error` on failure

### Error Categories

| Error | Cause | Hint for User |
|-------|-------|---------------|
| "connection refused" | Ollama not running | Start Ollama |
| "model not found" | Model not pulled | Pull model |
| "empty embeddings" | Unknown model issue | Re-pull model |

> **Note**: Timeout errors (Ollama overloaded or
> unresponsive) fall into the "other" category in the
> behavior matrix above and receive the combined hint.
> The 5-second timeout is set in `defaultEmbedCheck`.

## Function: `checkDewey` (modified)

### Changes from Current Behavior

1. **New check**: After the existing "embedding model"
   check, add an "embedding capability" check that calls
   `opts.EmbedCheck`
2. **Ollama demotion**: When Dewey is installed, the
   existing "embedding model" check message is annotated
   with "(Dewey manages Ollama lifecycle)" to indicate
   that Ollama serving status is managed by Dewey
3. **Skip logic**: When Dewey is not installed, the
   embedding capability check is skipped (added to the
   skip block alongside "embedding model" and "workspace")

### Updated Check Order in Dewey Group

1. `dewey binary` — existing, unchanged
2. `embedding model` — existing, message updated
3. `embedding capability` — NEW
4. `workspace` — existing, unchanged
