## Context

The Unbound Force scaffold system deploys convention packs to
projects via `uf init`. Packs are Markdown files with YAML
frontmatter, embedded in the binary from
`internal/scaffold/assets/opencode/uf/packs/` and deployed to
`.opencode/uf/packs/` in target projects.

The Divisor SRE and Adversary agents reference `[PACK]` CI/CD
guidance during review, but no pack provides CI rules today. CI
workflow conventions exist only in the user-level
`~/.config/opencode/coding-standards.md`, which is invisible to
the `[PACK]` mechanism.

This design adds a `ci.md` convention pack as the canonical,
centralized home for all CI workflow authoring rules.

## Goals / Non-Goals

### Goals

- Add a `ci.md` canonical convention pack with rules for action
  pinning, SHA verification, workflow structure, permissions,
  and reusable workflow design
- Follow all established scaffold patterns (dual-copy drift
  detection, `-custom` companion, always-deployed, validator
  exemption)
- Fill the SRE/Adversary `[PACK]` gap for CI/CD guidance
- Centralize CI rules: `ci.md` pack is the single source of
  truth; trim the duplicated section from `coding-standards.md`

### Non-Goals

- CI guardrail implementation (reusable workflow in org-infra) --
  tracked separately as
  [complytime/org-infra#337](https://github.com/complytime/org-infra/issues/337)
- Changes to Divisor agent files -- the existing `[PACK]` tags
  will automatically resolve to the new pack's rules
- Container image conventions -- these remain in
  `coding-standards.md` (different concern)
- Python-specific or Rust-specific CI rules -- the pack is
  language-agnostic

## Decisions

### D1: Always-deployed, concern-specific pack

**Decision**: `ci.md` is always deployed regardless of detected
language, following the `content.md` and `severity.md` pattern.

**Rationale**: CI workflows are language-agnostic. Every project
that uses GitHub Actions benefits from these rules, regardless
of whether it's Go, TypeScript, or Python.

**Alternatives rejected**: Language-specific deployment (would
miss projects with CI but no detected language); `default.md`
extension (would mix CI concerns into the general pack and
inflate an already 214-line file).

### D2: Validator exemption for non-coding H2 sections

**Decision**: `ci.md` uses its own H2 section structure and is
exempted from the pack validator's 6-section requirement, like
`severity.md` and `content.md`.

**Rationale**: The standard coding pack sections (Coding Style,
Architectural Patterns, Security Checks, Testing Conventions,
Documentation Requirements, Custom Rules) do not map to CI
concerns. Forcing them would create empty sections or distort
the content.

**Sections chosen**:
- `## Action Pinning & Supply Chain` (CI-001, CI-002, CI-003)
- `## Workflow Structure` (CI-010, CI-011, CI-012)
- `## Permissions & Secrets` (CI-020, CI-021, CI-022)
- `## Reusable Workflow Design` (CI-030, CI-031, CI-032)
- `## Custom Rules` (empty, for `ci-custom.md`)

### D3: Rule numbering scheme

**Decision**: `CI-NNN` prefix with section-based grouping.
Supply chain: 001-009. Structure: 010-019. Permissions: 020-029.
Reusable design: 030-039. Leaves room for growth within each
section.

**Rationale**: Consistent with existing pack schemes (CS-NNN,
AP-NNN, SC-NNN, TC-NNN, DR-NNN in `default.md`; VB-NNN, TD-NNN,
etc. in `content.md`).

### D4: Centralization with `coding-standards.md` trim

**Decision**: The `ci.md` pack is the canonical home for CI
rules. The "YAML / GitHub Actions Workflows" section in
`~/.config/opencode/coding-standards.md` is replaced with a
reference to the pack.

**Rationale**: Avoids duplication. Convention packs are the
mechanism Divisor agents use for `[PACK]` lookups. All repos
using Unbound Force use `uf init`, so the pack is always
available. The `coding-standards.md` precedence hierarchy
("repository-level rules take precedence") already supports
this pattern.

**What stays in `coding-standards.md`**: The Containers section
(different concern -- container image building and supply chain,
not workflow authoring).

### D5: SHA verification rule semantics (CI-002)

**Decision**: Strong verification: agents MUST look up the tag
first, derive the SHA from it, and write both. SHAs MUST NOT be
sourced from memory or training data.

**Rationale**: The root cause of complytime/.github#114 was an
agent writing a SHA from training data. The rule must prevent
this at the behavioral level. Tag-first lookup eliminates the
hallucination vector entirely.

**Version comment (CI-003)**: SHOULD, not MUST. Missing comments
produce a warning, not a failure. This matches the CI guardrail
behavior defined in org-infra#337.

## Risks / Trade-offs

### R1: Pack proliferation

Adding a fourth always-deployed pack (default, severity, content,
ci) increases the CLAUDE.md and AGENTS.md import lists. This is
acceptable: each pack serves a distinct concern and the scaffold
system manages imports automatically.

### R2: Existing test count updates

Multiple test functions hardcode expected pack counts (e.g.,
`TestCollectDeployedPacks_Default` expects 5, `_Go` expects 7).
Adding ci.md and ci-custom.md increases each by 2. This is
mechanical but touches several test cases.

### R3: Stale `coding-standards.md` in other environments

Other developers using the same `coding-standards.md` will still
have the full CI section until they update. This is acceptable:
the pack takes precedence per the stated hierarchy, and the trim
is a user-config change, not a repo change.

### R4: CI guardrail dependency

The CI-002 rule (verify SHAs at authoring time) is a behavioral
rule enforced by convention, not automation. The enforcement
layer (org-infra#337) is tracked separately. Until implemented,
the convention pack provides prevention but not enforcement.
This is the same defense-in-depth pattern used for all other
convention pack rules.
