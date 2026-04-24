## Context

Pinkman (Spec 032) stores scouting results in Dewey
using a flat `pinkman` tag and a brief 2-3 sentence
summary. Other heroes already query Dewey in their
Step 0 knowledge retrieval phase via
`dewey_semantic_search`. The current flat tag and terse
content limit the probability that hero queries match
Pinkman's stored learnings.

The proposal (constitution alignment: all PASS) calls
for enriching the tag taxonomy and content structure
within Pinkman's Dewey integration section. No hero
agents, schema registry, or Go code are modified.

## Goals / Non-Goals

### Goals
- Increase semantic search hit rate for Pinkman's
  learnings when heroes query Dewey in Step 0.
- Enable intentional filtered discovery via
  `dewey_semantic_search_filtered(has_tag: ...)`
  using mode-specific tags (`pinkman-discover`,
  `pinkman-audit`, etc.).
- Provide richer context in stored learnings so hero
  agents can extract actionable signals (verdicts, risk
  levels, dependency overlaps) without re-scouting.
- Maintain backward compatibility with the existing
  `dewey_store_learning` API (free-form `tag` and
  `information` strings per Spec 021).

### Non-Goals
- Registering a new artifact type in `schemas/`. The
  envelope system is not used.
- Modifying any hero agent file. Discovery is
  probabilistic via existing `dewey_semantic_search`.
- Changing the `dewey_store_learning` API or Dewey
  server behavior. Only Pinkman's usage of the API
  changes.
- Guaranteed delivery of scouting results to specific
  heroes. This is opportunistic, context-relevant
  surfacing.

## Decisions

### D1: Mode-specific tag convention

Use `pinkman-<mode>` as the tag value:

| Mode | Tag |
|------|-----|
| Discover | `pinkman-discover` |
| Trend | `pinkman-trend` |
| Audit | `pinkman-audit` |
| Report | `pinkman-report` |

**Rationale**: Spec 021 data model defines tags as
free-form strings with recommended conventions. The
`pinkman-` prefix is a new convention compatible with
Spec 021's free-form tag model (existing conventions
use branch names, dates, categories, and file paths).

Slash-separated tags (`pinkman/discover`) were
considered but rejected: `dewey_store_learning` strips
`/` from tag values, producing `pinkmandiscover`
(empirically verified 2026-04-24). Hyphens are
preserved by the API.

`dewey_find_by_tag` was considered for broad discovery
(`find_by_tag("pinkman")` with `includeChildren`) but
rejected: `find_by_tag` searches `#tag` references in
Logseq block content, not the `tag` property on stored
learnings (empirically verified 2026-04-24). The
primary discovery path for learnings is
`dewey_semantic_search` (content similarity) and
`dewey_semantic_search_filtered(has_tag: ...)` (tag
filtering on the embedding index).

Old learnings stored with the flat `pinkman` tag remain
discoverable via `dewey_semantic_search` -- no
migration needed.

### D2: Content prefix convention

Each stored learning's information string starts with a
mode-specific label:

| Mode | Prefix |
|------|--------|
| Discover | `scouting-report:` |
| Trend | `trend-report:` |
| Audit | `dependency-audit:` |
| Report | `adoption-report:` |

**Rationale**: Prefixes serve two purposes: (1) they
create a consistent text pattern that improves semantic
search relevance when a hero queries for a specific
kind of finding, and (2) they enable grep-style
substring matching in `dewey_search` for exact lookups.
This is a text convention, not a schema -- no
validation or enforcement.

### D3: Structured prose content

Replace the current "2-3 sentence summary" instruction
with mode-specific content templates that include:

- **Discover**: project names, license verdicts
  (adopt/evaluate/defer/avoid), overlapping deps
- **Trend**: project names, composite trend rank, star
  growth %, release velocity, contributor activity
- **Audit**: manifest path, dep count, deps with
  updates, deps with license changes, risk levels
  (healthy/warning/critical)
- **Report**: project URL, overall verdict, key risk
  factors, license classification

**Rationale**: Richer content increases the surface area
for semantic matching. A hero querying "testify
patterns" is more likely to match a learning that
contains "testify (MIT, adopt, 22k stars)" than one
that says "Found 3 Go testing libraries."

### D4: Keep category as `reference`

The `category` parameter remains `reference` for all
modes. The mode distinction is carried by the tag.

**Rationale**: Dewey's `category` field is used for
coarse classification (decision, pattern, gotcha,
context, reference). Scouting results are reference
material regardless of mode. Adding mode-specific
categories would fragment search results without
benefit -- the mode-specific tag already carries the
mode signal.

These fields are MUST-level per the delta spec
(dewey-integration.md, Structured Prose Content
requirement).

## Risks / Trade-offs

### R1: Probabilistic discovery (accepted)

Heroes find Pinkman's output only when their semantic
query overlaps with stored content. There is no
guarantee of delivery. This is an accepted trade-off
for zero coupling -- the alternative (envelope-based
guaranteed delivery) would require schema registry
changes and hero agent modifications.

### R2: Tag namespace collision (low risk)

If another agent or user stores learnings with a tag
starting with `pinkman-`, results would mix. Mitigated
by the `pinkman-` prefix being clearly owned by the
Pinkman agent. No other agent uses this prefix.

### R3: Content length (low risk)

Richer structured prose means longer `information`
strings passed to `dewey_store_learning`. Dewey does not
enforce a length limit on learnings (per Spec 021). The
embedding model (granite-embedding:30m) accepts
variable-length input up to its context window.
Structured prose for a single scouting session is well
within typical limits. Embedding quality may degrade
for very long learnings; content templates should aim
for conciseness.

### R4: No migration for existing flat-tagged learnings

Learnings previously stored with the flat `pinkman` tag
are not retroactively re-tagged. They remain
discoverable via `dewey_semantic_search` (content
similarity is tag-independent). This is low risk because
Pinkman is newly implemented (Spec 032) and the volume
of existing flat-tagged learnings is near zero.
