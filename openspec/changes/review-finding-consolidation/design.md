## Context

The `/review-pr` and `/review-council` commands are the
primary code review workflows. `/review-pr` is a
single-pass AI review with CI integration; `/review-council`
orchestrates multiple Divisor personas in parallel.

Both workflows generate findings per-category or
per-persona, then present them without synthesis.
This causes three classes of error:

1. Related findings split across categories/personas
   at inconsistent severities (fragmentation)
2. Severity levels anchored on first impressions
   rather than rigorously calibrated against
   definitions (anchoring bias)
3. CI bot findings dismissed as "related but different"
   instead of used as corroborating evidence

The Adversary persona checks security categories but
does not enumerate inputs adversarially — it asks
"are there injection risks?" rather than "for each
new input, what can a malicious caller pass?"

The org constitution governs quality through four
principles but has no security-specific principle,
leaving security enforcement to per-finding judgment.

## Goals / Non-Goals

### Goals

- Add synthesis steps that consolidate, calibrate,
  and cross-reference findings before output
- Add adversarial input enumeration to catch
  input-specific threats the category checklist misses
- Add compound severity escalation to `severity.md`
- Add Principle V (Security by Default) to the
  org constitution
- Extend PR #229's existing changes with #233's
  severity calibration and input enumeration gaps

### Non-Goals

- No programmatic enforcement (these are agent
  instructions, not code)
- No changes to the CI security pipeline (OSV-Scanner,
  Trivy, Scorecards remain as-is)
- No changes to the feedback learning loop (#207)
- No Gemara integration (deferred to separate work)
- No changes to the Divisor personas other than
  Adversary

## Decisions

### D1: Severity calibration as a separate step

Add Step 8g (Severity Calibration) to `/review-pr`
as a post-generation pass rather than inlining
calibration into each category step. This ensures
all findings are calibrated in one pass with
consistent reference to `severity.md` definitions.

The calibration step requires: for each finding,
quote the severity definition that matches and
confirm or adjust the assignment. This counters
anchoring bias by forcing a second look.

### D2: Finding consolidation after CI cross-ref

The ordering is: 8e (CI bot cross-ref) → 8f
(consolidation) → 8g (calibration). This ensures
bot evidence is available for consolidation, and
calibration happens last on the final set of
(potentially consolidated) findings.

### D3: Adversarial input enumeration in both files

Add the enumeration substep to both
`divisor-adversary.md` (the persona definition) and
`/review-pr` Step 8b (the command workflow). The
persona definition is authoritative for `/review-council`
(which delegates to the persona), while the command
workflow is authoritative for standalone `/review-pr`
usage. Both must be consistent.

### D4: Constitution Principle V placement

Add as Principle V after Testability. The principle
covers supply chain integrity, input validation, and
least privilege. These are foundational security
properties that apply to all heroes, not just the
review system.

The version bump is MINOR (1.1.0 → 1.2.0) since
this adds a new principle without altering existing
ones. The Sync Impact Report at the top of the
constitution file must be updated.

### D5: Cross-persona consolidation scope

`/review-council` consolidation applies only at the
aggregation step (after all personas return findings),
not during individual persona execution. Each persona
still operates independently — consolidation is a
post-processing step that merges findings sharing a
root cause across personas.

### D6: Anti-consolidation guard

Both `severity.md` and the consolidation steps
include an explicit anti-consolidation rule: findings
with independent root causes and independent blast
radii MUST remain separate even if they appear in the
same file or PR. This prevents artificial inflation of
severity through unrelated-finding grouping.

## Risks / Trade-offs

### Risk: Over-consolidation

Agents may consolidate unrelated findings that happen
to touch the same file. Mitigated by the
anti-consolidation guard (D6) and the three-part
test: same component, shared root cause, combined risk
greater than individual.

### Risk: Calibration step adds review latency

An additional pass over all findings adds token cost
and time. Trade-off is acceptable because the
alternative (under-classified findings that miss real
security issues) has a higher cost.

### Risk: Constitution amendment process

Adding Principle V requires hero constitution alignment
review. Mitigated by the constitution's own amendment
process — hero repos must open alignment issues within
one release cycle. The principle is additive (no
existing principles change), minimizing conflict risk.

### Trade-off: Duplicate enumeration instructions

The adversarial input enumeration appears in both
`divisor-adversary.md` and `review-pr.md`. This is
intentional (D3) but creates a maintenance surface
for keeping them in sync.
