## Why

Pinkman stores scouting results in Dewey via
`dewey_store_learning` using a flat `pinkman` tag and
generic `reference` category. The stored information is
a brief 2-3 sentence summary. This limits cross-hero
discoverability: other heroes' Step 0 Dewey queries
only surface Pinkman's findings when the query happens
to semantically overlap with the terse summary text.

Enriching the tag taxonomy and content structure of
Pinkman's Dewey learnings improves the probability that
hero agents discover relevant scouting insights during
their existing Step 0 knowledge retrieval -- without
modifying any hero agent file.

## What Changes

Modify Pinkman's `dewey_store_learning` instructions in
the agent file to use:

1. **Mode-specific tags** -- `pinkman-discover`,
   `pinkman-trend`, `pinkman-audit`, `pinkman-report`
   instead of flat `pinkman`. Hyphen-separated because
   `dewey_store_learning` strips `/` from tag values.
2. **Content prefixes** -- mode-specific labels
   (`scouting-report:`, `trend-report:`,
   `dependency-audit:`, `adoption-report:`) at the start
   of the information string.
3. **Structured prose** -- richer content including
   project names, license verdicts, risk levels,
   dependency overlaps, trend metrics, and adoption
   recommendations, enabling more precise semantic
   matching.

## Capabilities

### New Capabilities
- `mode-specific-tag-taxonomy`: Pinkman learnings use
  mode-specific tags under the `pinkman-` prefix
  (`pinkman-discover`, `pinkman-audit`, etc.), enabling
  filtered discovery via
  `dewey_semantic_search_filtered(has_tag: ...)`.
- `structured-content-convention`: Information strings
  follow a prefix + structured prose convention that
  increases semantic search hit rates for domain-relevant
  hero queries.

### Modified Capabilities
- `dewey-learning-storage`: The existing Dewey
  integration section in pinkman.md is updated with
  richer tag/content conventions. No functional change
  to the storage mechanism itself.

### Removed Capabilities
- None.

## Impact

- **Files affected**: `.opencode/agents/pinkman.md` and
  its scaffold copy at
  `internal/scaffold/assets/opencode/agents/pinkman.md`.
- **No hero agent modifications**: Cobalt-Crush,
  Muti-Mind, Mx F, Gaze, and all Divisor agents are
  unchanged. Their existing Step 0 `dewey_semantic_search`
  queries will surface enriched Pinkman learnings through
  improved semantic similarity -- no explicit wiring
  needed.
- **No schema registry changes**: No new artifact type
  registered in `schemas/`. The envelope format is not
  used.
- **No Go code changes**: Agent-only Markdown change.
- **Backward compatible**: The `dewey_store_learning` API
  accepts any string for `tag` and `information`. The
  mode-specific tags (`pinkman-discover`, etc.) are valid
  free-form strings per Spec 021 conventions. Old
  learnings stored with flat `pinkman` remain discoverable
  via `dewey_semantic_search`.
- **Website documentation gate**: Exempt -- internal agent
  behavioral instruction change with no user-facing CLI,
  workflow, or output format changes.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

Pinkman is a non-hero utility agent. The inter-hero
artifact envelope mandate (line 51-53) applies to heroes,
not utility agents. Dewey learnings are a shared knowledge
layer, not a direct inter-hero communication channel.
Heroes discover Pinkman's findings through their existing
`dewey_semantic_search` Step 0 queries -- no synchronous
interaction, no runtime coupling, no ad-hoc exchange
format. The learning content is self-describing (includes
producer context via tag prefix, mode, and structured
result data).

### II. Composability First

**Assessment**: PASS

No hero is modified. No new dependency introduced. If
Pinkman is not installed, no hero is affected. If Dewey
is unavailable, Pinkman still writes Markdown reports to
`.uf/pinkman/reports/` (graceful degradation already
implemented in Spec 032). The enrichment is purely
additive.

### III. Observable Quality

**Assessment**: PASS

Pinkman's existing Markdown reports with YAML frontmatter
provenance metadata are unchanged. The Dewey learnings
gain richer content (project names, verdicts, metrics)
that improves observability of scouting results across
sessions. No machine-parseable output is removed.

### IV. Testability

**Assessment**: PASS

No new Go code introduced. The change modifies behavioral
instructions in a Markdown agent file. The existing
scaffold drift detection test verifies the live agent
file matches its scaffold copy, providing automated
regression coverage. No external services or shared
mutable state involved.
