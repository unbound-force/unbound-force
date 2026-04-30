## 1. Scaffold Engine — OpenPackage Delegation

- [x] 1.1 Add `ExecCmdInDir` field to `scaffold.Options`
  with signature
  `func(dir string, name string, args ...string) ([]byte, error)`
  and default implementation `defaultExecCmdInDir()` that
  wraps `exec.Command` with `cmd.Dir` set
- [x] 1.2 Add `SkipOpenPackage` boolean field to
  `scaffold.Options` that disables opkg delegation
- [x] 1.3 Implement `openPackageInstall()` — detect opkg
  via `LookPath`, select packages based on `DivisorOnly`
  flag, run `opkg install` per package via `ExecCmdInDir`
  in `TargetDir` using absolute paths to embedded package
  source trees, return `[]subToolResult`; report failing
  package name with re-run hint on error
- [x] 1.4 Add `OpenPackagePlatforms` field to `scaffold.Options`;
  append `--platforms <value>` to each `opkg install`
  invocation when set
- [x] 1.5 Wire `openPackageInstall()` into `Run()` after
  the `fs.WalkDir` loop (post-walk overlay model: embedded
  baseline always written, opkg adds harness-specific
  routing on top)
- [x] 1.6 Include opkg results in `subResults` slice
  passed to `printSummary()`
- [x] 1.7 Default `ExecCmdInDir` in `Run()` (nil check
  alongside existing LookPath/ExecCmd/ReadFile/WriteFile
  defaults)

## 2. OpenPackage Source Trees

- [x] 2.1 Create `.openpackage/packages/review-council/openpackage.yml`
  manifest with name, version, description, keywords,
  author, license
- [x] 2.2 Create `.openpackage/packages/review-council/README.md` with
  install instructions, persona table, contents list
- [x] 2.3 Create `.openpackage/packages/review-council/mcp.jsonc` with
  Dewey MCP config template
- [x] 2.4 Populate `.openpackage/packages/review-council/agents/review-council/`
  with all 9 Divisor agent files (divisor-adversary through
  divisor-testing, plus curator/envoy/herald/scribe)
- [x] 2.5 Populate `.openpackage/packages/review-council/commands/review-council/`
  with review-council.md and review-pr.md
- [x] 2.6 Populate `.openpackage/packages/review-council/rules/review-council/`
  with default.md, default-custom.md, severity.md
- [x] 2.7 Create `.openpackage/packages/workflows/openpackage.yml`
  manifest with dependency on `@unbound-force/review-council`
  at `^0.1.0`
- [x] 2.8 Create `.openpackage/packages/workflows/README.md`
- [x] 2.9 Populate `.openpackage/packages/workflows/agents/workflows/`
  with constitution-check.md
- [x] 2.10 Populate `.openpackage/packages/workflows/commands/workflows/`
  with all Speckit and OpenSpec command files

## 3. Setup — installOpkg Step

- [x] 3.1 Implement `installOpkg()` in `internal/setup/setup.go`
  following the `installGaze()` pattern: check LookPath,
  try `brew install openpackage`, return "skipped" on
  Homebrew absence or formula failure
- [x] 3.2 Wire as step 15/15 in `Run()` with
  `shouldSkipTool("opkg")` guard

## 4. Documentation

- [x] 4.1 Update AGENTS.md "Recent Changes" section with
  opkg-delegate entry
- [x] 4.2 File Website Documentation Gate issue for
  `unbound-force/website` — new `uf init --platforms` flag
  and `uf setup` step 15 are user-facing (issue #TBD)

## 5. Verification

- [x] 5.1 Run `make check` (build + test + vet + lint)
- [x] 5.2 Verify existing scaffold tests pass with
  `SkipOpenPackage: true`
- [x] 5.3 Verify constitution alignment: Autonomous
  Collaboration (self-describing artifacts, no runtime
  coupling), Composability First (opkg optional, fallback
  works), Observable Quality (summary reports delegation
  outcome), Testability (injectable ExecCmdInDir,
  SkipOpenPackage flag)
