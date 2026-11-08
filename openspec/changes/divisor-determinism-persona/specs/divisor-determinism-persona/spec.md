<!-- capability: divisor-determinism-persona
     change: divisor-determinism-persona -->

Requirement numbering: this change allocates DDP-FR-001
through DDP-FR-004 for the four requirements below. It is the
founding delta of the divisor-determinism-persona capability,
so no ranges are reserved for prior changes; future changes of
this capability allocate from DDP-FR-005.

## ADDED Requirements

### Requirement: Determinism review ownership

**ID: DDP-FR-001** This requirement MUST be satisfied.

The repository MUST ship a repo-local Divisor persona
(`.opencode/agents/divisor-determinism.md`) that owns
determinism pack rules DE-001 through DE-008, enrolled in
`.uf/reviewer-capabilities.yaml`. The divisor-linguist
persona MUST NOT run a DE review; it MUST defer DE findings
and the IL-040..IL-042 hints to the determinism persona.

#### Scenario: Linguist defers a DE finding

- **GIVEN** a diff that violates a determinism pack rule
- **WHEN** the linguist persona reviews it
- **THEN** the linguist defers to the determinism persona
- **AND** does not report a duplicate DE finding

### Requirement: Deterministic evidence commands

**ID: DDP-FR-002** This requirement MUST be satisfied.

The persona's review MUST consume mechanical evidence:
instruction-lint deltas since a merge base (DE-004), the tool
inventory, and the seam report (DE-007). Every evidence CLI
MUST detect its entry point through the shared isMainModule
helper (import.meta.main is undefined before Node 24.2), MUST
convert crashes to exit 2, and MUST stamp JSON output with
producer provenance (tool, version, generatedAt).

#### Scenario: CLI runs where import.meta.main is undefined

- **GIVEN** a runtime where import.meta.main is undefined
- **WHEN** an evidence CLI is invoked directly
- **THEN** its entry block still runs
- **AND** a crash exits 2 with a stderr message, not 0 or 1

### Requirement: Deployment-neutral DE-008 deferral

**ID: DDP-FR-003** This requirement MUST be satisfied.

divisor-guard's DE-008 checklist MUST defer to a determinism
persona only when the deployment's reviewer capabilities
enroll one; the wording MUST NOT assume this repository's
instantiation. The deferral text MUST ship identically in the
canonical agent and its scaffold mirror.

#### Scenario: Deployed Guard without the persona

- **GIVEN** a deployment whose reviewer capabilities enroll
  no determinism persona
- **WHEN** divisor-guard reviews a DE-008 concern
- **THEN** the guard reports it under its own checklist
- **AND** defers nothing

### Requirement: Least-privilege persona surface

**ID: DDP-FR-004** This requirement MUST be satisfied.

Every Divisor persona whose bash permission is a map MUST end
the map with the terminal metacharacter denies (`*>*`, `*<*`,
`*|*`, `*;*`, `*&*`, `*$(*`, `` *`* ``). The determinism
persona MUST treat all review inputs -- diff, pre-flight
output, tool output, file contents -- as untrusted data.

#### Scenario: Terminal deny wins over a prefix allow

- **GIVEN** an allowed command prefix
- **WHEN** the command is suffixed with a shell
  metacharacter (for example `; cmd`, `> file`, `| sh`)
- **THEN** the last matching entry denies it
