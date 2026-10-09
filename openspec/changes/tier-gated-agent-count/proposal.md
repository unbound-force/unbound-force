## Why

The `diffTier()` classification (lightweight / standard / heavy)
in `review-dispatch/index.ts` currently controls only which
model profile is selected for each agent run. All discovered
agents still pass through `relevance()` scope matching
regardless of tier. This means a 20-line test fix that happens
to touch categories matching 5 agents runs the full 5-agent
council with expensive models.

Observed on PR #653: a scaffold sync plus one test assertion
fix profiled as heavy (1337 lines, security_sensitive) and
dispatched 5 agents. The meaningful code change was ~20 lines.
This wastes cost and latency on trivial changes without
improving coverage quality.

## What Changes

Introduce tier-based agent count caps that limit how many
agents are included in the dispatch plan, applied after
`relevance()` scope matching and before plan entry assembly.

- **Lightweight tier**: Cap at 2 included agents. With the
  default always-required agents (`divisor-guard` and
  `divisor-adversary`), a lightweight cap of 2 yields 2
  always-required + 0 scope-matched agents. Raising the
  cap to 3 would add 1 scope-matched agent.
- **Always-required floor**: The effective cap is clamped
  to `max(cap, always_required_count)` so that
  always-required agents are never dropped by the tier cap.
  A custom `tier_caps.lightweight: 1` is clamped to 2 when
  2 agents are always-required.
- **Standard tier**: No cap (current behavior).
- **Heavy tier**: No cap (current behavior). When a custom
  `tier_caps.heavy` is configured and the diff is
  security-sensitive, the planner MUST emit a plan-level
  advisory warning for observability.
- Agent caps are configurable in the review matrix under
  `limits`, alongside existing `max_personas` and
  `max_total_runs`.
- The `full` flag (`--full`) bypasses tier caps, same as
  it currently bypasses scope pruning.

## Capabilities

### New Capabilities
- `tier-based agent cap`: Limits the number of included
  agents based on the change profile's tier classification.
  Lightweight changes dispatch fewer agents, reducing cost
  and latency for trivial changes.
- `configurable tier caps`: New `tier_caps` field in the
  review matrix `limits` section allows per-tier override
  of the default agent count caps.

### Modified Capabilities
- `relevance()` output: Unchanged. Agents still go through
  relevance filtering. The tier cap is applied as a
  post-filter that limits-skips excess included agents
  after relevance decisions are made.
- `plan_review_dispatch`: Plan entries for agents excluded
  by the tier cap use a `limit_skipped` status with a
  `tier-cap` reason code, distinct from scope-miss skips.

### Removed Capabilities
- None.

## Impact

- **Files**: `.opencode/plugins/review-dispatch/index.ts`
  (dispatch logic), `.uf/review-matrix.yaml` (optional
  config), `.opencode/test/` (test updates)
- **Behavior**: Lightweight-tier changes will dispatch
  fewer agents by default. Existing standard and heavy
  behavior is unchanged. The `--full` flag bypasses tier
  caps entirely.
- **Cost**: Reduced token usage and latency for lightweight
  changes (estimated 60-70% reduction in agent invocations
  for trivial PRs).
- **Risk**: Low. The cap is a post-filter on an existing
  mechanism. Standard/heavy paths are unmodified. Full
  panel mode is unaffected.
- **Documentation**: AGENTS.md and README.md are not
  affected. CHANGELOG.md will include an entry for the
  new tier-cap feature.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

The change operates within the dispatch plugin's existing
artifact-based flow. Agent inclusion decisions remain
self-describing in the plan output (entries include
reason codes and tier-cap metadata). No runtime coupling
is introduced.

### II. Composability First

**Assessment**: PASS

Each Divisor agent remains independently usable. The
tier cap is a dispatch-level optimization that does not
alter any agent's standalone functionality. When deployed
alone, a single agent functions identically. The cap
only affects multi-agent composition and is additive
(reduces cost without removing capability).

### III. Observable Quality

**Assessment**: PASS

Tier-cap decisions are recorded in the dispatch plan
with machine-parseable reason codes (`tier-cap`) and
status (`limit_skipped`). The plan output includes the
configured cap values and which agents were excluded,
maintaining full auditability.

### IV. Testability

**Assessment**: PASS

The tier cap logic is a pure function of the change
profile tier and configured limits. It is testable in
isolation with no external dependencies. Existing test
infrastructure covers the dispatch plan output format.

### V. Security by Default

**Assessment**: PASS

Security-sensitive changes profile as heavy tier, which
has no agent cap by default. The `divisor-adversary`
agent is always included for security-sensitive profiles
regardless of tier. The always-required floor enforcement
(`max(cap, always_required_count)`) guarantees that no
configuration can drop always-required agents. When a
custom `tier_caps.heavy` is configured and the diff is
security-sensitive, the planner emits a plan-level
advisory warning to make the operator's choice observable.
The cap does not weaken security review coverage under
default configuration; custom overrides are explicitly
surfaced.
