## ADDED Requirements

### FR-001: Tier-Based Agent Count Cap

The dispatch planner MUST apply a configurable agent
count cap based on the change profile's tier
classification. After `relevance()` scope matching
determines which agents are included, the planner
MUST limit the number of included agents to the
tier's configured cap value.

Default cap values:
- `lightweight`: 2
- `standard`: null (no cap)
- `heavy`: null (no cap)

When the cap is null or absent for a tier, the planner
MUST NOT limit agent count for that tier.

When the cap is exceeded, the planner MUST select
agents to retain using the following priority order:
1. Always-required agents (`always-required` reason
   code)
2. Scope-matched agents (`scope-match` reason code),
   ordered alphabetically by agent name
3. Curator agents (`curator-relevant` reason code)

The effective cap MUST be clamped to
`max(cap, always_required_count)` where
`always_required_count` is the number of agents in
the included set with reason code `always-required`.
This guarantees that always-required agents are never
dropped by the tier cap regardless of configuration.

#### Scenario 1: Lightweight change dispatches capped agents

- **GIVEN** a diff with 30 lines, 2 files, 1 component,
  and no security sensitivity (lightweight tier)
- **AND** 4 agents pass relevance scope matching
- **WHEN** `plan_review_dispatch` builds the dispatch plan
- **THEN** the plan MUST include at most 2 agents
- **AND** agents excluded by the cap MUST have
  `decision: "skip"` with `reason_code: "tier-cap"`

#### Scenario 2: Standard change is not capped

- **GIVEN** a diff with 100 lines, 5 files, 2 components
  (standard tier)
- **AND** 4 agents pass relevance scope matching
- **WHEN** `plan_review_dispatch` builds the dispatch plan
- **THEN** all 4 agents MUST be included
- **AND** no agent MUST have `reason_code: "tier-cap"`

#### Scenario 3: Heavy change is not capped

- **GIVEN** a diff with 500 lines, 12 files, and security
  sensitivity (heavy tier)
- **AND** 5 agents pass relevance scope matching
- **WHEN** `plan_review_dispatch` builds the dispatch plan
- **THEN** all 5 agents MUST be included

#### Scenario 4: Always-required agents are exempt from tier cap

- **GIVEN** a lightweight-tier change with cap of 2
- **AND** `divisor-guard` and `divisor-adversary` are both
  always-required
- **AND** `divisor-architect` is scope-matched
- **WHEN** the planner applies the tier cap
- **THEN** `divisor-guard` and `divisor-adversary` MUST be
  included (floor enforcement: 2 always-required = cap)
- **AND** `divisor-architect` MUST be skipped with
  `reason_code: "tier-cap"`

#### Scenario 5: Floor clamps cap below always-required count

- **GIVEN** a review matrix with
  `limits.tier_caps.lightweight: 1`
- **AND** 2 always-required agents and 1 scope-matched
  agent pass relevance
- **WHEN** the planner applies the tier cap
- **THEN** the effective cap MUST be clamped to 2
  (`max(1, 2)`)
- **AND** both always-required agents MUST be included
- **AND** the scope-matched agent MUST be skipped with
  `reason_code: "tier-cap"`

### FR-002: Full Flag Bypasses Tier Cap

When the `full` flag is true, the planner MUST NOT
apply tier-based agent count caps. All agents that
pass relevance scope matching MUST be included.

#### Scenario 6: Full panel ignores tier cap

- **GIVEN** a lightweight-tier change
- **AND** `full` is true
- **AND** 5 agents pass relevance scope matching
- **WHEN** `plan_review_dispatch` builds the dispatch plan
- **THEN** all 5 agents MUST be included
- **AND** no agent MUST have `reason_code: "tier-cap"`

### FR-003: Configurable Tier Caps

The review matrix `limits` section MUST support an
optional `tier_caps` field with per-tier agent count
limits:

```yaml
limits:
  tier_caps:
    lightweight: 2
    standard: ~
    heavy: ~
```

Each tier value MUST be either a positive integer or
null. When null, no cap is applied for that tier.
When the `tier_caps` field is absent from the matrix,
the planner MUST use default values (lightweight: 2,
standard: null, heavy: null).

#### Scenario 7: Custom lightweight cap overrides default

- **GIVEN** a review matrix with
  `limits.tier_caps.lightweight: 3`
- **AND** a lightweight-tier change with 4 agents passing
  relevance
- **WHEN** `plan_review_dispatch` builds the plan
- **THEN** at most 3 agents MUST be included

#### Scenario 8: Missing tier_caps uses defaults

- **GIVEN** a review matrix with no `tier_caps` in limits
- **AND** a lightweight-tier change with 4 agents passing
  relevance
- **WHEN** `plan_review_dispatch` builds the plan
- **THEN** at most 2 agents MUST be included (default
  lightweight cap)

#### Scenario 8a: Invalid tier cap values are rejected

- **GIVEN** a review matrix with
  `limits.tier_caps.lightweight: 0` (or a negative
  integer, or a non-integer like 2.5)
- **WHEN** the dispatch planner validates the schema
- **THEN** the value MUST be rejected by Zod schema
  validation (`z.number().int().positive()`)

### FR-004: Tier Cap Plan Entry Format

Agents excluded by the tier cap MUST receive plan
entries with:
- `decision: "skip"`
- `reason_code: "tier-cap"`
- `reason`: A human-readable message including the cap
  value and tier name

Tier-cap exclusions are plan-level decisions and do not
produce runs. They are observable via plan entries with
`decision: "skip"` and `reason_code: "tier-cap"`. The
`run_counts` object tracks actual run statuses only.

#### Scenario 9: Tier-cap skip is recorded in plan

- **GIVEN** a lightweight-tier change with cap of 2
- **AND** `divisor-guard` (always-required),
  `divisor-architect` (scope-match), and
  `divisor-testing` (scope-match) pass relevance
- **WHEN** the planner applies the tier cap
- **THEN** the plan MUST include entries for all 3 agents
- **AND** `divisor-guard` and `divisor-architect` MUST
  have `decision: "include"`
- **AND** `divisor-testing` MUST have `decision: "skip"`
  with `reason_code: "tier-cap"`

### FR-005: Mode Exclusion

Tier caps MUST apply only to `code` and `specs`
command modes. The `triage`, `feedback`, and `test`
modes MUST NOT apply tier caps.

#### Scenario 10: Issue triage ignores tier caps

- **GIVEN** an issue profile with lightweight tier
- **AND** 4 agents pass relevance scope matching
- **AND** command mode is `triage`
- **WHEN** `plan_review_dispatch` builds the plan
- **THEN** all 4 agents MUST be included
- **AND** no agent MUST have `reason_code: "tier-cap"`

#### Scenario 11: Feedback mode ignores tier caps

- **GIVEN** a lightweight-tier change
- **AND** 3 agents pass relevance scope matching
- **AND** command mode is `feedback`
- **WHEN** `plan_review_dispatch` builds the plan
- **THEN** all 3 agents MUST be included
- **AND** no agent MUST have `reason_code: "tier-cap"`

### FR-006: Heavy Tier Cap Advisory Warning

When `tier_caps.heavy` is configured (non-null) and
the change profile has `securitySensitive: true`, the
planner MUST emit a plan-level advisory with severity
`HIGH` noting that a custom heavy cap is active on
a security-sensitive review. HIGH severity ensures
the advisory surfaces in the verdict reason and cannot
be silently ignored, per Constitution V (Security by
Default).

#### Scenario 12: Advisory emitted for heavy cap on
  security-sensitive diff

- **GIVEN** a review matrix with
  `limits.tier_caps.heavy: 3`
- **AND** a heavy-tier, security-sensitive diff
- **AND** 5 agents pass relevance scope matching
- **WHEN** `plan_review_dispatch` builds the plan
- **THEN** at most 3 agents MUST be included
- **AND** the plan MUST contain an advisory with
  severity `HIGH`

## References

- [proposal.md](../proposal.md): Motivation and
  constitution alignment
- [design.md](../design.md): Design decisions D1-D7
- [tasks.md](../tasks.md): Implementation task breakdown

## MODIFIED Requirements

(None)

## REMOVED Requirements

(None)
