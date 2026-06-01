## Why

Two related issues exposed structural gaps in the review
infrastructure that cause security findings to be
under-classified and fragmented:

1. **Finding fragmentation** (#228): During a supply
   chain review (complytime-collector-components#229),
   the Adversary persona split one compound finding
   (unsigned binary download + privileged CI container)
   into three separate findings at LOW/MEDIUM instead
   of recognizing them as one MEDIUM-with-escalation.
   Root causes: no dependency necessity check, no
   compound severity rule, no finding consolidation
   pass, CI bot annotations dismissed as unrelated,
   no cross-persona consolidation in `/review-council`.

2. **Severity anchoring bias** (#233): During a
   different review (complytime/org-infra#285), three
   HIGH findings were initially classified as MEDIUM
   because the reviewer anchored on a positive first
   impression. Root causes: no severity calibration
   step forcing re-read of definitions, no adversarial
   input enumeration, no linked-issue suggestion gap
   detection.

Both issues share a common theme: the security
*detection* infrastructure is comprehensive, but the
*synthesis* layer is missing. Agents find issues
individually but lack reasoning steps to consolidate,
calibrate, and cross-reference.

Additionally, the org constitution has no explicit
security principle. Security is enforced indirectly
through gatekeeping integrity and testability, leaving
supply chain integrity and input validation subject to
per-finding severity judgment rather than constitutional
authority.

## What Changes

### New Capabilities

- **Compound Severity Escalation**: New section in
  `severity.md` defining when individually-moderate
  findings compose into higher-severity findings.
- **Severity Calibration Step**: New post-Step-8 pass
  in `/review-pr` forcing re-read of `severity.md`
  definitions with quoted evidence for each finding.
- **Adversarial Input Enumeration**: New substep in
  `/review-pr` Step 8b and `divisor-adversary.md`
  requiring per-input threat analysis for each new
  input/parameter introduced by the PR.
- **Constitution Principle V: Security by Default**:
  New org-level principle elevating supply chain
  integrity, input validation, and least privilege
  to constitutional status.

### Modified Capabilities

- **`divisor-adversary.md`**: Audit Checklist item 2
  adds dependency necessity check, content-integrity
  verification guidance, CI bot corroboration
  instruction, and adversarial input enumeration.
- **`severity.md`**: Adds Compound Severity Escalation
  section with escalation rule, example patterns, and
  anti-consolidation guard.
- **`/review-pr`**: Adds Steps 8e (CI bot annotation
  cross-referencing), 8f (finding consolidation),
  8g (severity calibration), and 8b substep
  (adversarial input enumeration).
- **`/review-council`**: Adds cross-persona finding
  consolidation in Code Review Mode Step 3 and Spec
  Review Mode Step 2.

### Removed Capabilities

- None.

## Impact

- `.opencode/agents/divisor-adversary.md` — expanded
  Audit Checklist items 2 and new item for adversarial
  input enumeration
- `.opencode/uf/packs/severity.md` — new Compound
  Severity Escalation section
- `.opencode/commands/review-pr.md` — new Steps
  8e, 8f, 8g, and 8b substep
- `.opencode/commands/review-council.md` — consolidation
  pass in Code Review Step 3 and Spec Review Step 2
- `.specify/memory/constitution.md` — new Principle V

## Constitution Alignment

### I. Autonomous Collaboration

**Assessment**: PASS

Finding consolidation and severity calibration operate
within each agent's own analysis pass. Cross-persona
consolidation in `/review-council` merges findings
from artifacts already produced by independent agents
— no runtime coupling is introduced.

### II. Composability First

**Assessment**: PASS

All changes are additive to existing agent files and
convention packs. No new mandatory dependencies are
introduced. The severity calibration step references
`severity.md` which is already a required dependency
for all Divisor personas.

### III. Observable Quality

**Assessment**: PASS

Consolidated findings preserve per-persona attribution
and cite contributing factors. Severity calibration
requires quoting the severity definition that matches
each finding, making classification decisions auditable.

### IV. Testability

**Assessment**: N/A

These changes modify agent instruction files and
convention packs (Markdown), not source code. No
coverage strategy is needed. Verification is via
review-based testing (run `/review-council` on a
known PR and check finding quality).
