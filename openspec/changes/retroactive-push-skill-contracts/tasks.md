<!-- [P] Tasks are parallel-eligible where marked; all work in
     section 1 predates this change and shipped in PR 948. -->

# Tasks

## 1. Implementation (already complete, council-reviewed)

- [x] Push tools require `branch` (currently checked-out
      branch, explicit refspec, `BRANCH_MISMATCH` refusal),
      validate `remote` as a nickname, and return structured
      `GitPushData` with an optional post-push `warning`
      (#947; PR 948 round-2 fixes: branch-probe distinction,
      remote validation, upstream-failure warning).
- [x] Skill resolution probes `.opencode/skills/` then a
      repo-root file fallback; directories are not file
      references; whitespace-bearing values fail with
      commands/make-targets guidance (#934).
- [x] `uf init` skill-directory migration warns on removal
      failure and keeps the file instead of counting it
      removed (#933).
- [x] Commands migrated to structured push-tool calls
      (uf.finale, uf.fix-ci, uf.rebase-pr); raw shell
      force-push forms removed and forbidden by contract
      pins.

## 2. Spec record (this change)

- [x] Write the three MODIFIED deltas (unleash UF-FR-002,
      plugin-tools Git operation tools, uf-rebase-pr Bounded
      force-push) restating the shipped contracts from the
      latest-generation delta text.
- [x] Annotate the five contradicted live deltas with inline
      superseded-by notes (delegation-skill-enforcement
      spec/design/tasks, model-matrix-delegation spec,
      finale-token-reduction plugin-tools, port-uf-rebase-pr
      spec, spawn-subtask-classification-params
      determinism-policy).
- [x] Record archive-order constraints in the delta headers,
      proposal, and design.
- [x] Amend CHANGELOG (BREAKING + re-run `uf init` guidance,
      branch-probe, remote validation, warning field, removal
      semantics; docs-issue references).

## 3. Verification

- [x] `openspec validate --strict retroactive-push-skill-contracts`
- [x] `make test` (go test -race -count=1 ./...)
- [x] `make plugin-test`
- [x] Scaffold mirror parity: canonical `.opencode/plugins/`
      and `internal/scaffold/assets/opencode/plugins/` files
      byte-identical (cmp).
