## 1. Argument-First Parsing Gate (FR-001, D1)

- [x] 1.1 In `.opencode/command/review-pr.md`, add an
  "Argument Parsing" section between the "Arguments"
  block and "## Execution Steps". The section MUST
  instruct: "Check the user's message for a PR number
  argument. If present, set `PR_NUMBER` to that value
  immediately. All subsequent steps use `<PR_NUMBER>`
  — no auto-detection commands are needed or
  permitted."
- [x] 1.2 In Step 1 "Resolve PR Number", replace the
  current "If a PR number was provided" line with:
  "If `PR_NUMBER` was already set from the argument:
  skip this step entirely. Do NOT run `gh pr view`,
  `git branch --show-current`, or any branch/PR
  detection commands."
- [x] 1.3 Keep the existing auto-detection logic for
  the "no argument" path unchanged: `gh pr view
  --json number --jq '.number'` with the STOP error.

## 2. Execution Mode Check (FR-002, D4)

- [x] 2.1 In Step 0 "Prerequisites", after the
  `gh auth status` check, add a "Mode Check"
  subsection.
- [x] 2.2 The mode check MUST test execution
  capability with a harmless command. Uses
  `echo "mode-check-ok"` as a runtime-agnostic
  probe — tests whether the agent runtime allows
  command execution at all, not whether specific
  tools exist.
- [x] 2.3 If the command cannot be executed (plan
  mode, read-only mode), the instruction MUST direct
  the agent to STOP with this message:
  "This review requires running local tools (build,
  test, lint) to verify the PR. I am currently in
  plan/read-only mode which prevents executing these
  checks. Switch to a mode that allows command
  execution and re-invoke `/review-pr <N>`."
- [x] 2.4 Add an explicit instruction: "Do NOT
  proceed with a partial review that skips local tool
  execution. The local tool results are the foundation
  of the review — without them, AI-only findings lack
  verification and the review does not meet the
  command's quality standard."

## 3. CI Coverage Matrix (FR-003, D5)

- [x] 3.1 In Step 4, before the "Execution" table,
  add a "CI coverage check" subsection marked as
  mandatory.
- [x] 3.2 The instruction MUST require the agent to
  build and display a coverage matrix with columns:
  `Local tool | CI check that covers it | CI status |
  Run locally?`
- [x] 3.3 Document the decision rules:
  - CI status PASS → skip locally ("No")
  - CI status FAIL → skip locally, captured in
    Step 3a ("No")
  - CI status NONE (no matching check) → MUST run
    locally ("Yes")
  - No CI checks reported at all → MUST run ALL
    detected local tools ("Yes" for every row)
- [x] 3.4 Move the existing "Execution" table to
  appear after the coverage matrix, with a qualifier:
  "Run only the tools marked 'Yes' in the matrix
  above."

## 4. Save-and-Navigate Diff Handling (FR-004, D2)

- [x] 4.1 In Step 5 "Fetch Diff (Scoped)", replace
  the current "Large diff handling" bullet list with
  a numbered procedure:
  1. Save the full diff once:
     `gh pr diff <PR_NUMBER> > /tmp/pr<N>.diff`
     (or note: "The tool runtime auto-saves truncated
     output to a file — use that path if available.")
  2. Find file boundaries:
     `grep -n '^diff --git' /tmp/pr<N>.diff`
  3. Read specific file sections using offset/limit
     on the saved file.
  4. Skip: lock files (`package-lock.json`, `go.sum`,
     `yarn.lock`, `bun.lock`), auto-generated files
     (`*.pb.go`, `vendor/`), binary files, CRAP
     baselines (`.gaze/baseline.json`).
- [x] 4.2 Add a "Do NOT attempt" list at the end of
  Step 5:
  - `gh pr diff <N> -- <path>` (unsupported, fails)
  - `git show <remote>/<branch>:<path>` (PR branch
    may not be on any configured remote)
  - `git fetch <remote> <branch>` (PR may come from
    a fork or push directly to PR refs)

## 5. PR Branch Access via GitHub API (FR-005, D3)

- [x] 5.1 Add a new subsection after Step 5 titled
  "Accessing full file contents from the PR branch".
- [x] 5.2 Document the GitHub API method:
  ```
  gh api repos/{owner}/{repo}/contents/<path>\
    ?ref=<headRefName> --jq '.content' | base64 -d
  ```
  Note that `<headRefName>` comes from Step 2
  metadata. Added error handling: fallback to saved
  diff on 404, 403, or empty content (>1 MB files).
- [x] 5.3 Add a DO NOT guard: "For accessing files
  on the PR branch, the agent MUST use `gh api`
  exclusively. Any `git` subcommand targeting the
  PR's head ref is prohibited." (Positive allowlist
  per Adversary review feedback.)

## 6. PR-Introduced Spec Detection (FR-006, D6)

- [x] 6.1 In Step 6 "Locate Associated Specification",
  after the existing filesystem check bullets, add:
  "If not found locally, check the PR's changed file
  list (from Step 2 metadata) for spec artifacts.
  The spec may be introduced by the PR itself."
- [x] 6.2 Add instruction to read spec content from
  the saved diff (Step 5) rather than from the
  filesystem when the spec is found only in the
  changed file list.

## 7. Scaffold Asset Synchronization

- [x] 7.1 Copy the updated `.opencode/command/
  review-pr.md` to `internal/scaffold/assets/opencode/
  command/review-pr.md` to keep the scaffold copy
  synchronized.
- [x] 7.2 Verify the two files are identical:
  `diff .opencode/command/review-pr.md internal/
  scaffold/assets/opencode/command/review-pr.md`

## 8. Verification

- [x] 8.1 Re-read the final `.opencode/command/
  review-pr.md` end-to-end to verify all six fixes
  are applied and the step numbering is consistent.
  Fixed pre-existing step numbering in Step 10
  (duplicate "4." → "5.").
- [x] 8.2 Verify the Argument Parsing section appears
  before Step 0. Confirmed at line 13.
- [x] 8.3 Verify the Mode Check appears in Step 0
  after `gh auth status`. Confirmed at line 41.
- [x] 8.4 Verify the CI Coverage Matrix appears in
  Step 4 before the Execution table. Confirmed at
  line 157.
- [x] 8.5 Verify the DO NOT lists appear in Steps 5
  and the new PR branch access section. Confirmed at
  lines 240 and 262.
- [x] 8.6 Verify the PR-introduced spec detection
  appears in Step 6. Confirmed at line 281.
- [x] 8.7 Assess Website Documentation Gate: this
  change modifies internal agent command instructions
  only. No user-facing CLI commands, flags, or
  workflows are changed. Exempt per AGENTS.md
  criteria.
- [x] 8.8 Update AGENTS.md Recent Changes section
  with a summary of this change following the
  established pattern: `- opsx/review-pr-reliability:
  <description>. Modified files: list. No Go code
  changes. N tasks completed.`
<!-- spec-review: passed -->
<!-- code-review: passed -->
