## 1. Add recursive: false to disk-org

- [x] 1.1 Add `b.WriteString("      recursive: false\n")`
  after the `path` line for the `disk-org` entry in
  `generateDeweySources()` in
  `internal/scaffold/scaffold.go` (around line 1056)

## 2. Add force parameter to generateDeweySources

- [x] 2.1 Add a `force bool` parameter to
  `generateDeweySources(opts *Options, force bool)` in
  `internal/scaffold/scaffold.go`

- [x] 2.2 When `force` is true, skip the
  `isDefaultSourcesConfig` check (bypass the early
  return at line 914-920) in
  `internal/scaffold/scaffold.go`

- [x] 2.3 Update the first-run call site in
  `initSubTools()` to pass `false`:
  `generateDeweySources(opts, false)` in
  `internal/scaffold/scaffold.go`

## 3. Call generateDeweySources from force block

- [x] 3.1 In the `opts.Force` block of `initSubTools()`
  (around line 724), call
  `generateDeweySources(opts, true)` BEFORE the
  `dewey index` call, append the result to the
  results slice, in `internal/scaffold/scaffold.go`

## 4. Update tests

- [x] 4.1 Update `TestGenerateDeweySources_SiblingsDetected`
  in `internal/scaffold/scaffold_test.go`: add assertion
  that the `disk-org` section contains
  `recursive: false`. Also update the function call to
  pass `false` for the new force parameter.

- [x] 4.2 Update `TestGenerateDeweySources_NoSiblings`
  in `internal/scaffold/scaffold_test.go`: same
  assertions for `recursive: false` on `disk-org` and
  updated function call.

- [x] 4.3 Update `TestGenerateDeweySources_AlreadyCustomized`
  in `internal/scaffold/scaffold_test.go`: update
  function call to pass `false`.

- [x] 4.4 Write `TestGenerateDeweySources_ForceOverwritesCustom`
  in `internal/scaffold/scaffold_test.go`: create a
  customized `sources.yaml` (>1 source entry), call
  `generateDeweySources(opts, true)`, verify the file
  is overwritten with auto-detected config including
  `recursive: false` on `disk-org`.

- [x] 4.5 Update `TestInitSubTools_DeweyForceReindex`
  in `internal/scaffold/scaffold_test.go`: verify that
  `generateDeweySources` is called before `dewey index`
  in the force path (check that sources.yaml is updated).

## 5. Verification

- [x] 5.1 Run `go build ./...` to verify clean build

- [x] 5.2 Run `go test -race -count=1 ./...` to verify
  all tests pass

- [x] 5.3 Run `golangci-lint run` to verify lint clean
