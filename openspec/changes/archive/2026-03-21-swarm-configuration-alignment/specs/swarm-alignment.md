## ADDED Requirements

### Requirement: Ollama Health Check in Doctor

`unbound doctor` SHOULD check for the `ollama` binary in
PATH and, when found, SHOULD verify that the
`mxbai-embed-large` model is pulled by parsing the output
of `ollama list`. The check MUST use informational
severity (Pass) regardless of outcome -- Ollama is
strictly optional.

#### Scenario: Ollama installed with model

- **GIVEN** `ollama` is in PATH and `ollama list` output
  contains `mxbai-embed-large`
- **WHEN** the developer runs `unbound doctor`
- **THEN** the Core Tools group shows Ollama as Pass with
  message "mxbai-embed-large model ready"

#### Scenario: Ollama installed without model

- **GIVEN** `ollama` is in PATH but `ollama list` output
  does not contain `mxbai-embed-large`
- **WHEN** the developer runs `unbound doctor`
- **THEN** the Core Tools group shows Ollama as Pass with
  install hint `ollama pull mxbai-embed-large`

#### Scenario: Ollama not installed

- **GIVEN** `ollama` is not in PATH
- **WHEN** the developer runs `unbound doctor`
- **THEN** the Core Tools group shows Ollama as Pass
  (informational) with install hint
  `brew install ollama && ollama pull mxbai-embed-large`

---

### Requirement: Speckit Workflow Swarm Skill

A Swarm skill file MUST be created at
`.opencode/skill/speckit-workflow/SKILL.md` that teaches
the Swarm coordinator to use `tasks.md` as the
authoritative task decomposition when present.

The skill MUST instruct the coordinator to:
- Check for `tasks.md` in the active spec directory before
  running CASS decomposition
- Map tasks.md phases to Swarm epics
- Interpret `[P]` markers as parallel-safe subtasks
- Interpret `[US?]` labels as cell metadata
- Respect phase dependency ordering
- Still enforce file reservations and completion protocol

#### Scenario: tasks.md exists

- **GIVEN** a Speckit feature branch with `tasks.md` in
  `specs/NNN-*/`
- **WHEN** the coordinator loads the `speckit-workflow`
  skill and decomposes a task
- **THEN** the coordinator reads `tasks.md` and maps its
  phases to Swarm epics rather than generating a new
  decomposition via CASS

#### Scenario: tasks.md does not exist

- **GIVEN** no `tasks.md` exists in the working directory
- **WHEN** the coordinator loads the `speckit-workflow`
  skill
- **THEN** the skill instructs the coordinator to proceed
  with standard CASS decomposition

---

### Requirement: Cobalt-Crush Swarm Coordination Protocol

The Cobalt-Crush agent file MUST be updated with a "Swarm
Coordination" section containing:

1. File reservation protocol: When operating as a Swarm
   worker, MUST call `swarmmail_reserve()` before editing
   files
2. Session lifecycle: Session MUST end with `hive_sync()`
   followed by `git push`
3. Progress reporting: SHOULD call `swarm_progress()` at
   milestones (25%, 50%, 75%)
4. Completion: MUST call `swarm_complete()` with
   `files_touched` list when done

#### Scenario: Cobalt-Crush as Swarm worker

- **GIVEN** Cobalt-Crush is spawned as a Swarm worker via
  `swarm_spawn_subtask()`
- **WHEN** the worker begins editing files
- **THEN** the worker calls `swarmmail_reserve()` with
  the file paths before making changes

#### Scenario: Session end

- **GIVEN** Cobalt-Crush has completed its assigned work
- **WHEN** the session is ending
- **THEN** the agent calls `hive_sync()` and verifies
  `git push` succeeds

---

### Requirement: Setup Ollama Guidance

`unbound setup` SHOULD include Ollama install instructions
in its completion summary when `ollama` is not detected in
PATH.

#### Scenario: Setup completes without Ollama

- **GIVEN** `ollama` is not in PATH when setup completes
- **WHEN** the completion summary is printed
- **THEN** the summary includes a "Tip" line with Ollama
  install instructions and a note that semantic memory
  falls back to full-text search without it

---

### Requirement: Scaffold Engine Deploys Speckit Workflow Skill

The scaffold engine MUST deploy
`.opencode/skill/speckit-workflow/SKILL.md` when running
`unbound init`. The file MUST be classified as tool-owned.

#### Scenario: unbound init deploys skill

- **GIVEN** a project directory without the
  `speckit-workflow` skill
- **WHEN** the developer runs `unbound init`
- **THEN** `.opencode/skill/speckit-workflow/SKILL.md` is
  created with the speckit workflow instructions

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
