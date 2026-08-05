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

## 1. Production code changes (internal/sandbox/devpod.go)

All tasks in this group modify the same file and MUST
run sequentially.

- [x] 1.1 Update `devpodWorkspaceName` GoDoc (line 45):
  remove "Podman persistent workspace naming convention
  (D5)" — replace with "DevPod workspace naming
  convention (D5)" or similar neutral wording.

- [x] 1.2 Update `Create()` GoDoc (lines 53-59): remove
  pre-flight item "1. podman in PATH (DevPod Podman
  provider requirement)" and renumber remaining items.
  Update the GoDoc to explain the docker-aliased-as-podman
  provider model: `--provider podman` references the
  docker provider registered under the name `podman` by
  `uf setup`. Keep `--provider podman` in the command
  signature comment.

- [x] 1.3 Remove `LookPath("podman")` pre-flight block
  (lines 64-70): delete the entire block including the
  comment at lines 64-65. Per design decision D1, no
  replacement check is added.

- [x] 1.4 Add diagnostic hint to `devpod up` error path
  (line 126): append
  `"\nRun 'uf doctor' to diagnose or 'uf setup' to configure."`
  to the error message. Per design decision D5, this
  replaces the removed install hint with broader guidance.

- [x] 1.5 Verify `--provider podman` at line 97 remains
  unchanged. Per design decision D2, this is the correct
  provider **name** matching `uf setup`'s registration.
  No modification needed — this task is a verification
  checkpoint only.

## 2. Test updates (internal/sandbox/sandbox_test.go)

All tasks in this group modify the same file and MUST
run sequentially.

- [x] 2.1 Update `TestDevPodCreate_Success` (line 3645):
  keep the `--provider podman` assertion (it is correct).
  Update the `LookPath` stub to return error for `"podman"`
  to verify the pre-flight removal — the test MUST still
  pass when podman is not in PATH (primary assertion:
  success + correct provider flag with podman absent).

- [x] 2.2 Replace `TestDevPodCreate_NotInstalled`
  (line 3689) with `TestDevPodCreate_NoPodmanInPath`:
  the new test MUST assert that `Create()` succeeds
  when `LookPath("podman")` returns an error. All other
  preconditions (DevPod version, devcontainer.json) are
  satisfied. This provides regression coverage per design
  decision D3 (primary assertion: success without podman
  in PATH).

- [x] 2.3 Delete `TestDevPodCreate_PodmanNotInstalled`
  (line 3992): this test is a near-duplicate of
  `TestDevPodCreate_NotInstalled` and asserts the same
  removed pre-flight behavior ("podman not found" /
  "DevPod requires Podman"). Its behavioral coverage is
  subsumed by the new `TestDevPodCreate_NoPodmanInPath`
  from task 2.2. Per design decision D3.

- [x] 2.4 Verify `TestStart_PodmanMissing` (line 458) is
  unchanged and still passes. This test covers the
  ephemeral Podman backend's `Start()` path and is
  explicitly out of scope per design decision D4. No
  modifications permitted.

## 3. Verification

- [x] 3.1 Run `go test -race -count=1 ./internal/sandbox/...`
  and confirm all tests pass.

- [x] 3.2 Run `go vet ./internal/sandbox/...` and confirm
  no warnings.

- [x] 3.3 Run `golangci-lint run ./internal/sandbox/...`
  and confirm no warnings.

- [x] 3.4 Constitution alignment verification: confirm
  Composability First (PASS) — DevPod backend no longer
  requires podman in PATH. Confirm Testability (PASS) —
  all behavioral changes have corresponding test coverage.
  Coverage impact: the removed LookPath branch reduces
  uncovered-branch risk. Net test count is stable (two
  tests replaced by one, one test updated). The existing
  coverage ratchet is not expected to regress.
