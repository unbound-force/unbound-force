# Changelog

All notable changes to this project are documented in this file.
Each entry follows the format: `- <change-name>: <summary>`.

## Unreleased

### Added
- `python-convention-pack`: Added Python convention pack
  (`python.md`, `python-custom.md`) with 46 rules across 7
  sections (Coding Style, Architectural Patterns, Security
  Checks, Testing Conventions, Type Annotations, Documentation
  Requirements, Custom Rules). Expanded `detectLang()` with 5
  additional Python markers (`setup.py`, `setup.cfg`,
  `requirements.txt`, `tox.ini`, `Pipfile`). Added conditional
  "Python Tools" doctor check group with 9 tool categories
  and tool-agnostic alternative detection (ruff satisfies
  formatter, linter, import sorter, security scanner).
  (Spec: openspec/changes/python-convention-pack/)
- `/address-feedback` slash command for structured PR review
  feedback triage (Spec: openspec/changes/address-feedback/)
- `feedback-triage` JSON schema (v1.0.0) for capturing triage
  decisions as artifacts (Spec: openspec/changes/address-feedback/)
- Constitution Principle V (Security by Default) covering
  supply chain integrity, input validation, least privilege,
  and dependency necessity (version 1.1.0 → 1.2.0)
  (Spec: openspec/changes/review-finding-consolidation/)
- Compound Severity Escalation section in `severity.md` —
  instructs personas to assess combined severity when
  multiple findings share a root cause
  (Spec: openspec/changes/review-finding-consolidation/)
- Severity Calibration step (8g) in `/review-pr` — forces
  re-reading `severity.md` definitions to counter anchoring
  bias (Fixes #233)
- Adversarial Input Enumeration in `divisor-adversary.md`
  (new checklist item 5) and `/review-pr` Step 8b — per-input
  threat analysis for new parameters/secrets
  (Fixes #233)
- Issue Suggestion Gap Detection in `/review-pr` Step 8a —
  checks linked issues for unimplemented code suggestions
  (Fixes #233)
- CI Bot Annotation Cross-referencing (Step 8e) and Finding
  Consolidation (Step 8f) in `/review-pr` — ingests bot
  findings as corroborating evidence and merges related
  findings by root cause (Fixes #228)
- Cross-persona finding consolidation in `/review-council`
  Code Review Step 3 and Spec Review Step 2 (Fixes #228)
- Dependency necessity check and content-integrity
  verification guidance in `divisor-adversary.md` Audit
  Checklist item 2 (Fixes #228)

### Changed
- `uf init` sub-tool initializations (dewey, replicator,
  specify, openspec, gaze) now run concurrently, reducing
  wall-clock time from sequential sum to parallel max.
  Dewey indexing (30-60s) no longer blocks the other four
  tools. Individual tool failures remain non-fatal.
  (Spec: openspec/changes/parallel-subtool-init/)
- `uf init` now respects `setup.tools.<name>.method: skip`
  in `.uf/config.yaml` — skipped tools are excluded from
  initialization entirely, even when the binary is on PATH.
  (Spec: openspec/changes/parallel-subtool-init/)

### Fixed
- `uf setup` auto mode now falls back to native package
  managers and `go install` when Homebrew is absent.
  Previously, 4 companion tools (Gaze, Replicator, Dewey,
  GitHub CLI) were skipped with download-link hints on
  Fedora/RHEL. Fallback chain: Homebrew -> dnf -> go install
  -> skip. Added `resolveMethod()` and `installViaGo()`
  helpers. `UF_PACKAGE_MANAGER=dnf` now forces the dnf
  install path. Fixes #214.
  (Spec: openspec/changes/setup-auto-native-fallback/)

## Recent Changes
- 036-subtool-error-reporting: Added Go 1.25+ + cobra, charmbracelet/log, lipgloss (no new deps)

- opsx/setup-sandbox-tools: Added Podman and DevPod to `uf setup` (13→16 steps) and `uf doctor`. Setup installs Podman via Homebrew with macOS Podman machine initialization (best-effort, with timeout via gtimeout/timeout fallback), installs DevPod via Homebrew, and configures the DevPod Podman provider alias (`devpod provider add docker --name podman -o DOCKER_PATH=podman`). Corrected DevPod provider option from `DOCKER_COMMAND` to `DOCKER_PATH` across setup, doctor, and spec artifacts. Doctor adds conditional DevPod check group (only when DevPod is detected): binary presence, version >= 0.5.0, podman provider registration, and `.devcontainer/devcontainer.json` existence. Podman doctor checks: version >= 4.3, runtime health via `podman info` (platform-aware), and Docker-to-Podman shim detection. Refactored setup step dispatch from 16 repetitive if/else blocks to data-driven `stepDef` slice with `executeSteps()` loop (CRAP 32 → 2). Added `hasProvider()` helper with exact first-column name matching for provider detection. Added `podmanMachineInit()` with `initMachineWithTimeout()` helper for macOS post-install (tries gtimeout, timeout, then no-timeout fallback). 6 new doctor tests, 12 new setup tests. Modified files: `internal/setup/setup.go`, `internal/setup/setup_test.go`, `internal/doctor/checks.go`, `internal/doctor/doctor.go`, `internal/doctor/doctor_test.go`, `internal/doctor/environ.go`. Website issue: unbound-force/website#121.
- opsx/sandbox-ide-flag: Added `--ide` flag to `uf sandbox create` and `uf sandbox start` for DevPod IDE selection (none, vscode, openvscode, fleet, jupyternotebook, cursor). Default `"none"` preserves backward compatibility. IDE resolution chain: CLI flag > `UF_SANDBOX_IDE` env var > `.uf/config.yaml` `sandbox.ide` > default. Added `validateIDE()` with strict allowlist validation. Added `waitForHealth()` with exponential backoff (500ms-5s, 60s timeout) to DevPod Start for health check polling after resume. Added `postStartCommand` to devcontainer.json template for OpenCode server auto-start. Fixed `Attach()` to detect persistent workspaces via `isPersistentWorkspace()` before ephemeral fallback. Fixed `Destroy()` to handle ephemeral cleanup directly when no persistent workspace exists. Added IDE field to unified config (`merge()`, `Defaults()`, `IsEmpty()`, `applyEnvOverrides()`). Fixed DevPod tunnel error suppression via status-based detection (D12). Added SSH fallback to start OpenCode serve when postStartCommand didn't run (D13). Replaced fmt.Fscanln with bufio.Scanner for destroy confirmation (D14). Added HealthCheckTimeout to Options for DI-based test injection. OS-aware devcontainer generation (D15): macOS uses `--userns=keep-id:uid=1000,gid=1000` for Podman VM UID mapping, Linux uses plain `--userns=keep-id` to avoid subuid range conflicts on container restart. `.devcontainer/` gitignored — generated per-user by `uf sandbox init`, no longer deployed by `uf init`. `json.NewEncoder` with `SetEscapeHTML(false)` preserves shell characters (`>`, `&`) in `postStartCommand`. Deleted orphaned root `.devcontainer.json`. 25+ new tests (including OS-aware devcontainerRunArgs, shell char preservation). Website issue: unbound-force/website#120. Modified files: `internal/sandbox/sandbox.go`, `internal/sandbox/devpod.go`, `internal/sandbox/config.go`, `internal/sandbox/workspace.go`, `internal/sandbox/sandbox_test.go`, `internal/config/config.go`, `internal/config/template.go`, `internal/scaffold/assets/devcontainer/devcontainer.json`, `internal/scaffold/scaffold.go`, `internal/scaffold/scaffold_test.go`, `.gitignore`, `cmd/unbound-force/sandbox.go`, `cmd/unbound-force/sandbox_test.go`, `cmd/unbound-force/main_test.go`.
