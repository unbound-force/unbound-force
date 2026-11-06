<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file --
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Authoring-time upstream verification (CI-002)

- [x] 1.1 Resolve the Gaze v2.3.0 tag commit via the GitHub
  API and confirm it exactly equals
  `8031d41943b06407cb92322791b177b41b7b2775`:
  `gh api repos/unbound-force/gaze/commits/8031d41943b06407cb92322791b177b41b7b2775 --jq .sha`
  and
  `gh api repos/unbound-force/gaze/tags --jq '.[] | select(.name=="v2.3.0") | .commit.sha'`.
  On mismatch, stop and report per the Gatekeeping rule.
- [x] 1.2 Confirm `reusable_gaze_ai_report.yml` is
  byte-identical between `8bdfb59f8b10de854ba289e5e545d1eb9fee0f56`
  and the target SHA (line count plus content SHA-256 via the
  GitHub contents API). On any difference, stop and report --
  the zero-behavior-change premise would not hold.

## 2. Apply the re-pin

- [x] 2.1 Edit `.github/workflows/ci_local.yml:101`: replace
  `...reusable_gaze_ai_report.yml@8bdfb59f8b10de854ba289e5e545d1eb9fee0f56`
  with
  `...reusable_gaze_ai_report.yml@8031d41943b06407cb92322791b177b41b7b2775  # v2.3.0`.
  One line, one file; no other workflow content changes.

## 3. Verification

- [x] 3.1 [P] Confirm the working diff is exactly the one
  repinned line and that `8bdfb59` has no remaining functional
  references (`uses:` pins, scripts, configs) via repo-wide
  grep; the rollback documentation in the change artifacts
  and CHANGELOG deliberately retains the old SHA.
- [x] 3.2 [P] Validate the edited workflow still parses as
  YAML (for example `python3 -c "import
  yaml,sys; yaml.safe_load(open('.github/workflows/ci_local.yml'))"`)
  and run the project's declared local checks for CI parity.
- [x] 3.3 Verify constitution alignment and Gatekeeping
  conditions on the final diff: the pin remains a full
  40-character immutable SHA with an accurate tag comment
  (Principle V), the pinact gate is strengthened not bypassed,
  and rollback to
  `8bdfb59f8b10de854ba289e5e545d1eb9fee0f56` is documented
  (see task 4.1).

## 4. PR record

- [x] 4.1 Ensure the fix PR body records the previous SHA
  `8bdfb59f8b10de854ba289e5e545d1eb9fee0f56` as the documented
  rollback path and closes the originating issue
  (`originating_issue: 961` in `.openspec.yaml` threads
  `Closes #961` via `/uf.finale`). Record the verification
  surface accurately: the `Standardized CI / Verify action
  versions` check does not run on this PR while
  `ci_checks.yml` pins org-infra reusable CI at v0.8.0 (the
  pinact job ships in v0.9.0, on the dependabot bump PR
  #864), so this PR's evidence is the authoring-time API
  verification (tasks 1.1-1.2) plus the written pin's
  compliance with the pinact `--verify` contract; the check
  itself first runs when the v0.9.0 bump merges.
