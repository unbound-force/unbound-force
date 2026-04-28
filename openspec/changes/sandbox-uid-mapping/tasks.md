## 1. Shared UID Mapping Args Helper

- [x] 1.1 Add `uidMappingArgs(opts Options) []string`
  function in `config.go` that returns
  `["--userns=keep-id:uid=1000,gid=1000"]` by default
  (FR-006).
- [x] 1.2 When `opts.UIDMap` is true, return the
  12-element explicit mapping slice instead:
  `--uidmap 1000:0:1`,
  `--uidmap 0:1:1000`,
  `--uidmap 1001:1001:64536`,
  `--gidmap 1000:0:1`,
  `--gidmap 0:1:1000`,
  `--gidmap 1001:1001:64536`
  (each flag-value as a separate arg pair). Do NOT
  include `--userns` when `UIDMap` is true.

## 2. UID Mapping Flags in buildRunArgs (Ephemeral)

- [x] 2.1 In `buildRunArgs()` (`config.go`), call
  `uidMappingArgs(opts)` and append the returned
  flags to the args slice after port mapping and
  before volume mounts (FR-007).

## 3. UID Mapping Flags in buildPersistentRunArgs

- [x] 3.1 In `buildPersistentRunArgs()` (`podman.go`),
  call `uidMappingArgs(opts)` and append the returned
  flags to the args slice (FR-008).

## 4. Options and Config Struct Updates

- [x] 4.1 Add `UIDMap bool` field to the `Options`
  struct in `sandbox.go` with GoDoc comment:
  "UIDMap enables explicit UID/GID mapping via
  --uidmap/--gidmap flags instead of --userns=keep-id.
  Use on macOS when the Podman machine's virtiofs
  does not support keep-id UID mapping."
- [x] 4.2 Add `UIDMap bool` field to `SandboxConfig`
  in `internal/config/config.go` with yaml tag
  `uid_map`.
- [x] 4.3 Add `UF_SANDBOX_UIDMAP` env var override
  in `applyEnvOverrides()`: if value is "1" or
  "true", set `cfg.Sandbox.UIDMap = true`.
- [x] 4.4 In `merge()`, add UIDMap merge logic:
  `if overlay.Sandbox.UIDMap {
  result.Sandbox.UIDMap = true }` (same pattern as
  other bool fields).
- [x] 4.5 In `applySandboxConfig()` (`cmd/unbound-
  force/sandbox.go`), propagate config UIDMap to
  opts: `if !opts.UIDMap && cfg.Sandbox.UIDMap {
  opts.UIDMap = true }`

## 5. CLI Flag Registration

- [x] 5.1 In `newSandboxStartCmd()` (`cmd/unbound-
  force/sandbox.go`), add `--uidmap` bool flag
  with help text: "Use explicit UID/GID mapping
  (for macOS when Podman machine virtiofs does not
  support --userns=keep-id)"
- [x] 5.2 In `runSandboxStart()`, read the `--uidmap`
  flag and set `opts.UIDMap`.
- [x] 5.3 In `newSandboxCreateCmd()`, add the same
  `--uidmap` flag.
- [x] 5.4 In `runSandboxCreate()`, read the `--uidmap`
  flag and set `opts.UIDMap`.

## 6. Platform Detection, Probes, and Guards

- [x] 6.1 Add `UIDMapSupported bool` field to
  `PlatformConfig` in `detect.go` (FR-009).
- [x] 6.2 Add `probeUIDMapping(opts Options) bool`
  function in `detect.go` that runs:
  `podman run --rm --entrypoint stat
  --userns=keep-id:uid=1000,gid=1000
  -v <ProjectDir>:/test:ro busybox:latest
  -c '%u' /test`
  via `opts.ExecCmd`. Uses `busybox:latest` (not the
  full dev image) with explicit `--entrypoint stat`
  to prevent entrypoint execution. Returns true if
  output (trimmed) is "1000". Returns false on any
  error (fail-safe). Print progress message
  "Checking Podman machine UID mapping..." to
  `opts.Stderr` before the probe.
- [x] 6.3 In `DetectPlatform()`, when `OS == "darwin"`,
  call `probeUIDMapping(opts)` and set
  `p.UIDMapSupported` from the result. On Linux,
  set `p.UIDMapSupported = true` unconditionally.
  Prerequisite: `DefaultConfig(opts)` must be called
  before `DetectPlatform(opts)` to ensure `Image` is
  populated.
- [x] 6.4 Add `Platform *PlatformConfig` field to
  `Options` in `sandbox.go` (FR-012). When non-nil,
  `Start()` uses it instead of calling
  `DetectPlatform()`. This allows tests to override
  `runtime.GOOS` and exercise macOS logic on Linux CI.
- [x] 6.5 In `Start()` (`sandbox.go`), use
  `opts.Platform` if non-nil, otherwise call
  `DetectPlatform(opts)`. Then check: if
  `platform.OS == "darwin" &&
  !platform.UIDMapSupported && !opts.UIDMap`, return
  error with the remediation message from design D2.
  The error MUST recommend both the Podman machine
  fix and the `--uidmap` escape hatch (FR-002).
- [x] 6.6 Add `parsePodmanVersion(opts Options)
  (major, minor int, err error)` function in
  `detect.go` that runs `podman --version` via
  `opts.ExecCmd`, parses "podman version X.Y.Z",
  returns major and minor. In `Start()`, call before
  using `--userns=keep-id:uid=1000,gid=1000`. If
  version < 4.3, return error: "Podman >= 4.3
  required for --userns=keep-id:uid=N,gid=N.
  Current: X.Y.Z" (FR-010).
- [x] 6.7 Add `isRootlessPodman(opts Options) bool`
  function in `detect.go` that runs
  `podman info --format '{{.Host.Security.Rootless}}'`
  via `opts.ExecCmd`. Returns true if output is
  "true". In `Start()`, when `opts.UIDMap` is true,
  check `isRootlessPodman()`. If false, return error:
  "--uidmap is only safe under rootless Podman"
  (FR-011).

## 7. Post-Copy chown for Persistent Workspaces

- [x] 7.1 In `PodmanBackend.Create()` (`podman.go`),
  after the `podman cp` step succeeds and before
  `waitForHealth()`, add:
  - Print progress: `"Fixing workspace permissions..."`
  - Run: `podman exec <ctrName> chown -R dev:dev
    /workspace` via `opts.ExecCmd`
  - On failure: clean up container and volume (same
    partial failure pattern as existing code), return
    error containing "failed to fix permissions"
    (FR-004).

## 8. Tests

- [x] 8.1 Add `TestUIDMappingArgs_Default` test:
  verify `uidMappingArgs(Options{})` returns
  `["--userns=keep-id:uid=1000,gid=1000"]`.
- [x] 8.2 Add `TestUIDMappingArgs_UIDMap` test:
  verify `uidMappingArgs(Options{UIDMap: true})`
  returns the 12-element `--uidmap`/`--gidmap` slice
  and does NOT contain `--userns`.
- [x] 8.3 Add `TestBuildRunArgs_IncludesUserNS` test:
  call `buildRunArgs()` with default `Options`,
  verify the args slice contains
  `--userns=keep-id:uid=1000,gid=1000`.
- [x] 8.4 Add `TestBuildRunArgs_UIDMapOverride` test:
  call `buildRunArgs()` with `Options{UIDMap: true}`,
  verify the args slice contains `--uidmap` and
  `--gidmap` pairs and does NOT contain `--userns`.
- [x] 8.5 Add
  `TestBuildPersistentRunArgs_IncludesUserNS` test:
  same as 8.3 for `buildPersistentRunArgs()`.
- [x] 8.6 Add
  `TestBuildPersistentRunArgs_UIDMapOverride` test:
  same as 8.4 for `buildPersistentRunArgs()`.
- [x] 8.7 Add `TestProbeUIDMapping_Success` test:
  inject `ExecCmd` that returns "1000\n", verify
  `probeUIDMapping()` returns true.
- [x] 8.8 Add `TestProbeUIDMapping_Failure` test:
  inject `ExecCmd` that returns "0\n", verify
  `probeUIDMapping()` returns false.
- [x] 8.9 Add `TestProbeUIDMapping_Error` test:
  inject `ExecCmd` that returns error, verify
  `probeUIDMapping()` returns false (fail-safe).
- [x] 8.10 Add `TestProbeUIDMapping_UnexpectedOutput`
  test: inject `ExecCmd` that returns "nobody\n",
  verify returns false (fail-safe for non-numeric
  output).
- [x] 8.11 Add `TestProbeUIDMapping_EmptyOutput` test:
  inject `ExecCmd` that returns "", verify returns
  false.
- [x] 8.12 Add
  `TestDetectPlatform_LinuxAlwaysSupported` test:
  call `DetectPlatform()` on any platform, verify
  `UIDMapSupported` is true when `runtime.GOOS` is
  "linux" (skip on non-Linux).
- [x] 8.13 Add `TestStart_DarwinUIDMapNotSupported`
  test: inject `opts.Platform = &PlatformConfig{
  OS: "darwin", UIDMapSupported: false}`, verify
  `Start()` returns error containing "Podman machine
  UID mapping". Works on any CI platform via
  Platform injection (FR-012).
- [x] 8.14 Add `TestStart_DarwinUIDMapOverride` test:
  inject `opts.Platform = &PlatformConfig{
  OS: "darwin", UIDMapSupported: false}` and
  `opts.UIDMap = true`, verify `Start()` proceeds
  without the probe error (override bypasses
  detection).
- [x] 8.15 Add `TestParsePodmanVersion_Valid` test:
  inject `ExecCmd` returning "podman version 4.9.3",
  verify major=4, minor=9.
- [x] 8.16 Add `TestParsePodmanVersion_TooOld` test:
  inject `ExecCmd` returning "podman version 4.2.1",
  verify Start() returns error containing
  "Podman >= 4.3 required".
- [x] 8.17 Add `TestIsRootlessPodman_True` test:
  inject `ExecCmd` returning "true", verify
  `isRootlessPodman()` returns true.
- [x] 8.18 Add `TestIsRootlessPodman_False` test:
  inject `ExecCmd` returning "false", verify returns
  false.
- [x] 8.19 Add `TestStart_UIDMapRejectedUnderRootful`
  test: inject `opts.UIDMap = true` and
  `isRootlessPodman` returning false, verify
  `Start()` returns error containing "only safe
  under rootless Podman".
- [x] 8.20 Add `TestApplyEnvOverrides_UIDMap` test:
  verify `UF_SANDBOX_UIDMAP=1` sets
  `cfg.Sandbox.UIDMap = true`, and
  `UF_SANDBOX_UIDMAP=0` leaves it false.
- [x] 8.21 Add `TestConfigMerge_UIDMap` test:
  verify `merge()` propagates `UIDMap = true` from
  overlay to result.
- [x] 8.22 Add `TestApplySandboxConfig_UIDMap` test:
  verify `applySandboxConfig()` sets `opts.UIDMap`
  from config when CLI flag is not set, and CLI
  flag takes precedence.
- [x] 8.23 Add `TestPodmanCreate_ChownAfterCopy` test:
  verify `PodmanBackend.Create()` calls
  `podman exec <ctr> chown -R dev:dev /workspace`
  after `podman cp` and before `waitForHealth`.
  Also verify `opts.Stderr` contains "Fixing
  workspace permissions...".
- [x] 8.24 Add `TestPodmanCreate_ChownFailure` test:
  inject `ExecCmd` that fails on the `chown` call,
  verify partial cleanup (container and volume
  removed) and error returned containing "failed to
  fix permissions".
- [x] 8.25 Update existing tests that call
  `buildRunArgs()` or `buildPersistentRunArgs()`
  directly to account for the new
  `--userns=keep-id:uid=1000,gid=1000` argument in
  the args slice. Specifically update:
  `TestBuildRunArgs_Isolated`,
  `TestBuildRunArgs_Direct`,
  `TestBuildRunArgs_NoParentNoWorkdir`,
  `TestBuildRunArgs_RootFallbackNoWorkdir`,
  `TestBuildRunArgs_SELinux`,
  `TestBuildRunArgs_CustomImage`,
  `TestPodmanCreate_HappyPath`,
  `TestPodmanCreate_WithDemoPorts`.
  For each, verify `--userns=keep-id:uid=1000,
  gid=1000` appears in the args slice at an index
  before the image argument.

## 9. Documentation and Verification

- [ ] 9.1 Update AGENTS.md Recent Changes section
  with a summary of this change.
- [ ] 9.2 Update QUICKSTART.md macOS section with
  Podman machine configuration requirements for
  `uf sandbox`. Include the
  `podman machine init --rootful=false` steps.
- [ ] 9.3 Assess Website Documentation Gate: this
  change adds a new CLI flag (`--uidmap`) and changes
  sandbox behavior. Create a GitHub issue in
  `unbound-force/website` to document the UID mapping
  requirements and `--uidmap` flag.
- [ ] 9.4 Run `make check` (build, test, vet, lint)
  and verify all checks pass.
- [ ] 9.5 Verify constitution alignment:
  - Principle III (Observable Quality): confirm
    macOS detection error is actionable with
    remediation steps
  - Principle IV (Testability): confirm all new
    detection and flag generation is testable via
    injected dependencies without external services
<!-- spec-review: passed -->
<!-- code-review: passed -->
