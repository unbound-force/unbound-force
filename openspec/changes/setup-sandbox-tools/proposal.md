## Why

The sandbox is becoming a non-optional part of the Unbound
Force workflow, yet its two infrastructure dependencies --
Podman (container runtime) and DevPod (workspace manager) --
are not covered by `uf setup` or `uf doctor`. Users must
manually install Podman, install DevPod, and configure the
DevPod Podman provider. The provider configuration changed:
the old standalone `podman` provider no longer exists in
DevPod; users must alias the Docker provider via
`devpod provider add docker --name podman -o DOCKER_COMMAND=podman`.

This manual process is error-prone and undiscoverable.
There is no diagnostic feedback when the provider is
misconfigured, leading to confusing runtime errors like
"couldn't find default provider podman".

## What Changes

Add Podman and DevPod to the `uf setup` installation
pipeline and enhance `uf doctor` diagnostics for both tools.

## Capabilities

### New Capabilities

- `setup: podman install`: Install Podman via Homebrew with
  platform-aware machine initialization on macOS.
- `setup: devpod install`: Install DevPod via Homebrew with
  fallback to download link.
- `setup: devpod provider config`: Automatically configure
  the DevPod Podman provider using the Docker provider alias
  (`devpod provider add docker --name podman -o DOCKER_COMMAND=podman`).
- `setup: podman smoke test`: After install and machine
  init (macOS), run `podman info` to verify Podman is
  functional. Report result but do not block subsequent
  steps on failure.
- `doctor: podman check`: Validate Podman presence and
  version (>= 4.3) as a required tool in `coreToolSpecs`.
- `doctor: podman runtime health`: Post-check after
  version passes -- runs `podman info` to verify Podman
  is functional. On macOS, also checks that a Podman
  machine exists and is running. Platform-aware hints
  on failure (macOS: `podman machine start`, Linux:
  `systemctl --user start podman.socket`).
- `doctor: devpod provider check`: Validate DevPod has a
  `podman` provider registered with actionable fix hint.
- `doctor: devpod version check`: Validate DevPod >= 0.5.0.

### Modified Capabilities

- `doctor: DevPod check group`: Enhanced with version check
  and provider registration check. Existing binary and
  devcontainer checks preserved.
- `setup: step count`: Increases from 13 to 16 steps.
- `doctor: coreToolSpecs`: Podman added as required tool
  with version parsing and minimum version enforcement.
- `doctor: install hints`: `homebrewInstallCmd`,
  `genericInstallCmd`, and `installURL` updated with
  entries for podman and devpod.

### Removed Capabilities

- None.

## Impact

- **`internal/setup/setup.go`**: Three new install/config
  functions, `Run()` step counter update (13 -> 16).
- **`internal/setup/setup_test.go`**: Tests for new install
  functions following existing injection patterns.
- **`internal/doctor/checks.go`**: Podman added to
  `coreToolSpecs` as required; `checkDevPod()` enhanced
  with version and provider checks.
- **`internal/doctor/environ.go`**: Install hints and URLs
  for podman and devpod.
- **`internal/doctor/doctor_test.go`**: Tests for podman
  core tool check and enhanced DevPod checks.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

This change adds tool installation and diagnostics to the
meta-repository CLI. It does not affect inter-hero artifact
communication or coupling.

### II. Composability First

**Assessment**: PASS

Podman and DevPod remain independently useful. Setup
installs each tool separately with independent skip
configuration (`shouldSkipTool("podman")`,
`shouldSkipTool("devpod")`). The DevPod provider
configuration step is gated on both tools being
available and degrades gracefully when either is absent.
Doctor checks use the existing conditional group pattern
(DevPod group hidden when not detected).

### III. Observable Quality

**Assessment**: PASS

Doctor produces machine-parseable output (JSON mode) for
all new checks with Pass/Warn/Fail severity, install
hints, and URLs. Setup reports step results with the
existing `stepResult` struct. All output follows
established formatting conventions.

### IV. Testability

**Assessment**: PASS

All new functions use the existing `Options` struct
dependency injection pattern. `LookPath`, `ExecCmd`,
and `ReadFile` are injected function fields, allowing
tests to verify behavior without real Podman or DevPod
binaries. No external service calls required.
