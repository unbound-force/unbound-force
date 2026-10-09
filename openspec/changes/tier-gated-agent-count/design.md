## Context

The review-dispatch plugin's `buildPlan()` function
(~line 1488) iterates over discovered agents, calls
`relevance()` for scope matching, then assembles plan
entries with run resolution. The `diffTier()` result
(lightweight / standard / heavy) currently flows only
into model profile selection — it does not affect which
agents are included.

The proposal (see proposal.md) introduces tier-based
agent count caps to reduce cost and latency for trivial
changes. This design documents how the cap integrates
into the existing dispatch machinery.

## Goals / Non-Goals

### Goals
- Cap included agent count for lightweight-tier changes
  (default: 2 agents)
- Make caps configurable per tier via the review matrix
  `limits` section
- Preserve existing standard and heavy behavior
- Record tier-cap decisions in plan output with distinct
  reason codes for observability (Observable Quality)
- Maintain `--full` flag bypass behavior

### Non-Goals
- Changing the `relevance()` scope-matching logic itself
- Adding per-agent priority or weight configuration
- Modifying `diffTier()` classification thresholds
- Tier-based model selection changes (already working)
- Changing issue-mode (`triage`) dispatch behavior

## Decisions

### D1: Post-Relevance Cap Filter

The tier cap is applied as a post-filter after all
`relevance()` decisions are made but before plan entry
assembly. This is a two-pass approach within `buildPlan()`:

**Pass 1** (existing): Evaluate `relevance()` for every
discovered agent. Collect the set of included agents and
their reason codes.

**Pass 2** (new): If the change profile's tier has a
configured cap and the included agent count exceeds it,
select agents to keep using a priority order and
limit-skip the rest.

**Rationale**: Post-filter preserves the existing relevance
logic untouched. The cap is purely subtractive — it never
adds agents, only removes excess ones. This maintains
Composability First: each agent's relevance decision is
independent of the cap.

### D2: Agent Priority Order for Cap Selection

When the cap requires dropping agents, the following
priority order determines which agents survive:

1. **Always-required** agents (`divisor-guard`,
   `divisor-adversary`) — reason code `always-required`
2. **Scope-matched** agents — reason code `scope-match`,
   ordered alphabetically for determinism
3. **Curator** — reason code `curator-relevant`

Within each priority tier, agents are ordered
alphabetically by agent name for deterministic,
reproducible plans.

**Floor enforcement**: The effective cap is clamped to
`max(cap, always_required_count)` where
`always_required_count` is the number of agents with
reason code `always-required` in the included set.
This guarantees that always-required agents are never
dropped, regardless of custom `tier_caps` configuration.

**Rationale**: Guard and adversary are baseline reviewers
that the existing `relevance()` function marks as
always-required. They provide the minimum coverage
guarantee. Floor clamping makes this invariant
non-bypassable via configuration. Alphabetical ordering
within tiers ensures plan determinism (Observable
Quality).

### D3: Configurable Tier Caps in Review Matrix

New optional `tier_caps` field in the review matrix
`limits` section:

```yaml
limits:
  max_personas: 16
  max_total_runs: 24
  tier_caps:
    lightweight: 2
    standard: ~   # null = no cap
    heavy: ~      # null = no cap
```

When `tier_caps` is absent or a tier's value is null,
no cap is applied for that tier (current behavior).

The `Limits` interface and `LimitsSchema` are extended
with the optional `tier_caps` field. `limitsFor()` merges
defaults with matrix overrides as it does today.

**Default values** (applied when `tier_caps` is not
configured):
- `lightweight`: 2
- `standard`: null (no cap)
- `heavy`: null (no cap)

**Rationale**: Configurable caps allow teams to tune
dispatch cost without code changes. The default of 2 for
lightweight matches the issue's proposed behavior.
With 2 always-required agents, a lightweight cap of 2
yields 2 always-required + 0 scope-matched; raising to
3 adds 1 scope-matched agent. Null for standard/heavy
preserves existing behavior exactly.

**Adversarial input handling**:
- `tier_caps.lightweight: 1` (below always-required
  count): clamped to `max(1, 2) = 2` by floor
  enforcement (D2). The agent is never silently dropped.
- Very large caps (e.g., `tier_caps.lightweight: 100`):
  no effect — the cap exceeds the included agent count,
  so all agents pass through.
- Inconsistent cross-tier values (e.g., lightweight > 
  standard): accepted — each tier is evaluated
  independently. The schema validates each value is a
  positive integer or null; cross-tier consistency is
  not enforced.
- `tier_caps.heavy: N` with security-sensitive diff:
  accepted but an advisory warning is emitted in the
  plan output (see D7).

### D4: Plan Entry Representation

Agents excluded by the tier cap receive plan entries with:
- `decision: "skip"`
- `reason_code: "tier-cap"`
- `reason: "tier cap of N exceeded for <tier> tier"`

This is distinct from `scope-miss` (agent scopes don't
match), `content-capability` (content-only persona), and
existing skip reasons. Tier-cap exclusions are plan-level
decisions and do not produce runs. They are observable
via plan entries with `decision: "skip"` and
`reason_code: "tier-cap"`. The `run_counts` object
tracks actual run statuses only.

**Rationale**: The distinct `tier-cap` reason code allows
consumers to distinguish tier-cap skips from scope-miss
and content-capability skips (Observable Quality). Plan
entries are the authoritative record; no run-level
counter is needed.

### D5: Full Flag Bypass

When `input.full` is true, tier caps are bypassed
entirely. The full panel already bypasses scope matching
in `relevance()` — tier caps follow the same bypass
path.

**Rationale**: `--full` is the explicit opt-in for
maximum coverage. Applying tier caps to full-panel
mode would contradict user intent.

### D6: Mode Exclusion

Tier caps apply only to `code` and `specs` command
modes (diff-based profiles). The `triage`, `feedback`,
and `test` modes are excluded from tier caps because
their dispatch logic serves different purposes
(classification or test review, not review cost
optimization).

**Rationale**: Issue triage already has distinct
profiling (`profileIssue()`) and different agent
relevance patterns. Feedback mode assesses reviewer
comments, not diff content. Test review mode uses
spec-based evaluation. Applying diff-tier caps to
these modes would conflate unrelated cost concerns.

### D7: Heavy Tier Cap Advisory Warning

When `tier_caps.heavy` is configured (non-null) and the
change profile has `securitySensitive: true`, the
planner MUST emit a plan-level advisory in the
`advisories` array with severity `HIGH` and a message
noting that a custom heavy cap is active on a
security-sensitive review.

**Rationale**: Heavy tier caps are legitimate for
non-sensitive heavy diffs (e.g., large documentation
changes). But when a diff is security-sensitive,
reducing the agent count is a deliberate trade-off
that should be observable (Constitution V: Security by
Default, Constitution III: Observable Quality). The
advisory does not block — it surfaces the operator's
choice for audit.

## Risks / Trade-offs

### R1: Reduced Coverage on Lightweight Changes

Capping at 2 agents means some scope-matched agents
will not review lightweight changes. This is the
intended trade-off: lightweight changes (< 50 lines,
< 3 files, single component, not security-sensitive)
are low-risk by definition.

**Mitigation**: The cap is configurable. Teams can raise
the lightweight cap or set it to null to restore full
coverage. The `--full` flag provides an escape hatch
for any individual review.

### R2: Priority Order May Miss Relevant Agents

Alphabetical ordering within priority tiers is simple
but not optimal — the most relevant scope-matched agent
may not be first alphabetically. For example,
`divisor-testing` (scope: test-quality) might be more
relevant than `divisor-architect` for a test-only change,
but architect sorts first.

**Mitigation**: The lightweight tier is for trivial
changes where deep specialty review has diminishing
returns. For changes where specific agent expertise
matters, the change should profile as standard or heavy
(> 50 lines, > 3 files, multi-component, or
security-sensitive).

### R3: Schema Backward Compatibility

Adding `tier_caps` to the limits schema is additive and
optional. Existing review matrices without `tier_caps`
get default values. The plan output schema gains no new
top-level fields — `tier-cap` is a new value for the
existing `reason_code` string field.
