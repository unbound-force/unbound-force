## Context

The Swarm plugin is installed by `unbound setup` and
checked by `unbound doctor`, but its opinionated behaviors
are not reflected in the Unbound Force agent configurations
or skill packages. When a developer runs `/swarm` on a
Speckit feature, Swarm's CASS engine re-decomposes work
that was already pre-decomposed in `tasks.md`. Hero agents
don't mention file reservations or session lifecycle. And
Ollama -- required for full semantic memory -- is invisible
to both doctor and setup.

## Goals / Non-Goals

### Goals
- Teach Swarm coordinator to follow Speckit's `tasks.md`
  as the authoritative decomposition when present
- Add Ollama health check to `unbound doctor` (optional,
  informational severity)
- Add Ollama install guidance to `unbound setup` output
- Update Cobalt-Crush agent with Swarm coordination
  protocol (file reservations, session lifecycle)
- Deploy the new skill via `unbound init` scaffold engine

### Non-Goals
- Modifying the Swarm plugin source code
- Adding cloud embedding provider support (tracked
  separately as a swarm-tools upstream PR)
- Making Ollama a required dependency
- Changing the hero lifecycle workflow (Spec 008)
- Auto-installing Ollama in setup (too large a download
  for automatic install)

## Decisions

### D1: `speckit-workflow` Skill Design

**Decision**: Create a Swarm skill at
`.opencode/skill/speckit-workflow/SKILL.md` that instructs
the coordinator to check for `tasks.md` before running
CASS decomposition.

**Rationale**: Swarm's skill system is the documented
extension point for teaching the coordinator about
domain-specific workflows. A skill file requires no code
changes to Swarm. The coordinator loads it via
`skills_use({ name: "speckit-workflow" })` and follows its
instructions.

**Skill content**:
- When `tasks.md` exists in a spec directory, it is the
  authoritative task decomposition
- Map tasks.md phases to Swarm epics
  (Phase 1 = epic "Setup", etc.)
- Map `[P]` parallel markers to concurrent worker spawning
- Map `[US?]` story labels to Swarm cell metadata
- Phase dependencies determine epic ordering
- Workers still use `swarmmail_reserve()` for file locking
- Workers still call `swarm_complete()` for learning

**Constitution alignment**: Autonomous Collaboration --
communication happens through the `tasks.md` artifact, not
runtime coupling between Speckit and Swarm.

### D2: Ollama Check in Doctor

**Decision**: Add Ollama as an optional check in the Core
Tools group of `unbound doctor`.

**Implementation**:
1. Check `ollama` binary via `opts.LookPath("ollama")`
2. If found, run `ollama list` via `opts.ExecCmd` and parse
   output for `mxbai-embed-large`
3. Report severity:
   - Binary found + model pulled = Pass (informational)
   - Binary found + model not pulled = Pass with hint
     `ollama pull mxbai-embed-large`
   - Binary not found = Pass (informational, optional)
     with hint `brew install ollama`
4. Never Fail or Warn -- Ollama is strictly optional

**Constitution alignment**: Composability First -- Ollama
remains optional with no mandatory dependency. Observable
Quality -- check result includes machine-parseable
provenance (binary path, model status).

### D3: Cobalt-Crush Agent Update

**Decision**: Add a "Swarm Coordination" section to the
Cobalt-Crush agent file with file reservation protocol and
session lifecycle guidance.

**Content**:
- When operating as a Swarm worker, MUST call
  `swarmmail_reserve()` before editing files
- Session end ritual: `hive_sync()` + `git push` before
  ending. "The plane is not landed until git push succeeds."
- Progress reporting via `swarm_progress()` at milestones
- Completion via `swarm_complete()` with files_touched list

**Constitution alignment**: This is guidance, not runtime
coupling. The agent works identically with or without
Swarm -- the guidance is conditional.

### D4: Scaffold Engine Update

**Decision**: Add the new `speckit-workflow/SKILL.md` to
the scaffold engine's embedded assets so `unbound init`
deploys it alongside the existing `unbound-force-heroes`
skill.

**Implementation**: Same pattern as other skill files:
- Add file to `internal/scaffold/assets/.opencode/skill/`
- Mark as tool-owned in `isToolOwned()`
- Update `knownEmbeddedFiles` count in tests
- Drift detection test ensures embedded copy matches
  canonical source

### D5: Setup Completion Summary Enhancement

**Decision**: When setup completes and Ollama is not
detected, add a line to the completion summary suggesting
Ollama installation for enhanced semantic memory.

**Implementation**: After printing "Setup complete!", check
for `ollama` binary. If not found, print:
```
Tip: Install Ollama for enhanced semantic memory:
  brew install ollama && ollama pull mxbai-embed-large
  (Without Ollama, semantic memory uses full-text search)
```

This is informational only, never an error or warning.

## Risks / Trade-offs

### R1: Skill Adoption Depends on Coordinator Loading It

The `speckit-workflow` skill only works if the coordinator
calls `skills_use({ name: "speckit-workflow" })`. If a
developer runs `/swarm` without loading the skill, Swarm
will re-decompose via CASS. Mitigation: the existing
`unbound-force-heroes` skill could reference
`speckit-workflow` as a prerequisite, so loading one
automatically suggests loading the other.

### R2: Ollama Check Adds Subprocess Call

Running `ollama list` adds ~200ms to the doctor run when
Ollama is installed. Acceptable given the 5-second target.
When Ollama is not installed, the check is a fast
LookPath failure with no subprocess.

### R3: Agent File Changes Are Opinionated

Adding Swarm-specific guidance to Cobalt-Crush may not
apply to all projects using unbound. Mitigation: the
guidance is conditional ("when operating as a Swarm
worker") and is in a clearly labeled section that users
can remove if not using Swarm.
