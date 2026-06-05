## ADDED Requirements

### Requirement: FR-001 Concurrent Sub-Tool Initialization

`initSubTools()` MUST run independent sub-tool initializations
concurrently. Tools with no interdependencies MUST NOT block
each other.

#### Scenario: Independent tools run concurrently
- **GIVEN** dewey, replicator, specify, openspec, and gaze
  are all available on PATH
- **WHEN** `initSubTools()` is called
- **THEN** replicator, specify, openspec, and gaze init
  commands start without waiting for dewey indexing to complete

#### Scenario: Missing tool does not block others
- **GIVEN** dewey is not available on PATH but replicator,
  specify, openspec, and gaze are
- **WHEN** `initSubTools()` is called
- **THEN** all available tools complete their init without
  delay from the missing tool

### Requirement: FR-002 Dewey Internal Sequencing

Within the Dewey initialization group, operations MUST execute
in order: `dewey init` -> `generateDeweySources` ->
`dewey index`. These three steps MUST NOT be parallelized
with each other.

#### Scenario: Dewey init precedes indexing
- **GIVEN** dewey is available on PATH and `.uf/dewey/` does
  not exist
- **WHEN** `initSubTools()` is called
- **THEN** `dewey init` completes before `generateDeweySources`
  runs, and `generateDeweySources` completes before
  `dewey index` runs

### Requirement: FR-003 Result Aggregation

`initSubTools()` MUST return `[]subToolResult` containing
results from all attempted tool initializations regardless
of execution order. The returned slice MUST include the same
entries as the sequential implementation.

#### Scenario: All results collected
- **GIVEN** all five tools are available on PATH
- **WHEN** `initSubTools()` completes
- **THEN** the returned `[]subToolResult` contains entries
  for all tools that produced results

### Requirement: FR-004 Post-Completion Configuration

`configureOpencodeJSON()` MUST NOT execute until all concurrent
sub-tool initializations have completed.

#### Scenario: OpenCode config runs after all tools
- **GIVEN** all tools are available on PATH
- **WHEN** `initSubTools()` is called
- **THEN** `configureOpencodeJSON()` runs only after all
  concurrent groups have finished

### Requirement: FR-005 Non-Fatal Failures

Individual sub-tool failures MUST NOT prevent other tools from
completing their initialization. A failure in one tool MUST NOT
cancel or abort other concurrent tool initializations.

#### Scenario: One tool fails, others succeed
- **GIVEN** dewey is available but `dewey init` returns an error
- **WHEN** `initSubTools()` is called
- **THEN** replicator, specify, openspec, and gaze still
  complete their initialization normally

### Requirement: FR-006 Config-Based Tool Skipping

`initSubTools()` MUST respect `setup.tools.<name>.method: skip`
in `.uf/config.yaml`. A tool configured with `method: skip`
MUST NOT be initialized, even if the binary is available on
PATH and `--force` is set.

#### Scenario: Skipped tool is not initialized
- **GIVEN** dewey is available on PATH and `.uf/config.yaml`
  contains `setup.tools.dewey.method: skip`
- **WHEN** `initSubTools()` is called
- **THEN** no dewey commands are executed and no dewey-related
  results appear in the returned slice

#### Scenario: Skip overrides force
- **GIVEN** dewey is available on PATH, `.uf/dewey/` exists,
  `--force` is set, and `.uf/config.yaml` contains
  `setup.tools.dewey.method: skip`
- **WHEN** `initSubTools()` is called
- **THEN** dewey is not re-indexed

#### Scenario: All tools skipped
- **GIVEN** all five tools are configured with `method: skip`
- **WHEN** `initSubTools()` is called
- **THEN** no tool commands are executed and only
  `configureOpencodeJSON` results appear

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
