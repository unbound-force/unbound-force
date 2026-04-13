# Feature Specification: Sandbox CDE Lifecycle

**Feature Branch**: `029-sandbox-cde-lifecycle`  
**Created**: 2026-04-13  
**Status**: Draft  
**Input**: User description: "issue #95 — Extend uf sandbox with create/destroy lifecycle and CDE backend"

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Persistent Workspace Creation (Priority: P1)

When an engineer runs `uf sandbox create`, a persistent
CDE workspace is provisioned that survives across
OpenCode sessions, `/unleash` runs, and stop/start
cycles. Today, `uf sandbox start` creates an ephemeral
container that is destroyed on `stop` — all agent
context, compiled binaries, Dewey index, and git history
are lost. After this change, `create` provisions a
persistent workspace (via Eclipse Che/Dev Spaces or
Podman named volume), and `start`/`stop` control its
running state without destroying data.

**Why this priority**: The iterative demo loop
(`/unleash` → demo → `/speckit.clarify` → `/unleash` →
repeat) requires persistent state between iterations.
Without it, each cycle forces an extract/recompile
workflow that breaks the tight feedback loop and makes
the process tedious enough that engineers will subvert
it.

**Independent Test**: Run `uf sandbox create`. Run
`uf sandbox start`, make changes, commit, run
`uf sandbox stop`. Run `uf sandbox start` again. Verify
the changes, Dewey index, and compiled binaries are
still present.

**Acceptance Scenarios**:

1. **Given** no sandbox exists, **When** the engineer
   runs `uf sandbox create`, **Then** a persistent
   workspace is provisioned with the project's source
   code and the full toolchain available.
2. **Given** a sandbox exists and is stopped, **When**
   the engineer runs `uf sandbox start`, **Then** the
   workspace resumes with all state preserved (git
   history, `.uf/` data, compiled binaries).
3. **Given** a sandbox is running, **When** the engineer
   runs `uf sandbox stop`, **Then** the workspace is
   stopped but its data persists on disk.
4. **Given** a sandbox exists, **When** the engineer
   runs `uf sandbox destroy`, **Then** the workspace
   and all its state are permanently deleted.
5. **Given** the engineer runs `uf sandbox create` on a
   machine without CDE access, **Then** a Podman-based
   persistent workspace is created using named volumes
   (fallback mode).

---

### User Story 2 — Iterative Demo Loop Without Extraction (Priority: P1)

When an engineer uses CDE, the demo review happens
inside the workspace without any extraction step. The
engineer runs slash commands via `opencode attach` on
their host terminal. They review the demo in the Che
IDE — viewing files in the editor, running CLI commands
in the terminal tab, opening web apps in a browser tab,
and testing APIs via curl. When they want changes, they
provide feedback via `/speckit.clarify` or edit the spec
in the Che IDE. `/unleash` resumes from the clarify
step with full context preserved.

**Why this priority**: This is the core workflow that
makes CDE the primary development model. If the demo
loop requires extraction, the CDE is no better than
local Podman.

**Independent Test**: Create a sandbox, run `/unleash`
to the demo step, review the demo in the Che IDE,
provide feedback via `/speckit.clarify`, re-run
`/unleash`, and verify the agent picks up the spec
changes and produces an updated demo.

**Acceptance Scenarios**:

1. **Given** a CDE sandbox is running, **When** the
   engineer runs `/unleash` and it reaches the demo
   step, **Then** the demo is reviewable inside the Che
   IDE (terminal tab for CLI, browser tab for web apps,
   curl for APIs).
2. **Given** the engineer provides feedback via
   `/speckit.clarify` in the opencode TUI, **When**
   `/unleash` is re-run, **Then** it resumes from the
   clarify step with the full workspace state preserved
   (no re-clone, no re-index, no re-compile from
   scratch).
3. **Given** the engineer edits the spec directly in the
   Che IDE and pushes via git, **When** `/unleash` is
   re-run, **Then** the agent detects the spec changes
   via git pull and incorporates them.
4. **Given** the engineer runs `opencode attach` from
   the host, **When** they type slash commands, **Then**
   the commands execute inside the CDE workspace with
   access to the full toolchain and workspace state.

---

### User Story 3 — Bidirectional Git Sync (Priority: P2)

Changes flow bidirectionally between the engineer's host
and the CDE workspace via git. The agent commits and
pushes from inside the workspace. The engineer can pull
the branch on the host to review, or push spec edits
that the agent picks up.

**Why this priority**: Git sync is the coordination
mechanism that replaces the `uf sandbox extract` workflow.
Without it, the CDE workspace is isolated in the wrong
way — changes can't leave or enter.

**Independent Test**: Start a CDE sandbox. Have the agent
commit a change inside the workspace and push. Verify
the commit appears on the host via `git pull`. Push a
spec edit from the host. Verify the workspace sees it.

**Acceptance Scenarios**:

1. **Given** the agent commits a change inside the
   workspace, **When** the agent pushes to the branch,
   **Then** the engineer can pull the branch on the host
   and see the changes.
2. **Given** the engineer edits a spec on the host and
   pushes, **When** the workspace pulls (manually or
   on next `/unleash` run), **Then** the updated spec
   is available inside the workspace.
3. **Given** both the engineer and agent push to the
   same branch, **When** a merge conflict occurs,
   **Then** the conflict is reported and must be
   resolved before the next `/unleash` run.

---

### User Story 4 — CDE Backend for Eclipse Che / Dev Spaces (Priority: P1)

The `uf sandbox` command supports a CDE backend that
provisions workspaces via Eclipse Che or Red Hat
OpenShift Dev Spaces. The engineer specifies the
backend with `--backend che` or the command auto-detects
it when a Che instance is configured.

**Why this priority**: CDE is the primary workflow for
the iterative demo loop. The Podman backend (Spec 028)
is the fallback for engineers without CDE access.

**Independent Test**: Configure a Che instance URL. Run
`uf sandbox create --backend che`. Verify the workspace
is created in Che from the project's devfile. Run
`opencode attach` from the host. Verify the TUI
connects to the OpenCode server in the Che workspace.

**Acceptance Scenarios**:

1. **Given** a Che/Dev Spaces instance is configured,
   **When** the engineer runs `uf sandbox create
   --backend che`, **Then** a workspace is provisioned
   from the project's devfile.
2. **Given** no Che instance is configured and no
   `--backend` flag is provided, **When** the engineer
   runs `uf sandbox create`, **Then** the Podman backend
   is used (fallback).
3. **Given** a CDE workspace is running, **When** the
   engineer runs `uf sandbox attach`, **Then** the TUI
   connects to the OpenCode server inside the workspace.
4. **Given** a CDE workspace is running, **When** the
   engineer opens the Che IDE in their browser, **Then**
   they can view files, edit specs, run terminal
   commands, and access exposed endpoints for demos.

---

### Edge Cases

- What happens when `uf sandbox create` is called but
  a sandbox already exists? The command SHOULD report
  "sandbox already exists" and suggest `start` or
  `destroy`.
- What happens when the CDE instance is unreachable?
  The command SHOULD fail with an actionable error and
  suggest checking the Che URL and authentication.
- What happens when the devfile is missing from the
  project? The command SHOULD fail with an error
  suggesting the engineer add a `devfile.yaml` or use
  the `--backend podman` fallback.
- What happens when git push fails inside the workspace
  (no credentials, no remote configured)? The command
  SHOULD warn and fall back to the extract workflow for
  getting changes out.
- What happens when the Podman fallback uses named
  volumes? The named volume persists across `stop`/
  `start` cycles. `destroy` removes the named volume.
  The volume name follows the pattern
  `uf-sandbox-<project-name>`.
- What happens when multiple projects each need a
  sandbox? Each project gets its own named sandbox
  (`uf-sandbox-<project-name>`). The single-container
  constraint from Spec 028 is relaxed to
  single-container-per-project.

## Out of Scope

- Docker backend (Podman and CDE only for v1)
- Windows platform support
- Multi-container workspaces (one container per project)
- Remote Podman (local Podman only; CDE handles remote)
- Automatic conflict resolution for git sync divergence
- Integration tests requiring real containers or CDE
  instances (unit tests with injected dependencies only)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `uf sandbox create` MUST provision a
  persistent workspace with the project's source code
  and full toolchain.
- **FR-002**: `uf sandbox create --backend che` MUST
  provision a workspace via Eclipse Che/Dev Spaces from
  the project's devfile.
- **FR-003**: `uf sandbox create --backend podman` MUST
  provision a persistent workspace using Podman named
  volumes for state preservation.
- **FR-004**: `uf sandbox create` without `--backend`
  MUST auto-detect: use CDE if configured, Podman
  otherwise.
- **FR-005**: `uf sandbox start` MUST start a stopped
  workspace without losing state.
- **FR-006**: `uf sandbox stop` MUST stop a running
  workspace while preserving all state (git history,
  `.uf/` data, compiled binaries, Dewey index).
- **FR-007**: `uf sandbox destroy` MUST permanently
  delete the workspace and all associated state.
- **FR-008**: The CDE workspace MUST expose endpoints
  for OpenCode server (port 4096), web app demos, and
  API testing.
- **FR-009**: `opencode attach` from the host MUST
  connect to the OpenCode server running inside the CDE
  workspace.
- **FR-010**: The engineer MUST be able to view and edit
  files in the Che IDE while the agent works via
  `opencode attach`.
- **FR-011**: Git sync MUST be bidirectional — the agent
  pushes from the workspace and the engineer pushes from
  the host or Che IDE.
- **FR-012**: The CDE workspace MUST persist across
  multiple OpenCode sessions (detach/reattach) and
  multiple `/unleash` runs.
- **FR-013**: The CDE workspace MUST persist across
  `stop`/`start` cycles.
- **FR-014**: API keys MUST be injectable via the CDE
  platform's secret management (Che user preferences or
  Kubernetes secrets), not via host environment variable
  forwarding.
- **FR-015**: The Ollama endpoint MUST be configurable
  for CDE deployments where
  `host.containers.internal` does not resolve (e.g.,
  Kubernetes service URL).
- **FR-016**: The Podman fallback MUST use named volumes
  for persistent state, following the pattern
  `uf-sandbox-<project-name>`.
- **FR-017**: All external dependencies (Podman, Che
  API, OpenCode) MUST be injectable for testability.
- **FR-018**: All existing tests and the existing
  `uf sandbox` subcommands (attach, extract, status)
  MUST continue to work.

### Key Entities

- **Workspace**: A persistent development environment
  with the project's source code, toolchain, and agent
  state. Can be running or stopped. Has a backend
  (CDE or Podman) and a name.
- **Backend**: The infrastructure that hosts the
  workspace. CDE backend uses Eclipse Che/Dev Spaces.
  Podman backend uses local containers with named
  volumes.
- **Demo Endpoint**: An exposed port in the workspace
  accessible from the engineer's browser for reviewing
  web app or API demos.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: An engineer can complete 3 iterations of
  the demo loop (`/unleash` → demo → `/speckit.clarify`
  → `/unleash`) without any extract/recompile step —
  verified by timing 3 full cycles.
- **SC-002**: A CDE workspace preserves all state (git
  history, `.uf/` data, Dewey index) across a
  `stop`/`start` cycle — verified by checking file
  timestamps and Dewey query results before and after.
- **SC-003**: A CDE workspace preserves all state across
  OpenCode session boundaries (detach and reattach) —
  verified by detaching, reattaching, and running
  `/unleash` to confirm resumability.
- **SC-004**: An engineer can review a CLI demo in the
  Che IDE terminal tab and a web app demo in a browser
  tab — verified by running demo commands and opening
  exposed endpoints.
- **SC-005**: Bidirectional git sync works: agent pushes
  from workspace, engineer pulls on host; engineer
  pushes from host, agent pulls in workspace — verified
  by completing both directions.
- **SC-006**: All existing Spec 028 tests and
  subcommands continue to pass — verified by running
  the full test suite.

## Dependencies & Assumptions

### Dependencies

- **Spec 028** (`uf sandbox` base command): Complete.
  Provides the existing `start`, `stop`, `attach`,
  `extract`, `status` subcommands and the
  `internal/sandbox/` package.
- **Containerfile repo** (`containerfile#1`): Complete.
  Provides the container image and devfiles.
- **Containerfile devfile updates** (`containerfile#3`):
  Open. Devfiles need demo port endpoints and
  configurable Ollama endpoint. Required for CDE
  backend to expose demo ports.
- **Containerfile credential docs** (`containerfile#4`):
  Open. Documentation for K8s secret injection. Required
  for CDE API key management.

### Assumptions

- Eclipse Che provides a REST API or `chectl` CLI for
  workspace provisioning from a devfile. The
  implementation will use whichever is available.
- The Che/Dev Spaces instance URL is configured via
  environment variable (`UF_CHE_URL`) or a config file
  (`.uf/sandbox.yaml`).
- Che workspace names follow the pattern
  `uf-<project-name>` for uniqueness across projects.
- The Podman named volume approach uses
  `podman volume create` and `-v <name>:/workspace`
  instead of bind mounts, preserving data across
  container lifecycle.
- The existing `uf sandbox start` from Spec 028
  continues to work as-is for engineers who don't
  use `create`/`destroy` — backward compatible.
- `opencode attach` from the host can connect to a
  CDE workspace's OpenCode server via the exposed Che
  endpoint URL.
