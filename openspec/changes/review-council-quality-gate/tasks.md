## 1. Rewrite Code Review Mode step 1

- [x] 1.1 Replace the single-paragraph step 1 in
  `.opencode/command/review-council.md` Code Review
  Mode with two explicit sub-phases:

  **Phase 1a -- CI Checks (mandatory, hard gate):**
  - Read `.github/workflows/` to extract CI commands
  - Execute each command locally in order
  - If any command fails: STOP, report as CRITICAL,
    do NOT proceed to Phase 1b or step 2
  - If all pass: proceed to Phase 1b

  **Phase 1b -- Gaze Quality Analysis (conditional):**
  - Check if `gaze` is available via `which gaze`
  - If available: invoke `gaze-reporter` agent in
    `full` mode via Task tool, capture report output
  - If not available: note informational message and
    proceed to step 2
  - Phase 1b only runs in Code Review Mode, never in
    Spec Review Mode

## 2. Update step 2 with Gaze context

- [x] 2.1 Modify the Code Review Mode step 2 Divisor
  agent delegation to include Gaze report data (when
  available from Phase 1b) as a "Quality Context"
  section in each agent's review prompt. When Gaze
  data is not available, the agents receive their
  standard prompt unchanged.

## 3. Sync scaffold asset

- [x] 3.1 Copy the updated
  `.opencode/command/review-council.md` to
  `internal/scaffold/assets/opencode/command/review-council.md`
  to keep the scaffold embed in sync.

## 4. Verification

- [x] 4.1 Run `go build ./...` to verify the build
  succeeds with the updated embedded asset.

- [x] 4.2 Run `go test -race -count=1 ./...` to verify
  all tests pass including scaffold drift detection.

- [x] 4.3 Verify Spec Review Mode section is unchanged
  -- no CI or Gaze steps are present in the Spec
  Review Mode instructions.
