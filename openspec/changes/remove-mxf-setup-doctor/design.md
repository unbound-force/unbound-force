## Context

`uf setup` step 3 calls `installMxF()` which checks for the
`mxf` binary in PATH. When missing (which is always, since the
binary is not distributed), it returns a "not found" result with
a misleading hint. `uf doctor` includes `mxf` in the `coreTools`
slice with `recommended: true`, producing an unresolvable
`[WARN]` on every run.

The Mx F hero is fully functional through its OpenCode agent
(`mx-f-coach.md`). The hero availability check in
`checkHeroAvailability()` detects Mx F via agent file presence,
not via the `mxf` binary. Removing the binary check has no
impact on Mx F functionality.

The proposal's constitution alignment confirms all four
principles are satisfied (PASS on I-IV). This design preserves
Composability by keeping Mx F independently usable through its
agent, and Observable Quality by removing a false warning from
doctor output.

## Goals / Non-Goals

### Goals

- Remove the `installMxF` function and setup step 3 entirely
- Remove `mxf` from the `coreTools` slice in doctor checks
- Update step numbering (14 → 13) and all related tests
- Close issue #160

### Non-Goals

- Building or distributing the `mxf` binary — that is a
  separate decision tracked outside this change
- Removing `cmd/mxf/` source code — the source may be built
  and distributed in the future
- Modifying the Mx F hero availability check — agent file
  detection is correct and should remain
- Updating website documentation — already tracked by
  unbound-force/website#56

## Decisions

### D1: Remove entirely, not downgrade to informational

The `installMxF` step could be downgraded to a skip or info
message instead of removed. However, a step that always
produces the same "not found" result with no remediation path
adds noise without value. Removing it is cleaner.

### D2: Keep hero availability check unchanged

The `checkHeroAvailability()` function at `checks.go:606`
detects Mx F through `orchestration.DetectHeroes()`, which
checks for agent files (`mx-f-coach.md`). This is the correct
detection method since Mx F operates as an OpenCode agent, not
a standalone binary. This check stays.

### D3: Remove `mxf` from LookPath stubs in unrelated tests

Several test functions (e.g., full-stack setup tests) include
`mxf` in their LookPath stubs to simulate a complete
environment. Since the setup code no longer checks for `mxf`,
these stub entries become dead code and should be removed for
clarity.

## Test Strategy

All changes are removals of existing code. The test strategy is:

1. Remove `TestSetupRun_MxFMissing_BundledHint` and
   `TestSetupRun_MxFPresent` (test the removed function)
2. Remove `mxf` from LookPath stubs in remaining tests
3. Update step count assertions (14 → 13)
4. Run `make check` to verify no regressions

No new tests are needed — this is a pure removal.

## Risks / Trade-offs

### R1: Future mxf distribution requires re-adding

If the `mxf` binary is later added to `.goreleaser.yaml`,
the setup and doctor code will need to be re-introduced.
This is acceptable — adding a binary to the distribution is
a feature change that would go through its own spec workflow.

### R2: Users who somehow have mxf installed see no change

If a user has a locally-built `mxf` binary, `uf doctor` will
no longer check it and `uf setup` will no longer acknowledge
it. This is a non-issue since no official distribution path
exists.
