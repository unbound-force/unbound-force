# Feature Specification: Doctor & Setup Dewey Alignment

**Feature Branch**: `023-doctor-setup-dewey`  
**Created**: 2026-04-03  
**Status**: Planned  
**Input**: User description: "issue #77 — Update uf doctor and uf setup for Dewey unified memory (Spec 021 FR-019–FR-021)"

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Dewey Embedding Health Check (Priority: P1)

When an engineer runs `uf doctor`, the Dewey health check
group verifies not only that Dewey is installed and the
embedding model is present, but also that Dewey can
actually generate embeddings. Today the doctor checks for
the Dewey binary and the embedding model file, but does
not verify that embeddings work end-to-end. This means an
engineer could pass all doctor checks yet encounter
failures when agents try to use semantic search. After
this change, `uf doctor` probes Dewey's embedding
capability directly, providing confidence that the full
semantic search pipeline is functional.

**Why this priority**: The embedding capability check
is the highest-value improvement because it catches the
most common failure mode — Ollama not serving or the
model not loaded — before the engineer encounters
cryptic errors during agent operations.

**Independent Test**: Run `uf doctor` with Dewey
installed and the embedding model pulled. Verify the
output shows an embedding capability check that passes.
Then stop Ollama (or remove the model) and re-run
`uf doctor`. Verify the embedding check fails with
actionable guidance.

**Acceptance Scenarios**:

1. **Given** Dewey is installed and the embedding model
   is available and serving, **When** the engineer runs
   `uf doctor`, **Then** the Dewey health check group
   includes an "embedding capability" check that reports
   PASS.
2. **Given** Dewey is installed but the embedding model
   is not loaded (Ollama not running or model not
   pulled), **When** the engineer runs `uf doctor`,
   **Then** the embedding capability check reports FAIL
   with an actionable hint (e.g., "Run `ollama pull
   granite-embedding:30m` or start Ollama").
3. **Given** Dewey is not installed, **When** the
   engineer runs `uf doctor`, **Then** the existing
   "dewey binary" check reports FAIL and the embedding
   capability check is skipped (cannot test embeddings
   without Dewey).

---

### User Story 2 — Forked Swarm Plugin Installation (Priority: P1)

When an engineer runs `uf setup`, the Swarm plugin is
installed from the organization's forked repository
(`unbound-force/swarm-tools`) instead of the upstream
source (`joelhooks/swarm-tools`). The fork contains
modifications needed for the Unbound Force ecosystem.
Today `uf setup` installs from the upstream source,
which may lack required patches or diverge from the
fork's behavior over time.

**Why this priority**: Installing the wrong Swarm plugin
version means the engineer's environment silently
diverges from the team's expected configuration, causing
hard-to-diagnose issues during swarm operations.

**Independent Test**: Run `uf setup` on a machine
without the Swarm plugin installed. Verify the
installation command references `unbound-force/swarm-tools`
(the fork), not `joelhooks/swarm-tools` (upstream).

**Acceptance Scenarios**:

1. **Given** the Swarm plugin is not installed, **When**
   the engineer runs `uf setup`, **Then** the plugin is
   installed from `unbound-force/swarm-tools`.
2. **Given** the upstream Swarm plugin
   (`joelhooks/swarm-tools`) is already installed,
   **When** the engineer runs `uf setup`, **Then** the
   setup replaces it with the forked version from
   `unbound-force/swarm-tools`.
3. **Given** the forked Swarm plugin is already
   installed and up to date, **When** the engineer runs
   `uf setup`, **Then** the setup detects the plugin is
   current and skips reinstallation.

---

### User Story 3 — Ollama Check Demotion (Priority: P3)

Since Dewey now manages the Ollama lifecycle (starting
and stopping Ollama as needed), the direct "Ollama
serving" health check in `uf doctor` may produce
misleading results. An engineer who sees "Ollama: FAIL"
might try to manually start Ollama, when in reality
Dewey handles this automatically. After this change,
the direct Ollama serving check is demoted to
informational status (not a failure condition),
reflecting that Dewey is the primary consumer and
lifecycle manager.

**Why this priority**: This is a polish change that
reduces confusion but does not block any functionality.
The embedding capability check (US-1) already covers
the critical path — if embeddings work, Ollama is
necessarily serving.

**Independent Test**: Run `uf doctor` with Ollama not
manually started but Dewey configured to manage it.
Verify the Ollama serving check shows as informational
(not FAIL) and the output explains that Dewey manages
Ollama's lifecycle.

**Acceptance Scenarios**:

1. **Given** Ollama is not manually running but Dewey is
   configured to manage it, **When** the engineer runs
   `uf doctor`, **Then** the Ollama serving check
   displays as informational (e.g., "Ollama: managed by
   Dewey") rather than a failure.
2. **Given** Ollama is manually running, **When** the
   engineer runs `uf doctor`, **Then** the Ollama
   serving check still shows PASS (no behavioral change
   for the positive case).

---

### Edge Cases

- What happens when Dewey is installed but its `serve`
  command is not running? The embedding check SHOULD
  fail with a hint to start Dewey (`dewey serve`), not
  to start Ollama directly.
- What happens when the npm/bun registry is unreachable
  during `uf setup`? The Swarm plugin install step
  SHOULD fail with a clear error and not leave the
  environment in a partially configured state.
- What happens when both the upstream and forked Swarm
  plugin packages are installed? The setup SHOULD
  uninstall the upstream package before installing the
  fork to avoid conflicts.
- What happens when the embedding check times out (e.g.,
  Ollama is overloaded or unresponsive)? The check
  SHOULD report WARN with a hint to retry. The timeout
  is 5 seconds per the implementation plan.
- What happens when the `unbound-force/swarm-tools` fork
  repository is unavailable (deleted, private, GitHub
  outage)? The install step fails with the package
  manager's error message. The engineer can retry or
  install manually.
- What happens when the engineer runs `uf doctor` with
  the `--format=json` flag? The embedding capability
  check and Ollama demotion MUST be reflected in the
  JSON output with the same semantics as the text
  output.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `uf doctor` MUST include a Dewey embedding
  capability check that verifies Dewey can generate
  embeddings end-to-end (per Spec 021 FR-020).
- **FR-002**: The Dewey embedding capability check MUST
  report FAIL with an actionable hint when embeddings
  cannot be generated (model not loaded, Ollama not
  serving, Dewey not running).
- **FR-003**: The Dewey embedding capability check MUST
  be skipped when Dewey is not installed (the existing
  "dewey binary" check already covers this).
- **FR-004**: The direct Ollama serving check in
  `uf doctor` MAY be demoted to informational status,
  reflecting that Dewey manages the Ollama lifecycle
  (per Spec 021 FR-021).
- **FR-005**: `uf setup` MUST install the Swarm plugin
  from `unbound-force/swarm-tools` instead of the
  upstream `joelhooks/swarm-tools` (per Spec 021
  FR-019).
- **FR-006**: `uf setup` MUST handle the case where the
  upstream Swarm plugin is already installed by
  replacing it with the forked version.
- **FR-007**: The embedding capability check and Ollama
  demotion MUST be reflected in `--format=json` output
  with consistent semantics.
- **FR-008**: All existing `uf doctor` and `uf setup`
  tests MUST continue to pass after these changes.

### Key Entities

- **Health Check**: A named diagnostic probe in
  `uf doctor` that reports PASS, FAIL, or INFO status
  with an optional hint message. Organized into groups
  (e.g., "Dewey Knowledge Layer").
- **Setup Step**: A named installation or configuration
  action in `uf setup` that installs a tool, configures
  a setting, or initializes a subsystem.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `uf doctor` reports Dewey embedding
  capability status in 100% of runs where Dewey is
  installed — verified by running doctor with Dewey
  available and checking the output includes the
  embedding check.
- **SC-002**: When embeddings are broken (model not
  loaded), 100% of `uf doctor` runs correctly identify
  the failure with an actionable hint — verified by
  removing the model and running doctor.
- **SC-003**: `uf setup` installs the Swarm plugin from
  the forked repository in 100% of fresh installations
  — verified by checking the install command targets
  `unbound-force/swarm-tools`.
- **SC-004**: All existing tests pass after the changes
  — verified by running the project's full test suite.
- **SC-005**: Engineers can resolve a failing embedding
  check without external documentation — the hint
  message alone provides sufficient guidance to fix the
  issue.

## Dependencies & Assumptions

### Dependencies

- **Issue #76** (Hivemind-to-Dewey agent migration):
  Completed. The agent migration ensures all learning
  storage and retrieval uses Dewey, making the embedding
  check relevant.
- **Dewey Ollama lifecycle management** (dewey#24):
  Dewey must manage Ollama's lifecycle for the Ollama
  check demotion (US-3) to be meaningful. Until then,
  the demotion can still be implemented but has less
  user impact.
- **Forked Swarm plugin** (`unbound-force/swarm-tools`):
  The fork must exist and be published to npm/bun
  registry for `uf setup` to install it.

### Assumptions

- The Dewey embedding capability can be tested by
  invoking Dewey's embedding functionality (e.g., a CLI
  command or health endpoint). The exact mechanism
  depends on what Dewey exposes — the implementation
  will use whatever Dewey provides.
- The forked Swarm plugin uses the same package name
  structure as the upstream, differing only in the
  registry source or git URL.
- Demoting the Ollama check to informational does not
  require a new check status type — it can reuse
  existing INFO or WARN semantics already present in
  the doctor framework.
