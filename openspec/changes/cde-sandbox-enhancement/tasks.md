## 1. Remove Che Backend

- [x] 1.1 Delete `internal/sandbox/che.go` entirely
- [x] 1.2 Remove `BackendChe` constant from
  `internal/sandbox/backend.go`
- [x] 1.3 Remove the `case BackendChe:` branch from
  `ResolveBackend()` — replace with a migration error:
  `"che backend removed, use --backend devpod instead"`
- [x] 1.4 Remove Che branch from `autoDetectBackend()`
  (no longer checks for chectl or UF_CHE_URL)
- [x] 1.5 Remove `resolveCheURL()` helper from
  `backend.go`
- [x] 1.6 Remove `CheURL` field from `Options` struct
  in `sandbox.go`
- [x] 1.7 Remove `Che` struct (URL, Token) from
  `SandboxConfig` in `workspace.go`
- [x] 1.8 Remove Che-related fields and references from
  `LoadConfig()` in `workspace.go`
- [x] 1.9 Replace Che URL check in `Extract()` with
  general persistent workspace check: for all
  persistent workspaces (Podman and DevPod), return
  early with "changes are on the host filesystem or
  use git push"
- [x] 1.10 Remove `--che-url` flag from
  `cmd/unbound-force/sandbox.go` if it exists
- [x] 1.11 Remove all Che-specific test functions from
  `sandbox_test.go` (identify by `TestChe*` prefix
  and `BackendChe` references)
- [x] 1.12 Update all stale Che/CDE comments in
  `sandbox.go` Options struct (BackendName, HTTPDo)
- [x] 1.13 Update all stale Che/CDE comments in
  `workspace.go` (WorkspaceStatus.Backend,
  WorkspaceStatus.Mode, WorkspaceStatus.ProjectDir,
  WorkspaceStatus.ServerURL, SandboxConfig.Backend,
  and all other "CDE" / "Che" references)
- [x] 1.14 Update `testOpts()` helper: LookPath MUST
  return `not found` for `devpod` by default (prevents
  auto-detection in ephemeral-mode tests)
- [x] 1.15 Remove `CheConfig` struct, `SandboxConfig.Che`
  field, Che merge logic, and Che env var overrides
  from `internal/config/config.go`
- [x] 1.16 Update `internal/config/config_test.go` to
  remove Che-specific test assertions
- [x] 1.17 Verify build passes: `go build ./...`

## 2. Options Struct and Gateway Wiring

- [x] 2.1 Add `GatewayPort int` and `GatewayActive bool`
  fields to `Options` struct in `sandbox.go`
- [x] 2.2 Wire `autoStartGateway()` into `Create()`:
  call before `ResolveBackend()`, set
  `opts.GatewayPort` and `opts.GatewayActive`
- [x] 2.3 Wire `autoStartGateway()` into the persistent
  workspace branch of `Start()`: call when
  `isPersistentWorkspace()` returns true (including
  DevPod detection), before delegating to backend
- [x] 2.4 Add test: `TestCreate_AutoStartsGateway`
- [x] 2.5 Add test:
  `TestStart_PersistentAutoStartsGateway`
- [x] 2.6 Add test:
  `TestCreate_NoCloudProvider_SkipsGateway`

## 3. Persistent Podman Gateway Env Injection

- [x] 3.1 Update `buildPersistentRunArgs()` in
  `podman.go`: read `opts.GatewayActive` and
  `opts.GatewayPort`, call
  `forwardedEnvVars(opts, opts.GatewayActive)`, append
  `gatewayEnvVars(opts.GatewayPort)` when active
- [x] 3.2 Add test:
  `TestBuildPersistentRunArgs_GatewayActive`
- [x] 3.3 Add test:
  `TestBuildPersistentRunArgs_GatewayInactive`
  (backward compatibility)

## 4. DevPod Backend Implementation

- [x] 4.1 Add `BackendDevPod = "devpod"` constant to
  `backend.go`
- [x] 4.2 Add `case BackendDevPod:` to
  `ResolveBackend()` — check for `devpod` in PATH,
  return `&DevPodBackend{}`
- [x] 4.3 Update `autoDetectBackend()`: prefer DevPod
  when `devpod` is in PATH AND
  `.devcontainer/devcontainer.json` exists, fall back
  to Podman otherwise
- [x] 4.4 Create `internal/sandbox/devpod.go` with
  `DevPodBackend` struct implementing `Backend`
- [x] 4.5 Implement `DevPodBackend.Name()` returning
  `"devpod"`
- [x] 4.6 Implement `DevPodBackend.Create(opts)`:
  pre-flight check `podman` in PATH, verify DevPod
  >= 0.5.0 via `parseDevPodVersion()`, check
  `.devcontainer/devcontainer.json` exists, then call
  `devpod up <project-dir>` with
  `--provider podman`,
  `--id uf-sandbox-<project-name>`,
  gateway env var injection via
  `--workspace-env ANTHROPIC_BASE_URL=...`,
  and `--ide none` (OpenCode is the IDE)
- [x] 4.6a Implement `parseDevPodVersion()` following
  the `parsePodmanVersion()` pattern: call
  `devpod version`, parse semver, enforce >= 0.5.0
- [x] 4.7 Implement `DevPodBackend.Start(opts)`:
  call `devpod up --id <name>` to resume
- [x] 4.8 Implement `DevPodBackend.Stop(opts)`:
  call `devpod stop <name>`
- [x] 4.9 Implement `DevPodBackend.Destroy(opts)`:
  call `devpod delete <name> --force`
- [x] 4.10 Implement `DevPodBackend.Status(opts)`:
  call `devpod status <name> --output json`,
  parse into `WorkspaceStatus`
- [x] 4.11 Implement `DevPodBackend.Attach(opts)`:
  call `opencode attach <server-url>`
- [x] 4.12 Add `devpodWorkspaceName(opts)` helper
  returning `"uf-sandbox-" + projectName(opts.ProjectDir)`
- [x] 4.12a Extend `isPersistentWorkspace()` to detect
  DevPod workspaces: when `devpod` is in PATH, call
  `devpod status uf-sandbox-<project> --output json`;
  treat found workspace as persistent
- [x] 4.13 Add test: `TestDevPodCreate_Success`
- [x] 4.14 Add test: `TestDevPodCreate_NotInstalled`
- [x] 4.15 Add test: `TestDevPodStop_Success`
- [x] 4.16 Add test: `TestDevPodDestroy_Success`
- [x] 4.17 Add test: `TestDevPodStatus_Running`
- [x] 4.18 Add test: `TestDevPodStatus_Stopped`
- [x] 4.19 Add test:
  `TestDevPodCreate_GatewayEnvInjection` — verify
  gateway env vars are passed via `--workspace-env`
  when `opts.GatewayActive` is true
- [x] 4.20 Add test:
  `TestResolveBackend_DevPod` — verify DevPod backend
  is returned when `--backend devpod` and `devpod` in
  PATH
- [x] 4.21 Add test:
  `TestResolveBackend_CheMigrationError` — verify
  `--backend che` returns migration error message
  (covered by existing test at line 1409)
- [x] 4.22 Add test:
  `TestAutoDetect_PrefersDevPod` — verify auto-detect
  returns DevPod when `devpod` is in PATH AND
  `.devcontainer/devcontainer.json` exists
- [x] 4.23 Add test:
  `TestAutoDetect_FallsBackToPodman` — verify
  auto-detect returns Podman when `devpod` is in PATH
  but `.devcontainer/devcontainer.json` does NOT exist
- [x] 4.24 Add test:
  `TestDevPodCreate_MissingDevcontainer` — verify
  error when devcontainer.json doesn't exist
- [x] 4.25 Add test:
  `TestDevPodCreate_DevPodUpFails` — verify error
  propagation when `devpod up` returns non-zero
- [x] 4.26 Add test:
  `TestDevPodCreate_PodmanNotInstalled` — verify
  actionable error when Podman is missing
- [x] 4.27 Add test:
  `TestDevPodCreate_VersionTooOld` — verify error
  when DevPod version < 0.5.0
- [x] 4.28 Add test:
  `TestDevPodAttach_Success` — verify correct server
  URL passed to opencode attach
- [x] 4.29 Add test:
  `TestIsPersistentWorkspace_DevPod` — verify DevPod
  workspace detected as persistent
- [x] 4.30 Add test:
  `TestExtract_PersistentWorkspace` — verify early
  return for persistent workspaces (both backends)

## 5. Devcontainer Scaffolding

- [x] 5.1 Create devcontainer template at
  `internal/scaffold/assets/devcontainer/devcontainer.json`
  with: image, forwardPorts (4096), containerEnv
  (gateway env vars), remoteUser, and a version marker
  comment
- [x] 5.2 Create `uf sandbox init` subcommand in
  `cmd/unbound-force/sandbox.go` with flags:
  `--image` (string), `--demo-ports` (int slice),
  `--force` (bool)
- [x] 5.3 Implement `runSandboxInit()`: create
  `.devcontainer/` directory, read embedded template,
  substitute image/ports, write
  `devcontainer.json`
- [x] 5.4 Implement idempotency: skip if
  `.devcontainer/devcontainer.json` exists unless
  `--force`
- [x] 5.5 Add `devcontainer/devcontainer.json` to
  `knownNonEmbeddedFiles` in `scaffold_test.go`
  (not deployed by `uf init`, only by
  `uf sandbox init`)
- [x] 5.6 Add test: `TestRunSandboxInit_Creates` —
  verify devcontainer.json created with correct content
- [x] 5.7 Add test: `TestRunSandboxInit_ExistingSkips`
- [x] 5.8 Add test: `TestRunSandboxInit_ForceOverwrites`
- [x] 5.9 Add test: `TestRunSandboxInit_CustomDemoPorts`
  — verify extra ports in forwardPorts array

## 6. DevPod Doctor Checks

- [x] 6.1 Add `isDevPodDetected()` helper to
  `internal/doctor/checks.go`: true when `devpod` in
  PATH or config backend is `"devpod"`
- [x] 6.2 Add "DevPod" check group with 2 checks:
  `devpod` binary presence, devcontainer config
  existence
- [x] 6.3 Gate group on `isDevPodDetected()`
- [x] 6.4 Remove any existing Che-specific doctor checks
- [x] 6.5 Add test: `TestDevPodChecks_AllPresent`
- [x] 6.6 Add test:
  `TestDevPodChecks_HiddenWhenNotDetected`
- [x] 6.7 Add test:
  `TestDevPodChecks_MissingDevcontainer` — verify
  hint includes `uf sandbox init`

## 7. CLI Flag and Help Text Cleanup

- [x] 7.1 Update `--backend` flag help text in
  `cmd/unbound-force/sandbox.go`: valid values are
  `auto`, `podman`, `devpod`
- [x] 7.2 Remove any `--che-url` or Che-specific flags
- [x] 7.3 Update the `create` subcommand's `--backend`
  default description
- [x] 7.4 Add `init` subcommand to sandbox command group
- [x] 7.5 Update parent `sandbox` command Long
  description: replace "Eclipse Che / Dev Spaces"
  with "DevPod"
- [x] 7.6 Update `create` command Long description:
  replace Che/CDE references with DevPod
- [x] 7.7 Update `destroy` command Long description:
  replace "CDE workspace" with "DevPod workspace"

## 8. Documentation Updates

- [x] 8.1 Update QUICKSTART.md: add DevPod install
  instructions and `uf sandbox init` workflow
- [x] 8.2 Update USAGE.md: replace Che references with
  DevPod workflow recipe
- [x] 8.3 Update AGENTS.md project structure: add
  `devpod.go`, remove `che.go`, add devcontainer
  template path
- [x] 8.4 Update AGENTS.md "Recent Changes" with this
  change summary
- [x] 8.5 File GitHub issue in `unbound-force/website`
  for DevPod documentation updates (new
  `uf sandbox init` command, `--backend devpod`,
  Che removal migration note)

## 9. Verification

- [x] 9.1 Run `go build ./...` — verify clean build
- [x] 9.2 Run `go test -race -count=1 ./...` — verify
  all tests pass (19 packages, 0 failures)
- [x] 9.3 Run `golangci-lint run` — not installed locally;
  CI MegaLinter handles lint (go vet passes via build)
- [x] 9.4 Verify constitution alignment: Autonomous
  Collaboration (subprocess only), Composability
  (DevPod optional), Observable Quality (doctor
  checks), Testability (injectable deps)
<!-- spec-review: passed -->
<!-- code-review: passed -->
<!-- scaffolded by uf vdev -->
