## 1. Remove dewey init/index from setup

- [x] 1.1 Remove the `initDewey()` function from
  `internal/setup/setup.go`

- [x] 1.2 Remove the `indexDewey()` function from
  `internal/setup/setup.go`

- [x] 1.3 Remove steps 13-14 (dewey init + dewey
  index) from the `Run()` function in
  `internal/setup/setup.go`, including both the
  normal path and the `deweyInitResult.action !=
  "failed"` conditional

- [x] 1.4 Renumber all step progress messages from
  `[N/15]` to `[N/13]` in `internal/setup/setup.go`
  (old step 15 `runUnboundInit` becomes step 13)

- [x] 1.5 Remove the comment at the `runUnboundInit()`
  call site that documents the dewey duplication
  (lines 787-790 area) in `internal/setup/setup.go`

## 2. Add force re-index to scaffold

- [x] 2.1 In `initSubTools()` in
  `internal/scaffold/scaffold.go`, add a new block
  after the existing dewey init/index block: when
  `.dewey/` already exists AND `opts.Force` is true
  AND `dewey` is in PATH, print progress message
  `"  Re-indexing Dewey sources..."`, run
  `opts.ExecCmd("dewey", "index")`, and append a
  `subToolResult` with name `"dewey index"` and
  action `"re-indexed"` (or `"failed"` on error)

## 3. Update setup tests

- [x] 3.1 Remove or update tests in
  `internal/setup/setup_test.go` that test
  `initDewey()` and `indexDewey()` behavior (these
  functions no longer exist)

- [x] 3.2 Update any step-count assertions in
  `internal/setup/setup_test.go` that reference
  15 steps to 13 steps

## 4. Add scaffold tests

- [x] 4.1 Write `TestInitSubTools_DeweyForceReindex`
  in `internal/scaffold/scaffold_test.go`: `.dewey/`
  exists, Force=true, dewey in LookPath → verify
  `dewey index` is called via ExecCmd and result
  action is `"re-indexed"`

- [x] 4.2 Write `TestInitSubTools_DeweyExistsNoForce`
  in `internal/scaffold/scaffold_test.go`: `.dewey/`
  exists, Force=false → verify dewey init and index
  are NOT called, no dewey-related results

## 5. Verification

- [x] 5.1 Run `go build ./...` to verify clean build

- [x] 5.2 Run `go test -race -count=1 ./...` to verify
  all tests pass

- [x] 5.3 Run `go vet ./...` to verify no vet warnings

- [x] 5.4 Verify constitution alignment: Composability
  (dewey remains optional), Observable Quality
  (progress messages preserved), Testability
  (injectable ExecCmd/LookPath used)
