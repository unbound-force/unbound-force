<!--
  SYNC IMPACT REPORT
  ==================
  Version change: 1.1.0 → 1.2.0 (MINOR: new principle added)

  Added principles:
    - V. Security by Default

  Unchanged principles:
    - I. Autonomous Collaboration
    - II. Composability First
    - III. Observable Quality
    - IV. Testability

  Unchanged sections:
    - Hero Constitution Alignment
    - Development Workflow
    - Governance

  Templates requiring updates:
    ✅ .specify/templates/plan-template.md — no changes needed;
       Constitution Check section is generic and will align at
       plan time using these five principles.
    ✅ .specify/templates/spec-template.md — no changes needed.
    ✅ .specify/templates/tasks-template.md — no changes needed.
    ✅ .specify/templates/checklist-template.md — no changes needed.
    ✅ .specify/templates/agent-file-template.md — no changes needed.

  Hero constitution alignment:
    ✅ Gaze v1.1.0 — Testability principle already matches.
    ⚠  Gaze v1.1.0 — Will need Principle V alignment review.
    ⚠  Website v1.0.0 — Will need Testability + Principle V
       alignment review.
-->

# Unbound Force Constitution

## Core Principles

### I. Autonomous Collaboration

Heroes MUST collaborate through well-defined artifacts — files,
reports, and schemas — rather than runtime coupling or synchronous
interaction.

- Every hero MUST be able to complete its primary function without
  requiring synchronous interaction with another hero. A hero MAY
  consume another hero's artifacts, but MUST NOT block waiting for
  a response.
- Hero outputs MUST be self-describing: each artifact MUST contain
  enough metadata (producer identity, version, timestamp, artifact
  type) for any consumer to interpret it without consulting the
  producing hero.
- Inter-hero communication MUST use the artifact envelope format
  defined by the Hero Interface Contract. Heroes MUST NOT invent
  ad-hoc exchange formats.
- Heroes SHOULD publish artifacts to a well-known location within
  the project repository so other heroes can discover them without
  explicit coordination.

**Rationale**: A swarm of autonomous agents cannot rely on real-time
negotiation. Artifact-based communication makes collaboration
asynchronous, auditable, and resilient to individual hero
unavailability. If one hero is not deployed, the others continue
to function — they simply have fewer artifacts to consume.

### II. Composability First

Every hero MUST be independently installable and usable without any
other hero being present. Combining heroes MUST produce additive
value without introducing mandatory dependencies.

- A hero MUST deliver its core value when deployed alone. No hero
  MAY require another hero as a hard prerequisite for installation
  or primary operation.
- Heroes MUST expose well-defined extension points (configuration,
  artifact consumption, convention packs) for integration rather
  than requiring modification of their internals. No hero MAY
  require patching or forking another hero to integrate.
- When two or more heroes are deployed together, their combination
  MUST produce value greater than the sum of their individual
  capabilities (e.g., Gaze quality reports informing Mx F metrics).
  This additive value MUST NOT come at the cost of standalone
  functionality.
- Heroes SHOULD auto-detect the presence of other heroes and
  activate enhanced functionality when peers are available, without
  requiring manual configuration.

**Rationale**: Adoption friction kills tools. A team that only needs
a tester should be able to deploy Gaze alone. A team that only needs
reviews should deploy The Divisor alone. Composability ensures each
hero earns its place independently, and the swarm becomes compelling
only when its parts are already individually valuable.

### III. Observable Quality

Every hero MUST produce machine-parseable output alongside any
human-readable output. All quality claims MUST be backed by
automated, reproducible evidence.

- Every hero that produces output MUST support at minimum a JSON
  format. Human-readable output (terminal text, Markdown) is
  RECOMMENDED but MUST NOT be the only format available.
- All artifacts MUST include provenance metadata: which hero
  produced the output, which version of the hero, when it was
  produced, and against what input (branch, commit, backlog item).
- Quality claims — accuracy rates, coverage percentages, scoring
  thresholds — MUST be backed by automated regression tests or
  benchmarks that can be re-run by any contributor.
- Metrics MUST be comparable across runs. Output formats MUST be
  stable enough that tooling built on a hero's output does not
  break between minor versions.
- Heroes SHOULD produce artifacts that conform to registered
  schemas in the shared data model, enabling cross-hero analysis
  without bespoke parsing.

**Rationale**: A swarm that cannot measure its own performance
cannot improve. Machine-parseable output enables Mx F to track
trends, Muti-Mind to make data-driven prioritization decisions,
and The Divisor to ground reviews in evidence rather than opinion.
Provenance metadata ensures that every data point is traceable to
its source, preventing "trust me" assertions.

### IV. Testability

Every component built within the Unbound Force ecosystem MUST be
testable in isolation without requiring external services, network
access, or shared mutable state.

- Test contracts MUST verify observable side effects (return values,
  state mutations, I/O operations) rather than implementation details.
- Coverage strategy (unit vs. integration vs. e2e, with specific
  targets) MUST be defined in the implementation plan for all new code.
- Coverage ratchets MUST be enforced by automated tests; any coverage
  regression MUST be treated as a test failure and block the build.
- Missing coverage strategy in a spec or plan is a CRITICAL-severity
  finding and MUST be resolved before implementation begins.

**Rationale**: AI agents generate code rapidly. If that code is not
structurally testable, the resulting system will quickly collapse under
its own unverified complexity. Testability is a first-class governance
concern because untestable code cannot be reliably verified by Gaze or
any other automated mechanism. Unverified code cannot be trusted.

### V. Security by Default

Every component built within the Unbound Force ecosystem MUST treat
security as a structural property, not a review-time afterthought.
Supply chain integrity, input validation, and least privilege MUST be
enforced by design.

- **Supply chain integrity**: Dependencies MUST be verified by content
  hash (SHA256 or equivalent) when downloaded outside a package manager's
  built-in verification. Pinning by name and version alone (tag, semver)
  is name-addressed and insufficient — publishers can replace artifacts
  under the same tag. CI pipelines MUST pin actions and reusable
  workflows by commit SHA, not mutable tags.
- **Input validation**: All external inputs (user input, API payloads,
  file contents, environment variables used as data, CI workflow inputs)
  MUST be validated and sanitized before reaching any security-sensitive
  operation (file path construction, shell execution, query building,
  privilege decisions). Validation MUST reject unexpected types, lengths,
  formats, and case variations.
- **Least privilege**: Components MUST operate with the minimum
  permissions necessary. CI containers SHOULD NOT run with `--privileged`
  unless explicitly justified. Secrets MUST be scoped to the narrowest
  context needed. File permissions MUST default to restrictive values
  (0o644 for files, 0o755 for executables and directories).
- **Dependency necessity**: Before adding an external dependency, the
  adopter MUST justify that the project's existing toolchain cannot
  cover the same use case. Every dependency is attack surface; the
  default answer is "do not add."

**Rationale**: AI agents make adding dependencies and generating code
trivially fast. Without structural security guardrails, the attack
surface of the system grows with each generation cycle. Security by
Default ensures that supply chain integrity, input hygiene, and
privilege boundaries are enforced before code reaches review — not
discovered during review when anchoring bias and finding fragmentation
can cause under-classification.

## Hero Constitution Alignment

Every hero repository MUST maintain its own constitution in
`.specify/memory/constitution.md`. Hero constitutions extend the
org constitution — they MUST NOT contradict any org principle.

- Hero constitutions MUST include a `parent_constitution` reference
  indicating which version of the Unbound Force org constitution
  they align with.
- Hero constitutions MAY add principles beyond the five org
  principles, provided the additional principles do not contradict
  any org-level MUST rule.
- When the org constitution is amended, all hero constitutions MUST
  be reviewed for continued alignment. If a MUST rule is added or
  changed, hero repositories MUST open an alignment issue within
  one release cycle and resolve it before the next major version.
- Hero constitutions that predate this org constitution MUST be
  reviewed for alignment but are not automatically invalidated.
  Contradictions MUST be resolved by amending the hero constitution.

## Development Workflow

- **Branching**: All work MUST occur on feature branches. Direct
  commits to the main branch are prohibited except for trivial
  documentation fixes.
- **Code Review**: Every pull request MUST receive at least one
  approving review before merge. When The Divisor is deployed,
  its council protocol SHOULD be used for review.
- **Continuous Integration**: The CI pipeline MUST pass (build,
  lint, tests) before a pull request is eligible for merge.
- **Releases**: Follow semantic versioning (MAJOR.MINOR.PATCH).
  Breaking changes to public APIs, artifact schemas, or analysis
  behavior require a MAJOR bump.
- **Commit Messages**: Use conventional commit format
  (`type: description`) to enable automated changelog generation.
- **Spec-Driven Development**: Features SHOULD follow the speckit
  pipeline (constitution → specify → clarify → plan → tasks →
  analyze → checklist → implement) to ensure requirements are
  captured before implementation begins.
- **Gatekeeping Integrity**: Agents MUST NOT modify values that
  serve as quality or governance gates — including but not limited
  to: coverage thresholds, severity definitions, MUST/SHOULD rule
  classifications, CI flags (`-race`, `-count=1`), review iteration
  limits, agent temperature and tool-access settings, and pinned
  dependency versions. When an implementation cannot meet a gate,
  the agent MUST report the failure and stop rather than weakening
  the gate.
- **Phase Discipline**: Each pipeline phase (specify,
  clarify, plan, tasks, analyze, checklist, implement,
  review) MUST produce only its designated artifacts.
  Implementation code MUST NOT be written during
  specification phases. Specification phases may only
  write to files within the `specs/NNN-*/` feature
  directory.
- **Cross-Repo Documentation**: When a change affects user-facing
  behavior, hero capabilities, CLI commands, or workflows, a
  GitHub issue MUST be created in the `unbound-force/website`
  repository to track any required documentation or website
  updates. The issue MUST be created before the implementing
  PR is merged. Changes that are purely internal (refactoring,
  test-only, CI-only) are exempt.

## Governance

This constitution is the highest-authority document in the Unbound
Force organization. On matters of cross-cutting concern — inter-hero
communication, output formats, standalone usability, quality
standards — this constitution supersedes all hero-level constitutions
and project-specific guidance.

- **Supremacy**: When a hero constitution and this org constitution
  conflict, the org constitution prevails. The hero constitution
  MUST be amended to resolve the conflict.
- **Amendments**: Any change to this constitution MUST be proposed
  via pull request, reviewed, and approved before merge. The
  amendment MUST include a migration plan if it alters or removes
  existing principles. All hero constitutions MUST be reviewed for
  continued alignment after any amendment.
- **Versioning**: The constitution follows semantic versioning:
  - MAJOR: Principle removal or incompatible redefinition of a
    MUST rule.
  - MINOR: New principle added or materially expanded guidance.
  - PATCH: Clarifications, wording, or non-semantic refinements.
- **Compliance Review**: At each planning phase (spec, plan, tasks),
  the Constitution Check gate MUST verify that the proposed work
  aligns with all active org principles. Constitution violations
  are CRITICAL severity and non-negotiable.
- **Conflict Resolution**: When two org principles appear to
  conflict in a specific scenario, the tradeoff MUST be explicitly
  documented in the relevant spec or plan. No principle has
  implicit priority over another; resolution is context-dependent
  and requires written justification.

**Version**: 1.2.0 | **Ratified**: 2026-02-25 | **Last Amended**: 2026-06-01
