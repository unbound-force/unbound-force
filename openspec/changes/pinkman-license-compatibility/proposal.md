## Why

Pinkman's License Classification procedure treats all
OSI-approved licenses as equivalent. GPL-3.0 and MIT
both receive an `approved` verdict, and a GPL-3.0
project can receive an `adopt` recommendation if its
health signals are strong.

This is dangerous for the Unbound Force ecosystem.
All hero repositories use Apache-2.0 (per Spec 002,
section 2.2: "Apache 2.0 is RECOMMENDED for all hero
repositories"). Adding a GPL-3.0 or AGPL-3.0
dependency would create derivative work obligations
that conflict with the Apache-2.0 license. Pinkman
should not recommend adopting dependencies whose
licenses are incompatible with the project's own
license.

## What Changes

Add a license compatibility tier to Pinkman's License
Classification procedure, layered on top of the
existing OSI-approved check:

1. **Compatibility tiers** -- classify every detected
   license as `permissive`, `weak-copyleft`, or
   `strong-copyleft` based on its derivative work
   requirements.
2. **Compatibility verdict** -- assess each project's
   license against the Unbound Force project license
   (Apache-2.0) and produce a verdict: `compatible`,
   `caution`, or `incompatible`.
3. **Recommendation gate** -- factor the compatibility
   verdict into the recommendation verdict. A
   `strong-copyleft` project MUST NOT receive `adopt`
   regardless of other health signals.
4. **Output enrichment** -- display the compatibility
   tier and verdict in all output formats (discover,
   trend, audit, report).
5. **Dewey enrichment** -- include the compatibility
   verdict in stored learnings so cross-hero
   discoverability reflects license risk.

## Capabilities

### New Capabilities
- `license-compatibility-tier`: Each detected license
  is classified into a compatibility tier (permissive,
  weak-copyleft, strong-copyleft) based on derivative
  work requirements.
- `compatibility-verdict`: Each project receives a
  compatibility assessment (compatible, caution,
  incompatible) relative to the Unbound Force
  Apache-2.0 license.
- `compatibility-gated-recommendation`: The
  recommendation verdict (adopt/evaluate/defer/avoid)
  factors in the compatibility verdict as a hard gate.

### Modified Capabilities
- `license-classification`: Extended with a
  compatibility tier step after the OSI-approved check.
  The existing OSI verdicts (approved, not_approved,
  unknown, manual_review, dual_approved) are unchanged.
- `recommendation-verdict`: The `adopt` verdict now
  requires `compatible` in addition to existing health
  criteria. The `avoid` verdict now includes
  `incompatible` license as a trigger.
- `dewey-learning-storage`: Structured prose templates
  updated to include the compatibility verdict per
  project.
- `output-formatting`: Discover/Trend result list,
  Audit result table, and Recommendation Report
  updated with compatibility tier and verdict fields.
- `fallback-license-list`: Annotated with tier
  classification for each license.

### Removed Capabilities
- None.

## Impact

- **Files affected**: `.opencode/agents/pinkman.md` and
  its scaffold copy at
  `internal/scaffold/assets/opencode/agents/pinkman.md`.
- **No hero agent modifications**: No hero agent files
  are changed. Heroes that consume Pinkman's Dewey
  learnings will see richer data (compatibility
  verdicts) through their existing `dewey_semantic_search`
  Step 0 queries.
- **No schema registry changes**: No new artifact type
  registered in `schemas/`.
- **No Go code changes**: Agent-only Markdown change.
- **Backward compatible**: All existing OSI verdicts
  are preserved. The compatibility tier is additive.
  Projects that previously received `approved` + `adopt`
  still receive `approved` if their license is
  OSI-approved. The compatibility verdict adds a
  second dimension that can downgrade the
  recommendation from `adopt` to `evaluate` or `avoid`
  when the license has copyleft obligations.
- **Website documentation gate**: Exempt -- internal
  agent behavioral instruction change with no
  user-facing CLI, workflow, or output format changes.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

Pinkman is a non-hero utility agent. The compatibility
tier enriches Pinkman's self-describing output (Markdown
reports with YAML frontmatter, Dewey learnings with
structured prose). No inter-hero communication channel
is added or modified. Heroes discover compatibility
data through their existing `dewey_semantic_search`
Step 0 queries.

### II. Composability First

**Assessment**: PASS

No hero is modified. No new dependency introduced. If
Pinkman is not installed, no hero is affected. The
compatibility tier is internal to Pinkman's
classification logic.

### III. Observable Quality

**Assessment**: PASS

The compatibility tier adds provenance to license
analysis -- each project's output now includes not just
the SPDX identifier and OSI status but also the
compatibility tier and verdict. This improves
observability of license risk. All existing
machine-parseable output is preserved.

### IV. Testability

**Assessment**: PASS

No new Go code introduced. The change modifies
behavioral instructions in a Markdown agent file. The
existing scaffold drift detection test verifies the
live agent file matches its scaffold copy, providing
automated regression coverage. The compatibility tier
classifications are deterministic (based on a static
mapping of SPDX identifiers to tiers) and verifiable
by inspection.
