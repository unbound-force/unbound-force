## Context

`uf setup` installs every required tool except Ollama.
Ollama is a Formula in homebrew-core (`brew install ollama`),
not a cask. The existing installation functions follow
a consistent pattern: check LookPath, try Homebrew,
fallback with link.

## Goals / Non-Goals

### Goals
- Install Ollama automatically during `uf setup`
- Follow the exact pattern of `installGaze()`
- Position before Dewey step (Ollama is a prerequisite
  for `pullEmbeddingModel()`)
- Remove the post-setup Ollama tip (no longer needed)
- Update tests to expect the new step

### Non-Goals
- Starting the Ollama server (user manages this)
- Configuring Ollama (default config is sufficient)
- Installing Ollama on non-Homebrew systems via
  curl|bash (unlike OpenCode, Ollama's installer is
  not a simple curl pipe)

## Decisions

**Pattern**: Follow `installGaze()` exactly -- it's
the simplest installation function (Homebrew only, no
curl fallback, skip with link if no Homebrew).

```go
func installOllama(opts *Options, env doctor.DetectedEnvironment) stepResult {
    if _, err := opts.LookPath("ollama"); err == nil {
        return stepResult{name: "Ollama", action: "already installed"}
    }
    if opts.DryRun {
        if doctor.HasManager(env, doctor.ManagerHomebrew) {
            return stepResult{name: "Ollama", action: "dry-run",
                detail: "Would install: brew install ollama"}
        }
        return stepResult{name: "Ollama", action: "skipped",
            detail: "Homebrew not available"}
    }
    if !doctor.HasManager(env, doctor.ManagerHomebrew) {
        return stepResult{name: "Ollama", action: "skipped",
            detail: "Homebrew not available. Download from https://ollama.com/download"}
    }
    if _, err := opts.ExecCmd("brew", "install", "ollama"); err != nil {
        return stepResult{name: "Ollama", action: "failed",
            detail: "brew install failed", err: err}
    }
    return stepResult{name: "Ollama", action: "installed",
        detail: "via Homebrew"}
}
```

**Step order**: Insert Ollama between the Swarm steps
(step 8) and the Dewey step (currently step 9):

```
Step 1: OpenCode
Step 2: Gaze
Step 3: Node.js
Step 4: Bun
Step 5: Swarm plugin
Step 6: swarm setup
Step 7: opencode.json
Step 8: .hive/
Step 9: Ollama      ← NEW
Step 10: Dewey      (was step 9)
Step 11: uf init    (was step 10)
```

**Tip removal**: Remove the "Tip: Install Ollama"
block (lines ~253-258) since Ollama is now installed
automatically. Keep the embedding model alignment note.

## Risks / Trade-offs

**Low risk**: This follows an established pattern with
no new dependencies or architectural changes. The only
risk is the step ordering -- Ollama must be installed
before Dewey's `pullEmbeddingModel()` runs. The
proposed ordering handles this correctly.

**Homebrew-only**: Like Gaze, Ollama installation is
Homebrew-only. Non-Homebrew users get a skip with a
download link. This is acceptable since macOS and
Linux both support Homebrew, and the Ollama download
page covers all other platforms.
