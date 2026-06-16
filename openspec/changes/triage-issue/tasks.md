<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. JSON Schema

Create the `issue-triage` schema, samples, and
documentation. All files in `schemas/issue-triage/`
are independent of each other and touch different
files.

- [x] 1.1 [P] Create `schemas/issue-triage/v1.0.0.schema.json`
  Define the issue-triage payload schema following
  JSON Schema draft 2020-12. Include all fields from
  the design (D7): `issue_number`, `issue_url`, `repo`,
  `title`, `author`, `category` (enum), `validity`
  (enum), `objectivity` (enum), `assessments` (array
  of agent results), `duplicate_of` (integer|null),
  `split_issues` (array), `actions_taken` (object),
  `summary` (object). Set `additionalProperties: false`
  at every level. Use `$id` following the existing
  pattern: `https://github.com/unbound-force/unbound-force/internal/schemas/issue-triage-payload`.
  Refs: FR-027, FR-028, FR-029, FR-030.

- [x] 1.2 [P] Create `schemas/issue-triage/README.md`
  Follow the `schemas/feedback-triage/README.md`
  pattern. Producer: The Divisor (panel) via
  `/triage-issue`. Consumers: Mx F (triage pattern
  trends), Muti-Mind (backlog health), Cobalt-Crush
  (learns from past triage). Include required/optional
  field tables, schema evolution policy, and version
  history.

- [x] 1.3 [P] Create `schemas/issue-triage/samples/sample-issue-triage.json`
  Valid sample showing a bug triage with five agent
  assessments, one dissenting, label applied, comment
  posted, no split. Must validate against the schema
  from 1.1.

- [x] 1.4 [P] Create `schemas/issue-triage/samples/invalid-extra-property.json`
  Invalid sample with an extra `priority` property on
  an assessment item. Must fail validation due to
  `additionalProperties: false`.

- [x] 1.5 [P] Create `schemas/issue-triage/samples/invalid-missing-required.json`
  Invalid sample with a required field omitted (e.g.,
  missing `assessments`). Must fail validation due to
  missing required field.

- [x] 1.6 [P] Create `schemas/issue-triage/samples/invalid-bad-enum.json`
  Invalid sample with an invalid enum value (e.g.,
  `category: "urgent"`). Must fail validation due to
  enum constraint violation.

- [x] 1.7 Add `"issue-triage"` to `handAuthoredSchemas`
  in `internal/schemas/ci_test.go`. This enables
  automated CI validation of all samples (positive
  and negative) against the schema. Single-line change.
  Refs: FR-029.

## 2. Slash Command

Create the `/triage-issue` command file. This is a
single file with no parallel opportunities.

- [x] 2.1 Create `.opencode/commands/triage-issue.md`
  Implement the four-phase pipeline (Ingest, Assess,
  Classify, Act) as described in the design. Include:
  - YAML frontmatter with description
  - Arguments section (issue number, required)
  - Prerequisites (gh CLI, auth, repo detection)
  - Phase 1: Ingest
    - Input validation (FR-001)
    - Fetch issue, validate open state (FR-002)
    - Repo detection (FR-003)
    - Duplicate check with keyword sanitization (FR-012)
    - gh CLI verification (FR-035)
  - Phase 2: Assess
    - Dynamic agent discovery (FR-006)
    - Parallel fan-out to 5 Divisor agents (FR-004)
    - Structured assessment return format (FR-005)
    - Graceful degradation (FR-007)
  - Phase 3: Classify
    - Majority verdict 3/5 with NEEDS-CLARIFICATION
      and tie-breaking rules (FR-008)
    - Category specificity hierarchy (FR-009)
    - Objectivity classification (FR-010)
    - Dissent recording (FR-011)
    - Duplicate resolution (FR-013)
    - Split synthesis (FR-022)
  - Phase 4: Act
    - Present analysis to user
    - Duplicate pre-check for child issues (FR-014)
    - Label application without confirmation,
      except `duplicate` (FR-015, FR-016)
    - Comment posting with confirmation via
      temp file + --input (FR-017, FR-018,
      FR-019, FR-020, FR-021)
    - Child issue creation with confirmation
      and shell injection prevention (FR-022,
      FR-023, FR-024, FR-025, FR-026)
    - Artifact production (FR-027, FR-028,
      FR-029, FR-030)
  - Comment tone tiers (valid, invalid/opinion,
    needs-clarification) per design D6.
  - Guardrails section covering:
    - No auto-close (FR-031)
    - No comments without confirmation (FR-032)
    - No child issues without confirmation (FR-033)
    - Single issue per invocation (FR-034)
    - gh CLI verification (FR-035)
    - API failure handling (FR-036)
    - Idempotent re-run (FR-037)
    - Shell injection prevention (FR-038)
    - Safe artifact paths (FR-039)

## 3. Scaffold Asset

- [x] 3.1 [P] Copy `.opencode/commands/triage-issue.md`
  to `internal/scaffold/assets/opencode/commands/triage-issue.md`
  so the command is deployed by `uf init`.

- [x] 3.2 [P] Update `expectedAssetPaths` in
  `internal/scaffold/scaffold_test.go` to include
  the new command file. This ensures scaffold drift
  tests detect if the embedded copy diverges from
  the canonical source.

## 4. Documentation

- [x] 4.1 [P] Update AGENTS.md project structure to
  list `schemas/issue-triage/` alongside
  `feedback-triage/`. Add `/triage-issue` to an
  "Issue Triage Commands" reference section.

- [x] 4.2 [P] Add CHANGELOG.md entry for the new
  `/triage-issue` command and `issue-triage` schema.

- [x] 4.3 File a `docs`-labeled issue in
  `unbound-force/website` for the new `/triage-issue`
  command documentation. Include coverage for:
  command usage, examples, the issue-triage artifact
  format, and The Divisor team page update.
  Constitution requirement: Cross-Repo Documentation.

- [x] 4.4 File a `blog`-labeled issue in
  `unbound-force/website` for a blog post covering
  the multi-agent issue triage capability.

## 5. Verification

- [x] 5.1 Run `make check` to verify build, test,
  vet, and lint all pass with the new schema and
  scaffold asset.

- [x] 5.2 Validate schema samples by running
  `go test -race -count=1 ./internal/schemas/...`
  to confirm the valid sample passes and all three
  invalid samples are rejected.

- [x] 5.3 Verify constitution alignment
  Confirm the implementation aligns with all five
  constitution principles:
  - I. Autonomous Collaboration: artifact produced with
    envelope, no synchronous inter-agent coupling
  - II. Composability First: graceful degradation with
    missing agents, no hero prerequisites
  - III. Observable Quality: JSON schema, provenance
    metadata, per-agent assessment provenance
  - IV. Testability: valid/invalid samples with CI
    regression testing via `handAuthoredSchemas`
  - V. Security by Default: shell injection prevention
    via temp files, input validation, keyword
    sanitization, dynamic repo detection

<!-- spec-review: passed -->
<!-- code-review: passed -->
