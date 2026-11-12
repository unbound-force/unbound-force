<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the same group, (b) has no dependency
  on prior tasks in the group, and (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Prerequisites

- [ ] 1.1 Settle OQ1 (Dewey and Replicator MCP servers in CI) and
      record the decision in design.md.
- [ ] 1.2 Merge eval-infra's `feat/trajectory-assertions` into
      `unbound-force/eval-infra` main and note the commit to pin.
- [ ] 1.3 Ask an org admin to add the `ENMAAS_KEY` Actions secret
      for this repo (D3).

## 2. Land the harness (EV-FR-001 to EV-FR-004)

- [ ] 2.1 Bring `.eval/` from `eval/review-council-arg-rejection`
      onto this branch.
- [ ] 2.2 Add the eval-infra driver to `prepare-eval-env.sh` as a
      pinned, cached, explicitly error-checked fetch, and make
      `run.sh` use it when `EVAL_DRIVER` is unset (D4).
- [ ] 2.3 Apply the OQ1 decision to the per-run config written by
      `write-run-config.sh`, if it disables the MCP servers.

## 3. Local parity (EV-FR-008)

- [ ] 3.1 Add `eval` and `eval-fast` targets to the Makefile,
      outside `check`.
- [ ] 3.2 Run `make eval-fast` and confirm the three
      argument-rejection fixtures pass.

## 4. Workflow (EV-FR-005 to EV-FR-007)

- [ ] 4.1 Add `.github/workflows/ci_eval.yml`: `schedule` and
      `workflow_dispatch`, repo guard, `contents: read`, one matrix
      job per fixture, 120-minute timeout, actions and driver
      pinned by SHA (D1, D2).
- [ ] 4.2 Install the toolchain (D6) and, per OQ1, the MCP server
      binaries; cache `~/.cache/uf-eval/plugins` keyed on
      `prepare-eval-env.sh`.
- [ ] 4.3 Write each fixture's result to the job summary and upload
      the trajectory, run log and `runs.db` with `if: always()`
      (EV-FR-006).

## 5. Verification

- [ ] 5.1 Run the workflow once through `workflow_dispatch` on this
      branch and confirm every job starts, reports, and uploads.
- [ ] 5.2 After merge, review the first nightly run by hand: check
      that MCP tools behaved as decided in OQ1 and that any failure
      is a real assertion failure, not a setup error.
- [ ] 5.3 Confirm no log, summary or artifact contains the provider
      key (EV-FR-007).
- [ ] 5.4 Verify constitution alignment against proposal.md (I to
      VI) and that neither coverage floor changed.

## 6. Documentation

- [ ] 6.1 Update AGENTS.md: add `ci_eval.yml` to the CI table,
      `.eval/` to the project structure, and an evals entry under
      Testing Conventions covering how to run and add a fixture.
- [ ] 6.2 Add a CHANGELOG.md entry.
- [ ] 6.3 Delete the superseded `eval/smoke-test-constitution-check`
      branch once this merges, with the owner's OK.
