## Why

`uf setup` (v0.13.0) includes a step that checks for the `mxf`
binary and reports misleading guidance when it is not found:
"Bundled with unbound-force — reinstall unbound-force to get
mxf." This is false — `mxf` was intentionally removed from the
GoReleaser build configuration and is not built or distributed
(see PR #102 revert). Similarly, `uf doctor` warns about `mxf`
being missing with an install hint that users cannot act on.

Issue #160 reports this as a setup failure on v0.13.0. While the
hard `brew install` failure was already fixed (replaced with a
PATH check), the current code still references a binary that
does not exist in any distribution channel. The `cmd/mxf/`
source code exists but is not compiled by `.goreleaser.yaml`.

Website issue unbound-force/website#56 also flags the misleading
install instruction on the public documentation.

## What Changes

Remove all `mxf` binary references from `uf setup` and
`uf doctor` core tool checks. The Mx F hero remains fully
available through its OpenCode agent (`mx-f-coach.md`) and the
hero availability check (which detects agent files, not
binaries). Only the binary distribution assumption is removed.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `uf setup`: Remove `installMxF` step. Reduce step count
  from 14 to 13. Renumber subsequent steps.
- `uf doctor`: Remove `mxf` from `coreTools` slice. The
  hero availability check for "Mx F (Manager)" via agent
  file detection is unaffected.

### Removed Capabilities

- `installMxF()` function in `internal/setup/setup.go`:
  removed because the binary is not distributed.
- `mxf` entry in `coreTools` slice in
  `internal/doctor/checks.go`: removed because doctor
  should not warn about a tool that cannot be installed.

## Impact

**Files modified**: 4 Go source files, 2 Go test files.

- `internal/setup/setup.go` — remove `installMxF` function,
  remove step 3, renumber steps 4-14 to 3-13, update step
  count constant from 14 to 13
- `internal/setup/setup_test.go` — remove
  `TestSetupRun_MxFMissing_BundledHint` and
  `TestSetupRun_MxFPresent`, remove `mxf` from LookPath
  stubs in other tests, update step count assertions
- `internal/doctor/checks.go` — remove `mxf` entry from
  `coreTools` slice
- `internal/doctor/doctor_test.go` — remove `mxf` from
  expected doctor results, update assertions

**Closes**: #160

**Website gate**: Exempt — internal toolchain change with no
user-facing documentation impact beyond fixing an already-filed
website issue (unbound-force/website#56).

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

No artifact communication is affected. The Mx F hero's agent
file remains available for OpenCode interactions. Removing a
non-existent binary check does not change inter-hero
collaboration patterns.

### II. Composability First

**Assessment**: PASS

Mx F remains independently usable through its OpenCode agent.
Removing a phantom binary check improves composability by
eliminating a broken dependency assumption.

### III. Observable Quality

**Assessment**: PASS

`uf doctor` output becomes more accurate — it no longer warns
about a tool that cannot be installed. The hero availability
check (agent file detection) continues to report Mx F status
with correct machine-parseable output.

### IV. Testability

**Assessment**: PASS

Existing tests for the removed function are removed alongside
the function. Remaining tests continue to verify `uf setup`
and `uf doctor` behavior in isolation with injected
dependencies. No shared mutable state is introduced.
