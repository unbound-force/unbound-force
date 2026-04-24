## ADDED Requirements

### Requirement: Mode-Specific Tag Taxonomy

Pinkman MUST use mode-specific tags when storing
learnings via `dewey_store_learning`:

| Mode | Tag Value |
|------|-----------|
| Discover | `pinkman-discover` |
| Trend | `pinkman-trend` |
| Audit | `pinkman-audit` |
| Report | `pinkman-report` |

Tags use hyphen separation (not slash) because
`dewey_store_learning` strips `/` from tag values.

The primary discovery path for stored learnings is
`dewey_semantic_search` (content similarity). Tags
serve as filters via
`dewey_semantic_search_filtered(has_tag: ...)`.
`dewey_find_by_tag` does not search the `tag` property
on stored learnings and MUST NOT be relied upon for
learning discovery.

#### Scenario: Discover mode stores with mode-specific tag
- **GIVEN** Pinkman completes a discover mode scouting
  session
- **WHEN** Pinkman stores the learning via
  `dewey_store_learning`
- **THEN** the `tag` parameter MUST be `pinkman-discover`

#### Scenario: Trend mode stores with mode-specific tag
- **GIVEN** Pinkman completes a trend mode session
- **WHEN** Pinkman stores the learning via
  `dewey_store_learning`
- **THEN** the `tag` parameter MUST be `pinkman-trend`

#### Scenario: Audit mode stores with mode-specific tag
- **GIVEN** Pinkman completes an audit mode session
- **WHEN** Pinkman stores the learning via
  `dewey_store_learning`
- **THEN** the `tag` parameter MUST be `pinkman-audit`

#### Scenario: Report mode stores with mode-specific tag
- **GIVEN** Pinkman completes a report mode session
- **WHEN** Pinkman stores the learning via
  `dewey_store_learning`
- **THEN** the `tag` parameter MUST be `pinkman-report`

#### Scenario: Filtered search finds mode-specific learnings
- **GIVEN** Pinkman has stored learnings with tags
  `pinkman-discover`, `pinkman-audit`, and
  `pinkman-report`
- **WHEN** a consumer calls
  `dewey_semantic_search_filtered(query: "...",
  has_tag: "pinkman-audit")`
- **THEN** only the audit-tagged learning MUST be
  returned

#### Scenario: Semantic search finds all learnings
- **GIVEN** Pinkman has stored learnings with tags
  `pinkman-discover`, `pinkman-audit`, and
  `pinkman-report`
- **WHEN** a consumer calls `dewey_semantic_search`
  with a query matching all three learnings' content
- **THEN** all three learnings SHOULD be returned
  (ranked by content similarity)

### Requirement: Content Prefix Convention

Pinkman MUST prefix the `information` string with a
mode-specific label:

| Mode | Prefix |
|------|--------|
| Discover | `scouting-report:` |
| Trend | `trend-report:` |
| Audit | `dependency-audit:` |
| Report | `adoption-report:` |

#### Scenario: Discover learning has correct prefix
- **GIVEN** Pinkman completes a discover mode session
- **WHEN** Pinkman stores the learning via
  `dewey_store_learning`
- **THEN** the `information` string MUST begin with
  `scouting-report:`

#### Scenario: Trend learning has correct prefix
- **GIVEN** Pinkman completes a trend mode session
- **WHEN** Pinkman stores the learning via
  `dewey_store_learning`
- **THEN** the `information` string MUST begin with
  `trend-report:`

#### Scenario: Audit learning has correct prefix
- **GIVEN** Pinkman completes an audit mode session
- **WHEN** Pinkman stores the learning via
  `dewey_store_learning`
- **THEN** the `information` string MUST begin with
  `dependency-audit:`

#### Scenario: Report learning has correct prefix
- **GIVEN** Pinkman completes a report mode session
- **WHEN** Pinkman stores the learning via
  `dewey_store_learning`
- **THEN** the `information` string MUST begin with
  `adoption-report:`

### Requirement: Structured Prose Content

Pinkman MUST include the following structured data in
stored learnings, based on the scouting mode:

**Discover mode**: project names, license verdicts
(adopt/evaluate/defer/avoid), query used, overlapping
dependencies if detected.

**Trend mode**: project names, composite trend rank,
star growth percentage, release velocity, contributor
activity metrics.

**Audit mode**: manifest path, total dependency count,
dependencies with available updates, dependencies with
license changes, risk levels (healthy/warning/critical).

**Report mode**: project URL, overall adoption verdict,
key risk factors, license classification.

#### Scenario: Discover learning contains structured data
- **GIVEN** Pinkman discovers 3 projects in discover mode
- **WHEN** Pinkman stores the learning
- **THEN** the information string MUST include each
  project's name, license verdict, and the original query
- **AND** the information string SHOULD include
  overlapping dependencies if any were detected

#### Scenario: Audit learning contains risk levels
- **GIVEN** Pinkman audits a go.mod with 2 critical-risk
  dependencies
- **WHEN** Pinkman stores the learning
- **THEN** the information string MUST include the risk
  level for each flagged dependency

### Requirement: API Parameter Reference

The `dewey_store_learning` tool accepts `information`
(required string), `tag` (required string), and
`category` (optional string). Note: Spec 021 data model
documents the input as `{ text, tags }` (array), but
the live MCP tool uses `information` (not `text`) and
`tag` (singular string, not `tags` array). The live
tool interface is authoritative.

## MODIFIED Requirements

### Requirement: Dewey After Scouting (Spec 032)

Previously: "Store a condensed summary via
`dewey_store_learning`: tag: `pinkman`, category:
`reference`, information: A 2-3 sentence summary
including project names, license verdicts, key metrics,
and the query used."

Updated: The tag, category, and information parameters
MUST follow the mode-specific tag taxonomy, content
prefix convention, and structured prose content
requirements defined above. The category MUST remain
`reference`.

## REMOVED Requirements

None.
