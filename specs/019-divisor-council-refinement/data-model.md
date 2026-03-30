# Data Model: Divisor Council Refinement

**Spec**: 019 | **Date**: 2026-03-30

## Persona Ownership Mapping

Each review dimension is owned by exactly one Divisor
persona. This mapping is the authoritative source for
de-duplication (per Spec 019 FR-004, FR-005).

### The Adversary — Security & Resilience

| Dimension | Scope | Examples |
|-----------|-------|---------|
| Secrets/credentials | Hardcoded secrets, API keys, tokens, passwords, internal hostnames | `.env` in VCS, plaintext password in config |
| Dependency CVEs/supply chain | Known vulnerabilities in direct/transitive deps, unpinned CI actions | CVE in `go.mod` dep, `actions/checkout@main` |
| Error handling/resilience | Panic vs error, unchecked assertions, nil dereferences, recovery paths | `panic()` in library code, ignored `err` return |
| Path/injection safety | Path traversal, command injection, YAML injection, symlink escape | `filepath.Join(userInput, "...")` without validation |

**Out of scope**: Test isolation (→ Tester), zero-waste
(→ Guard), efficiency/performance (→ SRE), file
permissions (→ SRE), plan alignment (→ Guard),
architectural patterns (→ Architect).

### The Tester — Test Quality & Coverage

| Dimension | Scope | Examples |
|-----------|-------|---------|
| Test architecture | Arrange/act/assert structure, table-driven tests, fixture quality | Missing assertion, test without cleanup |
| Coverage strategy | Contract surface coverage, risk-appropriate depth, acceptance traceability | Function with 0% coverage, missing edge case |
| Assertion depth | Specific value checks vs. `err == nil`, field-level verification | `if err != nil` without checking error message |
| Test isolation | Shared mutable state, execution order dependency, external access, race conditions | Package-level var modified by tests, network call in unit test |
| Regression protection | Known-good/bad scenarios, bug regression tests, schema validation | Missing regression test for fixed bug |

**Out of scope**: Security (→ Adversary), operational
concerns (→ SRE), plan alignment (→ Guard), architectural
patterns (→ Architect).

### The Guard — Intent & Governance

| Dimension | Scope | Examples |
|-----------|-------|---------|
| Plan alignment/intent drift | Spec-to-implementation fidelity, scope creep, scope contraction | Feature not in spec, acceptance criterion unaddressed |
| Zero-waste mandate | Orphaned code/specs, unused imports/deps, dead code, gold plating | Unused function, aspirational doc with no tasks |
| Constitution alignment | Principle compliance, trade-off documentation, governance adherence | Violating Composability without justification |
| Cross-component value | Neighborhood rule, downstream impact, documentation consistency | Shared schema change without consumer update |

**Out of scope**: Security (→ Adversary), test quality
(→ Tester), operational readiness (→ SRE), coding
conventions (→ Architect).

### The SRE (Operator) — Operations & Efficiency

| Dimension | Scope | Examples |
|-----------|-------|---------|
| File permissions/hardcoded config | File mode on created files, hardcoded paths/hostnames, env assumptions | `0o777` on config file, hardcoded `/usr/local/bin` |
| Efficiency/performance | O(n²) loops, redundant I/O, unnecessary allocations, memory copies | Nested loop over same data, reading file twice |
| Release pipeline integrity | Reproducible builds, pinned deps, signing, platform coverage | Floating dependency version, missing smoke test |
| Dependency health | Unused deps, update mechanisms, version pinning | Stale dep with no update strategy |
| Runtime observability | Exit codes, actionable errors, structured output, version metadata | Generic "error occurred" message, no JSON output |
| Upgrade/migration paths | Version skew handling, backward compatibility, breaking change docs | Format change without migration path |
| Operational documentation | README completeness, failure mode docs, runbook | Missing troubleshooting section |
| Backup/recovery | Destructive operations, partial failure handling, re-runnability | `--force` without confirmation, corrupted half-state |

**Out of scope**: Security/credentials (→ Adversary),
test quality (→ Tester), intent drift (→ Guard),
architectural patterns (→ Architect).

### The Architect — Structure & Conventions

| Dimension | Scope | Examples |
|-----------|-------|---------|
| Architectural alignment | Project structure, layer separation, package boundaries, asset sync | Business logic in CLI layer, import cycle |
| Key pattern adherence | Established patterns (Options/Result, delegation, file ownership) | Competing pattern for same abstraction |
| Coding convention compliance [PACK] | Formatting, naming, comments, error handling per pack | Missing GoDoc on exported function |
| Testing convention compliance [PACK] | Test framework, assertion style, naming per pack | Using testify when pack requires stdlib |
| Documentation compliance [PACK] | Code comments, spec writing conventions per pack | Missing RFC-style language in requirements |
| DRY/structural integrity | Duplicated logic, unnecessary abstractions, refactoring difficulty | Copy-pasted function, premature abstraction |

**Out of scope**: Security (→ Adversary), test coverage
depth (→ Tester), intent drift (→ Guard), operational
readiness (→ SRE).

---

## Severity Level Definitions

These definitions are shared across all 5 Divisor personas
via the `severity.md` convention pack. They define the
boundary between levels and provide domain-specific
examples for each persona's review domain.

### CRITICAL

**Definition**: The change introduces a defect that will
cause data loss, security breach, build failure, or
constitutional violation. The change MUST NOT be merged.

**Boundary**: Immediate, concrete harm. Not theoretical
risk — actual breakage or exposure.

| Persona | Examples |
|---------|---------|
| Adversary | Hardcoded production secret, SQL injection vector, panic in library code |
| Tester | Missing coverage strategy in spec/plan (Constitution IV violation), test that masks a real failure |
| Guard | Constitution principle violated without justification, implementation contradicts spec acceptance criteria |
| SRE | Release pipeline broken (won't produce artifacts), destructive operation without guard, critical CVE in dependency |
| Architect | Fundamental misalignment with project architecture (score 1-2), circular dependency introduced |

### HIGH

**Definition**: The change introduces significant risk or
technical debt that will cause problems if not addressed
before merge. Blocks the review.

**Boundary**: Likely to cause problems in the near term.
Requires action but not an emergency.

| Persona | Examples |
|---------|---------|
| Adversary | Credentials logged at INFO level, unpinned CI action on mutable tag, unchecked type assertion |
| Tester | Vague acceptance criteria ("works correctly"), shallow assertions (err == nil only), missing regression test for known bug |
| Guard | Scope creep beyond spec, acceptance criterion with no corresponding task, undocumented constitution trade-off |
| SRE | Missing upgrade path for format change, hardcoded environment values, no error recovery for I/O failure |
| Architect | Notable architectural deviation (score 5-6), competing pattern for same abstraction, significant DRY violation |

### MEDIUM

**Definition**: The change has a quality issue that should
be addressed but does not block the merge. In Spec Review
Mode, auto-fixable.

**Boundary**: Improvement opportunity. The code/spec works
but could be better.

| Persona | Examples |
|---------|---------|
| Adversary | Overly broad file permissions (0o755 → 0o644), missing context in error wrap, redundant file read |
| Tester | Missing fixture specification, test isolation concern (shared state but no observed failure), convention deviation |
| Guard | Minor scope addition beyond spec (gold plating), stale cross-reference, metadata inconsistency |
| SRE | Missing operational documentation section, incomplete platform support, unquantified performance requirement |
| Architect | Minor convention deviation, missing GoDoc on exported function, test naming doesn't follow pattern |

### LOW

**Definition**: Minor style or documentation improvement.
Non-blocking. In Spec Review Mode, auto-fixable.

**Boundary**: Cosmetic or informational. No functional
impact.

| Persona | Examples |
|---------|---------|
| Adversary | Comment suggesting security review for future feature, minor naming inconsistency in error variable |
| Tester | Minor test naming convention issue, optional observability enhancement in test output |
| Guard | Minor documentation wording improvement, optional cross-reference addition |
| SRE | Style improvement in error messages, optional health check enhancement, minor doc gap |
| Architect | Formatting preference, optional structural improvement, minor comment enhancement |

---

## Auto-Fix Policy (Spec Review Mode)

| Severity | Action | Rationale |
|----------|--------|-----------|
| LOW | Auto-fix | Cosmetic; safe to fix without human judgment |
| MEDIUM | Auto-fix | Quality improvement; deterministic fix |
| HIGH | Report only | Requires human judgment on intent/scope |
| CRITICAL | Report only | Requires human judgment; may indicate design issue |

This policy is already implemented in `/review-council`
Spec Review Mode (step 3). The severity definitions above
ensure all 5 personas classify the same type of issue at
the same level, making the auto-fix boundary predictable.

---

## Qualified FR Reference Format

All functional requirement references in Divisor agent
files MUST use the fully qualified format:

```
per Spec NNN FR-XXX
```

Examples:
- "per Spec 005 FR-020" (not "FR-020")
- "per Spec 019 FR-004" (not "FR-004")
- "per Spec 002 FR-010" (not "FR-010")

This eliminates cross-spec ambiguity. FR-020 exists in
7 files across 6 specs with different meanings.

### Affected References in Current Agent Files

| Agent | Current Reference | Qualified Form |
|-------|------------------|----------------|
| divisor-adversary.md | "FR-020" (§5 heading) | "per Spec 005 FR-020" |

All other FR references in the current agent files are
generic (not spec-specific). The refactored agents will
use qualified references wherever a specific FR is cited.
