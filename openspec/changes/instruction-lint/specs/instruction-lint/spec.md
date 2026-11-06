<!-- capability: instruction-lint
     change: instruction-lint -->

Requirement numbering: this change allocates IL-FR-001
through IL-FR-004 for the four requirements below. It is the
founding delta of the instruction-lint capability, so no
ranges are reserved for prior changes; future changes of this
capability allocate from IL-FR-005.

## ADDED Requirements

### Requirement: Deterministic instruction lint

**ID: IL-FR-001** This requirement MUST be satisfied.

The repository MUST ship a deterministic linter
(`.opencode/lib/instruction-lint.ts`) over instruction files
(agents, commands, skills, convention packs, AGENTS.md) that
checks frontmatter shape, slash-command / agent / skill / path
references, pack rule markers and RFC 2119 force, vague
qualifiers inside normative sentences, per-file and load-set
token budgets, and near-duplicate paragraphs. Every check MUST
be mechanical: same input, same output.

#### Scenario: Findings carry review severity

- **GIVEN** any instruction file with a lint-detectable defect
- **WHEN** the linter emits a finding
- **THEN** the finding carries both its linter severity and its
  Divisor review severity
- **AND** JSON output reports file, line, rule, and message

### Requirement: Exit-code contract at every entry point

**ID: IL-FR-002** This requirement MUST be satisfied.

The linter MUST exit 0 when clean, 1 when it reports errors (or
warnings under --strict), and 2 when it could not run. Every
entry point that invokes the linter — the make target, the
pre-flight check, and the divisor-linguist persona — MUST
invoke it through the `.uf/scripts/instruction-lint` wrapper,
which verifies the linter entry file and the yaml and zod
packages before starting node.

#### Scenario: Fresh checkout without dependencies

- **GIVEN** a checkout where .opencode/node_modules is absent or
  incomplete
- **WHEN** the wrapper runs
- **THEN** it exits 2 with a remediation message naming the
  install command
- **AND** never exits 1 from a module-resolution crash

#### Scenario: Typo in --only

- **GIVEN** a --only path that matches no lintable file
- **WHEN** the linter runs
- **THEN** it exits 2 with the unmatched path on stderr
- **AND** reports no findings

### Requirement: Untrusted config validation

**ID: IL-FR-003** This requirement MUST be satisfied.

Config files MUST be validated against a strict schema before
use: unknown keys are rejected, and every ignorePaths entry
must be a non-empty string. An invalid config MUST exit 2
naming the offending field.

#### Scenario: Unknown config key

- **GIVEN** a config file containing {"disabled": ["IL-050"]}
- **WHEN** the linter runs
- **THEN** it exits 2 and names the unrecognized key on stderr
- **AND** never runs with default options as if the file were
  valid

#### Scenario: Empty ignorePaths entry

- **GIVEN** a config file containing {"ignorePaths": [""]}
- **WHEN** the linter runs
- **THEN** it exits 2 and names ignorePaths on stderr
- **AND** never suppresses findings

### Requirement: Hardened reviewer persona

**ID: IL-FR-004** This requirement MUST be satisfied.

The divisor-linguist persona MUST consume the linter as its
mechanism of record, MUST treat all inputs (diffs, pre-flight
output, linter output, file contents) as untrusted data, and
its bash permissions MUST deny edit and webfetch, allow only
the wrapper, `git diff`, and `git status`, and place terminal
metacharacter denies (redirection, pipes, separators, command
and process substitution) after every allow rule.

#### Scenario: Redirection under an allowed prefix

- **GIVEN** the persona's permission block
- **WHEN** a command like `git status > file` is evaluated
- **THEN** the terminal catch-all deny matches last and the
  command is denied

## MODIFIED Requirements

None. This is the founding delta of the instruction-lint
capability; there are no existing requirements to modify.

## REMOVED Requirements

None. No instruction-lint requirements existed before this
change.
