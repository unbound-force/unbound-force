## Why

The Swarm plugin has opinionated behaviors -- TDD mandate,
file reservations, session lifecycle ("land the plane"),
CASS task decomposition -- that are not reflected in the
Unbound Force agent configurations, doctor checks, or setup
automation. Developers using `unbound setup` get Swarm
installed but not properly configured for Unbound's
spec-driven workflows. There is no guidance on how Swarm
interacts with Speckit's `tasks.md` decomposition, no
awareness of file reservation protocol in hero agents, and
no Ollama health check for semantic memory.

## What Changes

1. Add Ollama health check to `unbound doctor` (optional
   binary + model pulled check)
2. Add Ollama install suggestion to `unbound setup`
   completion summary
3. Create a `speckit-workflow` Swarm skill that teaches the
   coordinator to use `tasks.md` as authoritative task
   decomposition instead of CASS re-decomposition
4. Update Cobalt-Crush agent with Swarm file reservation
   protocol and session-end ritual (hive_sync + git push)
5. Update scaffold engine to deploy the new skill file

## Capabilities

### New Capabilities
- `speckit-workflow` skill: Teaches Swarm coordinator to
  follow Speckit's pre-decomposed `tasks.md` rather than
  generating its own decomposition via CASS. Includes
  phase dependency rules and parallel marker interpretation.
- Ollama doctor check: Reports whether Ollama is installed
  and whether `mxbai-embed-large` model is pulled. Enables
  developers to know if semantic memory will use vector
  search (with Ollama) or fall back to full-text search.

### Modified Capabilities
- `unbound doctor`: Enhanced with Ollama check in the Core
  Tools group (optional/informational severity)
- `unbound setup`: Completion summary includes Ollama
  install instructions when Ollama is not detected
- `cobalt-crush-dev.md`: Updated with Swarm coordination
  protocol (file reservations, session lifecycle)
- Scaffold engine: Deploys the new `speckit-workflow` skill
  alongside the existing `unbound-force-heroes` skill

### Removed Capabilities
- None

## Impact

- `internal/doctor/checks.go`: Add Ollama binary and model
  checks to `checkCoreTools()`
- `internal/setup/setup.go`: Add Ollama suggestion to
  completion summary
- `.opencode/skill/speckit-workflow/SKILL.md`: New file
- `.opencode/agents/cobalt-crush-dev.md`: Updated content
- `internal/scaffold/`: Updated embedded assets and tests
  for new skill file

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

The new `speckit-workflow` skill communicates through a
well-defined artifact (`tasks.md`) rather than runtime
coupling. It teaches the Swarm coordinator to read and
follow an existing file-based plan. The Ollama doctor check
produces self-describing results in the existing Report
struct with all metadata fields (severity, install hint,
detail).

### II. Composability First

**Assessment**: PASS

Ollama remains optional -- semantic memory falls back to
full-text search without it. The `speckit-workflow` skill
is independently installable and only activates when
`tasks.md` exists. Cobalt-Crush continues to work without
Swarm (the file reservation guidance is conditional on
operating as a Swarm worker). No new mandatory dependencies
are introduced.

### III. Observable Quality

**Assessment**: PASS

The Ollama check produces machine-parseable JSON output
through the existing Report/CheckResult structure with
provenance metadata (binary path, model name). The skill
file uses the standard SKILL.md format with YAML
frontmatter that passes existing validation.

### IV. Testability

**Assessment**: PASS

The Ollama check uses the injected ExecCmd function to run
`ollama list`, making it testable with stub responses.
The skill file is validated by the existing
`checkSkillIntegrity()` function. The scaffold engine
changes are tested by the existing drift detection tests.
No external services or network required for testing.
