## ADDED Requirements

### Requirement: Implement-specific guardrails

Step 6 MUST inject a command-specific guardrail block for
`speckit.implement.md` that:
- States this command writes source code as its primary
  purpose
- Restricts modifications to the active feature's
  implementation scope
- MUST NOT contain "NEVER modify source code"

#### Scenario: Fresh inject on implement command
- **GIVEN** `speckit.implement.md` has no `## Guardrails`
  section
- **WHEN** `/uf.init` Step 6 runs
- **THEN** the implement-specific guardrail block is
  appended, and the report shows
  `✅ speckit.implement.md: guardrails injected`

#### Scenario: Replace incorrect guardrails on implement
- **GIVEN** `speckit.implement.md` has a `## Guardrails`
  section containing "NEVER modify source code"
- **WHEN** `/uf.init` Step 6 runs
- **THEN** the entire `## Guardrails` section is replaced
  with the implement-specific variant, and the report
  shows `✅ speckit.implement.md: guardrails corrected`

### Requirement: Constitution-specific guardrails

Step 6 MUST inject a command-specific guardrail block for
`speckit.constitution.md` that:
- Lists `.specify/memory/constitution.md` and
  `.specify/templates/*-template.md` as allowed write
  targets (intentionally broad to cover template additions)
- MUST NOT restrict writes to `FEATURE_SPEC` or
  `FEATURE_DIR`
- MUST NOT contain "NEVER modify source code"

#### Scenario: Fresh inject on constitution command
- **GIVEN** `speckit.constitution.md` has no `## Guardrails`
  section
- **WHEN** `/uf.init` Step 6 runs
- **THEN** the constitution-specific guardrail block is
  appended, and the report shows
  `✅ speckit.constitution.md: guardrails injected`

#### Scenario: Replace incorrect guardrails on constitution
- **GIVEN** `speckit.constitution.md` has a `## Guardrails`
  section containing "ONLY files this command may write"
  with `FEATURE_SPEC`
- **WHEN** `/uf.init` Step 6 runs
- **THEN** the entire `## Guardrails` section is replaced
  with the constitution-specific variant, and the report
  shows `✅ speckit.constitution.md: guardrails corrected`

### Requirement: Taskstoissues-specific guardrails

Step 6 MUST inject a command-specific guardrail block for
`speckit.taskstoissues.md` that:
- States this command creates GitHub issues via MCP API
- States issues MUST only be created in the repository
  matching the current Git remote
- States this command does NOT write local files
- MUST NOT contain "files this command may write"

#### Scenario: Fresh inject on taskstoissues command
- **GIVEN** `speckit.taskstoissues.md` has no `## Guardrails`
  section
- **WHEN** `/uf.init` Step 6 runs
- **THEN** the taskstoissues-specific guardrail block is
  appended, and the report shows
  `✅ speckit.taskstoissues.md: guardrails injected`

#### Scenario: Replace incorrect guardrails on taskstoissues
- **GIVEN** `speckit.taskstoissues.md` has a `## Guardrails`
  section containing "ONLY files this command may write"
- **WHEN** `/uf.init` Step 6 runs
- **THEN** the entire `## Guardrails` section is replaced
  with the taskstoissues-specific variant, and the report
  shows `✅ speckit.taskstoissues.md: guardrails corrected`

### Requirement: Correctness marker detection

Step 6 MUST use a correctness marker for each
execution/utility command to detect whether existing
guardrails are correct or incorrect:

| Command | Correctness marker |
|---------|-------------------|
| `speckit.implement.md` | "writes source code" |
| `speckit.constitution.md` | ".specify/memory/" |
| `speckit.taskstoissues.md` | "GitHub issues via" |

If `## Guardrails` exists but the correctness marker is
absent, the guardrails MUST be replaced.

**Known limitation**: The correctness marker is a positive
substring check only. The regression tests (requirement
below) provide negative checks (e.g., implement MUST NOT
contain "NEVER modify source code") that catch
contradictory content at test time, even if the marker is
present. This two-layer defense (runtime marker + test-time
negative assertions) is acceptable for this change.

**Heading detection**: The `## Guardrails` check MUST match
a markdown heading outside of fenced code blocks. This is
important because `uf.init.md` itself contains `## Guardrails`
inside template code fences.

#### Scenario: Correct guardrails already present
- **GIVEN** `speckit.implement.md` has a `## Guardrails`
  section containing "writes source code"
- **WHEN** `/uf.init` Step 6 runs
- **THEN** the guardrails are not modified, and the report
  shows
  `⊘ speckit.implement.md: guardrails already present (skipped)`

### Requirement: Guardrail content regression tests

The test suite MUST include assertions verifying the
guardrail templates in `uf.init.md` contain correct content
for each command type. These tests MUST fail if incorrect
restrictions are reintroduced.

#### Scenario: Regression test catches reintroduced error
- **GIVEN** someone modifies the implement guardrail in
  `uf.init.md` to include "NEVER modify source code"
- **WHEN** `go test ./...` runs
- **THEN** the guardrail content regression test fails

## MODIFIED Requirements

### Requirement: Step 6 command categorization

Previously: Step 6 defined two categories — spec-phase
(6 commands) and execution/utility (3 commands) — using
two guardrail templates that differed only by the
review-rationale sentence.

Now: Step 6 defines five categories:
1. Spec-phase (6 commands): unchanged guardrails with
   review-rationale
2. `speckit.implement.md`: command-specific guardrails
3. `speckit.constitution.md`: command-specific guardrails
4. `speckit.taskstoissues.md`: command-specific guardrails
5. (The shared execution/utility variant is removed)

### Requirement: Dual-file lockstep update

Both `.opencode/commands/uf.init.md` and
`internal/scaffold/assets/opencode/commands/uf.init.md`
MUST be updated with identical content.

Previously: This was an implicit requirement enforced by
`TestEmbeddedAssets_MatchSource`.

Now: Explicitly stated. The test continues to enforce it.

## REMOVED Requirements

### Requirement: Workaround Note for implement exception

The Note at lines 627–632 documenting that
`speckit.implement.md` is an exception and that "the
implement command's own instructions override the guardrails
where they conflict" is removed. The fix eliminates the
conflict, making the workaround unnecessary. Retaining it
would violate the zero-waste principle.
