## ADDED Requirements

### Requirement: Idempotent marker insertion

`insertMarkerAfterFrontmatter` MUST produce output
containing exactly one provenance marker line, regardless
of how many marker lines the input content contains.

The function MUST strip all existing marker lines before
inserting the new one. Marker lines are identified by the
prefixes `<!-- scaffolded by uf ` (HTML comment) and
`# scaffolded by uf ` (hash comment).

#### Scenario: Input with no existing markers

- **GIVEN** content with no scaffold provenance markers
- **WHEN** `insertMarkerAfterFrontmatter` is called
- **THEN** the output contains exactly one marker line
  after the YAML frontmatter closing delimiter (or at end
  of file if no frontmatter)

#### Scenario: Input with one existing marker

- **GIVEN** content with one scaffold provenance marker
- **WHEN** `insertMarkerAfterFrontmatter` is called
- **THEN** the output contains exactly one marker line
  (the old one is replaced)

#### Scenario: Input with multiple existing markers

- **GIVEN** content with three or more scaffold markers
- **WHEN** `insertMarkerAfterFrontmatter` is called
- **THEN** the output contains exactly one marker line
  and all original content (non-marker lines) is preserved

#### Scenario: Repeated calls produce stable output

- **GIVEN** content that has already been processed by
  `insertMarkerAfterFrontmatter`
- **WHEN** `insertMarkerAfterFrontmatter` is called again
  with the same version marker
- **THEN** the output is byte-identical to the input

### Requirement: Version marker correctness

`versionMarker` MUST produce markers with a single `v`
prefix followed by the semantic version (e.g.,
`v0.6.1`). Release builds MUST NOT produce double-v
prefixes (e.g., `vv0.6.1`).

#### Scenario: Release build version marker

- **GIVEN** GoReleaser injects the version via ldflags
- **WHEN** `versionMarker` formats the provenance marker
- **THEN** the marker reads `<!-- scaffolded by uf vX.Y.Z -->`
  with exactly one `v` prefix

#### Scenario: Development build version marker

- **GIVEN** the version variable defaults to `"dev"`
- **WHEN** `versionMarker` formats the provenance marker
- **THEN** the marker reads `<!-- scaffolded by uf vdev -->`

### Requirement: Embedded asset marker regression guard

Automated tests MUST verify that no embedded asset under
`internal/scaffold/assets/` contains more than one
scaffold provenance marker line.

#### Scenario: Drift test catches marker accumulation

- **GIVEN** an embedded asset file is modified to contain
  two or more scaffold marker lines
- **WHEN** the test suite runs
- **THEN** the test fails with a message identifying the
  file and the number of markers found

## MODIFIED Requirements

### Requirement: Tool-owned file update semantics

Previously: `Run()` achieves idempotency via the
`bytes.Equal` check for tool-owned files, compensating
for the non-idempotent `insertMarkerAfterFrontmatter`.

Now: `insertMarkerAfterFrontmatter` is itself idempotent.
The `bytes.Equal` check in `Run()` remains as a
performance optimization (avoids unnecessary disk writes)
but is no longer the sole defense against marker
accumulation.

## REMOVED Requirements

None.
