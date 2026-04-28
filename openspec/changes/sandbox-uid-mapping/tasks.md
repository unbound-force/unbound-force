## 1. Shared UID Mapping Args Helper

- [ ] 1.1 Add `uidMappingArgs(opts Options) []string`
  function in `config.go` that returns
  `["--userns=keep-id:uid=1000,gid=1000"]` by default
  (FR-006).
- [ ] 1.2 When `opts.UIDMap` is true, return the
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

- [ ] 2.1 In `buildRunArgs()` (`config.go`), call
  `uidMappingArgs(opts)` and append the returned
  flags to the args slice after port mapping and
  before volume mounts (FR-007).

## 3. UID Mapping Flags in buildPersistentRunArgs

- [ ] 3.1 In `buildPersistentRunArgs()` (`podman.go`),
  call `uidMappingArgs(opts)` and append the returned
  flags to the args slice (FR-008).

## 4. Options and Config Struct Updates

- [ ] 4.1 Add `UIDMap bool` field to the `Options`
  struct in `sandbox.go` with GoDoc comment:
  "UIDMap enables explicit UID/GID mapping via
  --uidmap/--gidmap flags instead of --userns=keep-id.
  Use on macOS when the Podman machine's virtiofs
  does not support keep-id UID mapping."
- [ ] 4.2 Add `UIDMap bool` field to `SandboxConfig`
  in `internal/config/config.go` with yaml tag
  `uid_map`.
- [ ] 4.3 Add `UF_SANDBOX_UIDMAP` env var override
  in `applyEnvOverrides()`: if value is "1" or
  "true", set `cfg.Sandbox.UIDMap = true`.
- [ ] 4.4 In `merge()`, add UIDMap merge logic:
  `if overlay.Sandbox.UIDMap {
  result.Sandbox.UIDMap = true }` (same pattern as
  other bool fields).
- [ ] 4.5 In `applySandboxConfig()` (`cmd/unbound-
  force/sandbox.go`), propagate config UIDMap to
  opts: `if !opts.UIDMap && cfg.Sandbox.UIDMap {
  opts.UIDMap = true }`

## 5. CLI Flag Registration

- [ ] 5.1 In `newSandboxStartCmd()` (`cmd/unbound-
  force/sandbox.go`), add `--uidmap` bool flag
  with help text: "Use explicit UID/GID mapping
  (for macOS when Podman machine virtiofs does not
  support --userns=keep-id)"
- [ ] 5.2 In `runSandboxStart()`, read the `--uidmap`
  flag and set `opts.UIDMap`.
- [ ] 5.3 In `newSandboxCreateCmd()`, add the same
  `--uidmap` flag.
- [ ] 5.4 In `runSandboxCreate()`, read the `--uidmap`
  flag and set `opts.UIDMap`.

## 6. macOS Podman Machine UID Detection

- [ ] 6.1 Add `UIDMapSupported bool` field to
  `PlatformConfig` in `detect.go` (FR-009).
- [ ] 6.2 Add `probeUIDMapping(opts Options) bool`
  function in `detect.go` that runs:
  `podman run --rm --userns=keep-id:uid=1000,gid=1000
  -v <ProjectDir>:/test:ro <Image> stat -c '%u' /test`
  via `opts.ExecCmd`. Returns true if output
  (trimmed) is "1000". Returns false on any error
  (fail-safe).
- [ ] 6.3 In `DetectPlatform()`, when `OS == "darwin"`,
  call `probeUIDMapping(opts)` and set
  `p.UIDMapSupported` from the result. On Linux,
  set `p.UIDMapSupported = true` unconditionally.
- [ ] 6.4 Verify that `DetectPlatform()` call sites
  have access to `Image` and `ProjectDir` in `opts`
  (needed for the probe container).
- [ ] 6.5 In `Start()` (`sandbox.go`), after
  `DetectPlatform()`, check: if `platform.OS ==
  "darwin" && !platform.UIDMapSupported &&
  !opts.UIDMap`, return error with the remediation
  message from design D2. The error MUST recommend
  both the Podman machine fix and the `--uidmap`
  escape hatch (FR-002).

## 7. Post-Copy chown for Persistent Workspaces

- [ ] 7.1 In `PodmanBackend.Create()` (`podman.go`),
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

- [ ] 8.1 Add `TestUIDMappingArgs_Default` test:
  verify `uidMappingArgs(Options{})` returns
  `["--userns=keep-id:uid=1000,gid=1000"]`.
- [ ] 8.2 Add `TestUIDMappingArgs_UIDMap` test:
  verify `uidMappingArgs(Options{UIDMap: true})`
  returns the 12-element `--uidmap`/`--gidmap` slice
  and does NOT contain `--userns`.
- [ ] 8.3 Add `TestBuildRunArgs_IncludesUserNS` test:
  call `buildRunArgs()` with default `Options`,
  verify the args slice contains
  `--userns=keep-id:uid=1000,gid=1000`.
- [ ] 8.4 Add `TestBuildRunArgs_UIDMapOverride` test:
  call `buildRunArgs()` with `Options{UIDMap: true}`,
  verify the args slice contains `--uidmap` and
  `--gidmap` pairs and does NOT contain `--userns`.
- [ ] 8.5 Add
  `TestBuildPersistentRunArgs_IncludesUserNS` test:
  same as 8.3 for `buildPersistentRunArgs()`.
- [ ] 8.6 Add
  `TestBuildPersistentRunArgs_UIDMapOverride` test:
  same as 8.4 for `buildPersistentRunArgs()`.
- [ ] 8.7 Add `TestProbeUIDMapping_Success` test:
  inject `ExecCmd` that returns "1000\n", verify
  `probeUIDMapping()` returns true.
- [ ] 8.8 Add `TestProbeUIDMapping_Failure` test:
  inject `ExecCmd` that returns "0\n", verify
  `probeUIDMapping()` returns false.
- [ ] 8.9 Add `TestProbeUIDMapping_Error` test:
  inject `ExecCmd` that returns error, verify
  `probeUIDMapping()` returns false (fail-safe).
- [ ] 8.10 Add
  `TestDetectPlatform_LinuxAlwaysSupported` test:
  verify `UIDMapSupported` is true on Linux
  regardless of probe result.
- [ ] 8.11 Add `TestStart_DarwinUIDMapNotSupported`
  test: set `platform.OS = "darwin"` and
  `UIDMapSupported = false`, verify `Start()` returns
  error containing "Podman machine UID mapping".
- [ ] 8.12 Add `TestStart_DarwinUIDMapOverride` test:
  set `platform.OS = "darwin"`,
  `UIDMapSupported = false`, `opts.UIDMap = true`,
  verify `Start()` proceeds without the probe error
  (override bypasses detection).
- [ ] 8.13 Add `TestPodmanCreate_ChownAfterCopy` test:
  verify `PodmanBackend.Create()` calls
  `podman exec <ctr> chown -R dev:dev /workspace`
  after `podman cp` and before `waitForHealth`.
- [ ] 8.14 Add `TestPodmanCreate_ChownFailure` test:
  inject `ExecCmd` that fails on the `chown` call,
  verify partial cleanup (container and volume
  removed) and error returned containing "failed to
  fix permissions".
- [ ] 8.15 Update existing tests that call
  `buildRunArgs()` or `buildPersistentRunArgs()`
  directly to account for the new
  `--userns=keep-id:uid=1000,gid=1000` argument in
  the args slice. Update assertion counts or slice
  index checks as needed.

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
