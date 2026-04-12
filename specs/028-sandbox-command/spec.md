# Feature Specification: Sandbox Command

**Feature Branch**: `028-sandbox-command`  
**Created**: 2026-04-12  
**Status**: Draft  
**Input**: User description: "issue #91 — Add uf sandbox command for containerized OpenCode sessions"

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Start an Isolated Agent Session (Priority: P1)

When an engineer runs `uf sandbox start`, a container
is launched with the full Unbound Force toolchain
running OpenCode in headless server mode, and the
engineer's terminal attaches to it automatically. Today,
starting a containerized session requires 4 manual
steps: check Ollama, build the podman run command with
correct flags, wait for the server to be ready, and
run `opencode attach`. After this change, one command
handles everything — prerequisites check, platform
detection, container start, health wait, and attachment.

**Why this priority**: This is the primary entry point
for isolated agent work. Without it, the container
image delivered by the containerfile repo is usable only
by engineers who know the exact podman flags.

**Independent Test**: Run `uf sandbox start` with Podman
and Ollama available. Verify the container starts, the
health check passes, and the terminal attaches to the
OpenCode session inside the container.

**Acceptance Scenarios**:

1. **Given** Podman is installed and Ollama is running,
   **When** the engineer runs `uf sandbox start`,
   **Then** a container is launched with the project
   directory mounted, OpenCode server starts, and the
   TUI attaches automatically.
2. **Given** the container is already running, **When**
   the engineer runs `uf sandbox start`, **Then** the
   command reports "sandbox already running" and offers
   to attach instead.
3. **Given** Podman is not installed, **When** the
   engineer runs `uf sandbox start`, **Then** the
   command fails with an actionable hint to install
   Podman.
4. **Given** the engineer runs `uf sandbox start
   --detach`, **When** the container starts, **Then**
   the command prints the server URL and exits without
   attaching.
5. **Given** the engineer runs `uf sandbox start
   --mode direct`, **When** the container starts,
   **Then** the project directory is mounted read-write
   (no isolation barrier).
6. **Given** the engineer runs `uf sandbox start
   --mode isolated` (default), **When** the container
   starts, **Then** the project directory is mounted
   read-only with a writable overlay.

---

### User Story 2 — Extract Changes from Container (Priority: P1)

When an engineer runs `uf sandbox extract`, changes
made by the agent inside the container are exported as
a patch and presented for review before being applied
to the host repo. Today, extracting changes requires
manually running `podman exec` with `git format-patch`
and then `git am` on the host. After this change, one
command handles the extraction, review, and application.

**Why this priority**: Without extraction, isolated mode
is a dead end — changes are trapped in the container.
This completes the round-trip workflow: start → work →
extract → review → apply.

**Independent Test**: Start a sandbox, make a change
inside it (create a file, commit), then run
`uf sandbox extract`. Verify the patch is presented for
review and applied to the host repo on confirmation.

**Acceptance Scenarios**:

1. **Given** the sandbox is running with committed
   changes, **When** the engineer runs `uf sandbox
   extract`, **Then** a patch is generated, displayed
   for review (files changed, insertions, deletions),
   and applied to the host repo on confirmation.
2. **Given** the sandbox has no uncommitted changes,
   **When** the engineer runs `uf sandbox extract`,
   **Then** the command reports "no changes to extract."
3. **Given** the sandbox is not running, **When** the
   engineer runs `uf sandbox extract`, **Then** the
   command fails with "no sandbox running."
4. **Given** the engineer declines the patch after
   review, **When** prompted, **Then** the patch is not
   applied and the command exits cleanly.

---

### User Story 3 — Manage Sandbox Lifecycle (Priority: P2)

Engineers can attach to, stop, and check status of the
running sandbox using `uf sandbox attach`,
`uf sandbox stop`, and `uf sandbox status`.

**Why this priority**: Lifecycle management is needed
for long-running sessions but is less critical than
start and extract — those are the core workflow.

**Independent Test**: Start a sandbox with `--detach`,
then use `attach`, `status`, and `stop` subcommands.
Verify each works correctly.

**Acceptance Scenarios**:

1. **Given** a sandbox is running, **When** the engineer
   runs `uf sandbox attach`, **Then** the TUI connects
   to the running container's OpenCode server.
2. **Given** a sandbox is running, **When** the engineer
   runs `uf sandbox stop`, **Then** the container is
   stopped and removed.
3. **Given** a sandbox is running, **When** the engineer
   runs `uf sandbox status`, **Then** the command shows
   container name, uptime, mode (isolated/direct),
   mounted project, and server URL.
4. **Given** no sandbox is running, **When** the
   engineer runs `uf sandbox attach`, **Then** the
   command fails with "no sandbox running. Run
   uf sandbox start."
5. **Given** no sandbox is running, **When** the
   engineer runs `uf sandbox stop`, **Then** the command
   reports "no sandbox to stop."

---

### User Story 4 — Platform-Aware Container Configuration (Priority: P2)

The sandbox command automatically detects the host
platform and configures the container accordingly.
On macOS arm64, it uses the arm64 image variant. On
Fedora with SELinux, it adds the `:Z` relabeling flag
to volume mounts. The engineer does not need to know
or specify platform-specific flags.

**Why this priority**: Platform detection is invisible
to the user but critical for correctness. Without it,
volume mounts fail on Fedora (SELinux) or the wrong
architecture is pulled.

**Independent Test**: Run `uf sandbox start` on both
macOS and Fedora. Verify the correct image architecture
is used and SELinux flags are applied when appropriate.

**Acceptance Scenarios**:

1. **Given** the host is macOS arm64, **When** the
   sandbox starts, **Then** the arm64 image variant is
   used and no SELinux flags are added.
2. **Given** the host is Fedora amd64 with SELinux
   enforcing, **When** the sandbox starts, **Then**
   volume mounts include the `:Z` flag.
3. **Given** the host is Fedora with SELinux disabled,
   **When** the sandbox starts, **Then** volume mounts
   do not include `:Z`.

---

### Edge Cases

- What happens when Ollama is not running? The command
  SHOULD warn but still start the container — Dewey
  will function without embeddings (keyword search
  only). The warning should suggest starting Ollama.
- What happens when the container image is not cached?
  The command MUST pull the image automatically before
  starting.
- What happens when port 4096 is already in use by
  something other than the sandbox? The command SHOULD
  fail with a clear error identifying the port conflict.
- What happens when the engineer's API key environment
  variables are not set? The command SHOULD warn but
  still start — the agent will fail on first prompt
  but the container itself runs fine.
- What happens when `opencode attach` is not available
  on the host? The command MUST check for `opencode`
  in PATH before attempting to attach and suggest
  installation if missing.
- What happens when the container crashes during a
  session? `uf sandbox status` SHOULD show the
  container as stopped with the exit code.
  `uf sandbox start` SHOULD clean up the dead container
  before starting a new one.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `uf sandbox start` MUST check for Podman
  in PATH and fail with an actionable hint if missing.
- **FR-002**: `uf sandbox start` MUST detect the host
  platform (architecture and SELinux status) and
  configure the container accordingly.
- **FR-003**: `uf sandbox start` MUST pull the container
  image if not cached locally.
- **FR-004**: `uf sandbox start` MUST start the
  container with the project directory mounted,
  environment variables forwarded, and resource limits
  applied.
- **FR-005**: `uf sandbox start` MUST wait for the
  OpenCode server health check to pass before attaching
  (timeout: 60 seconds).
- **FR-006**: `uf sandbox start` MUST call
  `opencode attach` to connect the TUI to the container
  (unless `--detach` is specified).
- **FR-007**: `uf sandbox start --detach` MUST print the
  server URL and exit without attaching.
- **FR-008**: `uf sandbox start --mode isolated` (default)
  MUST mount the project directory as read-only.
- **FR-009**: `uf sandbox start --mode direct` MUST mount
  the project directory as read-write.
- **FR-010**: `uf sandbox extract` MUST generate a patch
  from the container using `git format-patch` or
  `git diff`.
- **FR-011**: `uf sandbox extract` MUST present the
  patch for human review before applying.
- **FR-012**: `uf sandbox extract` MUST apply the patch
  via `git am` only after confirmation.
- **FR-013**: `uf sandbox attach` MUST connect the TUI
  to the running container's OpenCode server.
- **FR-014**: `uf sandbox stop` MUST stop and remove the
  container.
- **FR-015**: `uf sandbox status` MUST display container
  state, uptime, mode, mounted project, and server URL.
- **FR-016**: Only one sandbox container is supported at
  a time. Starting a second MUST fail with a message
  directing the user to stop the existing one first.
- **FR-017**: The container image MUST be configurable
  via `--image` flag or `UF_SANDBOX_IMAGE` environment
  variable (default:
  `quay.io/unbound-force/opencode-dev:latest`).
- **FR-018**: Resource limits MUST be configurable via
  `--memory` and `--cpus` flags (defaults: 8g and 4).
- **FR-019**: All external dependencies (Podman,
  OpenCode) MUST be injectable for testability (same
  `LookPath`/`ExecCmd` pattern as setup and doctor).
- **FR-020**: `uf sandbox start` MUST forward Google
  Vertex AI environment variables
  (`GOOGLE_CLOUD_PROJECT`, `VERTEX_LOCATION`,
  `GOOGLE_APPLICATION_CREDENTIALS`) to the container
  when they are set in the host environment.
- **FR-021**: When `GOOGLE_APPLICATION_CREDENTIALS` is
  set and points to a file, `uf sandbox start` MUST
  mount that file into the container as a read-only
  volume and set the env var to the container-internal
  path.
- **FR-022**: When `GOOGLE_APPLICATION_CREDENTIALS` is
  not set, `uf sandbox start` MUST mount
  `~/.config/gcloud/application_default_credentials.json`
  into the container as a read-only volume if the file
  exists on the host.
- **FR-023**: All existing tests MUST continue to pass.

### Key Entities

- **Sandbox**: A running Podman container with the
  OpenCode server, identified by the container name
  `uf-sandbox`. Has a mode (isolated or direct), a
  mounted project directory, and a server URL.
- **Patch**: A set of changes generated from the
  container's git history, presented for review and
  applied to the host repo.
- **Platform Config**: Detected host properties
  (architecture, SELinux status) that influence
  container flags.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: An engineer can start an isolated agent
  session with a single command (`uf sandbox start`)
  in under 30 seconds (excluding image pull time) —
  verified by timing the start-to-attach flow.
- **SC-002**: Changes made inside the sandbox are
  extractable and applicable to the host repo in under
  5 steps (extract, review, confirm) — verified by
  completing a round-trip workflow.
- **SC-003**: 100% of destructive agent actions inside
  the sandbox do not affect the host filesystem when
  running in isolated mode — verified by running a
  destructive command inside the sandbox and checking
  the host.
- **SC-004**: The command works on both macOS arm64 and
  Fedora amd64 without platform-specific user
  intervention — verified by running on both platforms.
- **SC-005**: All existing tests pass after the changes
  — verified by running the full test suite.

## Dependencies & Assumptions

### Dependencies

- **Container image**
  (`unbound-force/containerfile#1`): Complete. The
  `quay.io/unbound-force/opencode-dev` image is
  available.
- **OpenCode `attach` command**: Built into OpenCode.
  Required for TUI connection from host to container.
- **Podman**: Required on the host. Not installed by
  `uf setup` — documented as a prerequisite.

### Assumptions

- The container image exposes OpenCode server on port
  4096 via `opencode serve --port 4096 --hostname
  0.0.0.0`.
- The container's entrypoint handles `uf init` on first
  start (per the containerfile repo's `entrypoint.sh`).
- `podman run` supports `--platform` flag for
  architecture selection on macOS (Docker Desktop
  compatibility layer in Podman).
- The `opencode attach` command accepts a URL argument
  (`opencode attach http://localhost:4096`).
- `host.containers.internal` resolves to the host from
  inside the Podman container on both macOS and Fedora.
- The engineer's LLM API key environment variables
  (e.g., `ANTHROPIC_API_KEY`) are available in the host
  shell and can be forwarded to the container via
  `-e` flags.
- Google Vertex AI authentication requires forwarding
  `GOOGLE_CLOUD_PROJECT`, `VERTEX_LOCATION`, and
  `GOOGLE_APPLICATION_CREDENTIALS` environment variables
  to the container.
- When `GOOGLE_APPLICATION_CREDENTIALS` points to a
  service account key file, that file MUST be mounted
  into the container as a read-only volume.
- When `GOOGLE_APPLICATION_CREDENTIALS` is not set,
  the gcloud Application Default Credentials file
  (`~/.config/gcloud/application_default_credentials.json`)
  MUST be mounted into the container if it exists,
  enabling `gcloud auth application-default login`
  based authentication.

## Clarifications

### Session 2026-04-12

- Q: Should Google Vertex AI env vars be forwarded? → A: Yes. Forward GOOGLE_CLOUD_PROJECT, VERTEX_LOCATION, and GOOGLE_APPLICATION_CREDENTIALS.
- Q: How should Google Cloud auth work inside the container? → A: Mount the service account key file (when GOOGLE_APPLICATION_CREDENTIALS is set) or mount ~/.config/gcloud/application_default_credentials.json (gcloud ADC fallback). Both as read-only volumes.
