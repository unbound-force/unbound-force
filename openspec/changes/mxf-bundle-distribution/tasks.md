## 1. GoReleaser: Bundle mxf in packages

- [x] 1.1 Add `bin.install "mxf"` to the `brews:` install
  block in `.goreleaser.yaml`
- [x] 1.2 Add `mxf` binary to the `nfpms:` contents in
  `.goreleaser.yaml` (entry for `/usr/bin/mxf`)

## 2. Setup: Simplify installMxF

- [x] 2.1 Replace `installMxF()` in
  `internal/setup/setup.go` with a PATH verification
  that checks `LookPath("mxf")` and returns "already
  installed" if found or "not found — bundled with
  unbound-force" if absent. Remove the
  `brew install unbound-force/tap/mxf` call
- [x] 2.2 Update existing `TestSetupRun_MxFMissing_BrewInstall`
  and `TestSetupRun_MxFNoHomebrew` tests in
  `internal/setup/setup_test.go` to match the new
  behavior (no brew install, just PATH check)
- [x] 2.3 Update `TestSetupRun_AllMissing` expected
  commands list to remove the
  `brew install unbound-force/tap/mxf` entry

## 3. Doctor: Update install hints

- [x] 3.1 Update `homebrewInstallCmd("mxf")` in
  `internal/doctor/environ.go` to return
  `brew install unbound-force/tap/unbound-force`
  (references the parent package)
- [x] 3.2 Update `genericInstallCmd("mxf")` or add a
  case for `mxf` that returns
  `Bundled with unbound-force`

## 4. Verification

- [x] 4.1 Run `go test -race -count=1 ./...` and verify
  all tests pass
  (setup and doctor packages pass; scaffold/schemas
  failures are pre-existing on main, unrelated to this
  change)
- [x] 4.2 Verify Composability First: `mxf` remains
  independently executable with no runtime dependency
  on `unbound-force`
  (mxf binary has no imports from cmd/unbound-force;
  bundling is distribution-only, not runtime coupling)
- [x] 4.3 Verify Testability: all changes use injectable
  dependencies and tests run without network access
  (installMxF uses injected LookPath; 2 new tests
  pass without network access)
