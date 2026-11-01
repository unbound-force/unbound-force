<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the same group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
  Phase checkpoints (MP-018): every phase below is terminated
  by the /uf.unleash implement-phase checkpoint — a hard-gate
  pre-flight run of the full detected CI/local suite
  (.opencode/commands/uf.unleash.md Step 5) — and the next
  phase MUST NOT begin while that checkpoint fails. Tasks
  whose text carries an explicit verification command express
  the same gate in-plan.
-->

## 1. Canonical pack assets

- [x] 1.1 Revise `.opencode/uf/packs/meta-prompt.md`
  self-describing comments only (design D2): the header note
  that calls the pack repo-local and not a scaffold asset, and
  the Custom Rules note that says it ships no custom
  companion, now describe a deployed scaffold asset (issue
  #889) with a `meta-prompt-custom.md` companion template.
  The MP-001 through MP-024 rules, their ids, tags, and
  ordering MUST remain byte-for-byte unchanged. Verify the
  file contains no literal, well-formed clarification-marker
  syntax (a marker scan must extract zero questions).

- [ ] 1.2 [P] Create `.opencode/uf/packs/meta-prompt-custom.md`
  following the `default-custom.md` shape exactly (design D2):
  frontmatter (`pack_id: meta-prompt-custom`, `language: Any`,
  `version: 1.0.0`), a Custom Rules heading with intro prose
  naming the agents and commands that load it alongside
  `meta-prompt.md`, the `CR-NNN` custom-rule prefix guidance
  with `[MUST]`/`[SHOULD]`/`[MAY]` severity indicators, a
  sentinel-terminated Custom Rules section, and no version
  marker (the deploy pass inserts markers; canonical custom
  templates carry none).

## 2. Embedded mirrors

- [ ] 2.1 Sync the three mirrors byte-identically (design
  D1/D3): copy the revised
  `.opencode/uf/packs/meta-prompt.md` and the new
  `.opencode/uf/packs/meta-prompt-custom.md` to
  `internal/scaffold/assets/opencode/uf/packs/`, and copy
  `.opencode/agents/model-matrix-classifier.md` to
  `internal/scaffold/assets/opencode/agents/
  model-matrix-classifier.md`. Verify with
  `go test -race -count=1 -run TestEmbeddedAssets_MatchSource
  ./internal/scaffold/` — the drift test now enforces the
  canonical pair (the custom template is skipped by the
  custom-pack rule; its mirror is verified in task 4.1).

## 3. Scaffold registration

All three tasks modify `internal/scaffold/scaffold.go` —
run sequentially, no `[P]`.

- [ ] 3.1 Add `meta-prompt` and `meta-prompt-custom` entries
  to `alwaysDeployedPacks` (design D1) so the pack deploys
  regardless of detected language, in full and DivisorOnly
  modes.

- [ ] 3.2 Add `meta-prompt.md` and `meta-prompt-custom.md` to
  the `collectDeployedPacks` candidates list (design D5) so
  the scaffolded AGENTS.md Convention Packs section lists
  them. Do not refactor the list duplication.

- [ ] 3.3 Flip `isToolOwned` for `uf/model-matrix.yaml` to
  false and rewrite the classification comment (design D4,
  SC-FR-010): the deployed tracked matrix is a project-owned
  extension point preserved across plain re-init (the
  `uf/review-matrix.yaml` precedent, issues #884/#886);
  `--force` remains the replace path; fresh deploys seed the
  template via the new-file path; the drift enforcement of
  the embedded template against the canonical source is
  unaffected. Remove the slice-1 "slice-2 semantics" deferral
  comment.

## 4. Test registry and behavior tests

All tasks modify `internal/scaffold/scaffold_test.go` — run
sequentially, no `[P]`.

- [ ] 4.1 Register the new assets: add
  `opencode/uf/packs/meta-prompt.md`,
  `opencode/uf/packs/meta-prompt-custom.md`, and
  `opencode/agents/model-matrix-classifier.md` to
  `expectedAssetPaths`; add `meta-prompt.md` and
  `model-matrix-classifier.md` to `task42CanonicalAssets`
  (the custom template is drift-exempt by rule, not listed);
  remove the two `knownNonEmbeddedFiles` entries slice 1
  marked as deferred to this slice
  (`.opencode/uf/packs/meta-prompt.md`,
  `.opencode/agents/model-matrix-classifier.md`) — the
  determinism entries stay. Verify with
  `go test -race -count=1 -run 'TestAssetPaths_MatchExpected|
  TestCanonicalSources_AreEmbedded|
  TestSCFR001_CanonicalAssetSurfaceMatchesMirrors'
  ./internal/scaffold/`.

- [ ] 4.2 Update `TestIsToolOwned` (SC-FR-010): flip the
  `uf/model-matrix.yaml` entry to `false` with a comment
  citing the project-extension-point rationale and design D4;
  add `opencode/uf/packs/meta-prompt.md` → `true`,
  `opencode/uf/packs/meta-prompt-custom.md` → `false`, and
  `opencode/agents/model-matrix-classifier.md` → `false`
  (canonical packs tool-owned; customs and agents
  user-owned).

- [ ] 4.3 Extend `TestIsDivisorAsset` and
  `TestShouldDeployPack`: meta-prompt.md and
  meta-prompt-custom.md are Divisor assets (`true` —
  DivisorOnly deploys all convention packs) and deploy for
  every language (`go`, `typescript`, `python`, and empty
  lang); the matrix and schema entries stay `false` in
  `TestIsDivisorAsset` (design D7).

- [ ] 4.4 Add the deployment completeness guard (SC-FR-009,
  design D6): a test that reads the embedded
  `uf/model-matrix.yaml` via the embed-FS seam, parses its
  `packs:` keys, and fails naming each missing
  `internal/scaffold/assets/opencode/uf/packs/<key>.md`.
  Sanity-check the guard by temporarily removing the
  meta-prompt mirror expectation (or reasoning from current
  `main`): the guard fails naming `meta-prompt`; with the
  mirrors in place it passes.

- [ ] 4.5 Add fresh-deploy parity assertions (SC-FR-008,
  scenario 1): a full-mode `Run()` over a temp dir must
  report `.uf/model-matrix.yaml`,
  `.uf/schemas/model-matrix/v1.schema.json`,
  `.opencode/uf/packs/meta-prompt.md`,
  `.opencode/uf/packs/meta-prompt-custom.md`, and
  `.opencode/agents/model-matrix-classifier.md` in the
  created list, and a DivisorOnly run must deploy the two
  packs while excluding the matrix and its schema.

- [ ] 4.6 Add `TestRun_ModelMatrixPreservedOnRerun`
  (SC-FR-010), following
  `TestRun_ReviewerCapabilitiesPreservedOnRerun`
  (scaffold_test.go ~line 2390): scaffold, edit the deployed
  `.uf/model-matrix.yaml` (add a profile), re-run plain
  `uf init`, and assert the matrix is skipped with content
  unchanged while an edited `meta-prompt-custom.md` is also
  preserved; cover the refresh side (a changed
  `.uf/schemas/model-matrix/v1.schema.json` or canonical
  `meta-prompt.md` deploys as updated) via the
  `TestRun_OverwriteOnDiff_ToolOwned` pattern, and the
  `--force` replace path for the team-edited matrix.

- [ ] 4.7 Update the `TestCollectDeployedPacks_*` family and
  the `TestEnsureAGENTSmdPackSection_*` expectations to
  include `meta-prompt.md` and `meta-prompt-custom.md` in the
  always-deployed candidate set and the generated AGENTS.md
  section (custom entries are filtered only when empty
  stubs, per the existing `filterEmptyCustomPacks` rule).

## 5. Documentation

- [ ] 5.1 [P] Extend `docs/configuration.md` (SC-FR-008
  ripple, design D4/D7): add a scaffold-deployment subsection
  under the Delegation Model Matrix section stating what
  `uf init` deploys (tracked matrix template, schema
  registry, meta-prompt pack and custom template, classifier
  subagent), the re-init ownership table (team-owned tracked
  matrix preserved; schema registry and canonical packs
  refreshed; custom packs and the override preserved;
  `--force` replaces), the DivisorOnly exclusion, and the
  note that re-running `uf init` heals slice-1-window repos
  whose matrix names `meta-prompt`.

- [ ] 5.2 [P] Add a `CHANGELOG.md` entry under Unreleased for
  model-matrix-scaffold-deployment: hero-repo deployment of
  the model-matrix surface (meta-prompt pack + custom
  template, classifier subagent, matrix load validity), the
  re-init ownership semantics, and the healing re-run note.

- [ ] 5.3 [P] Update the repo `AGENTS.md` Convention Packs
  list to include `.opencode/uf/packs/meta-prompt-custom.md`
  (meta-prompt.md is already listed); no other structural
  AGENTS.md changes are expected (agent count and project
  tree are unchanged).

## 6. Bookkeeping and verification

- [ ] 6.1 File the documentation issues (proposal Impact):
  one against this repo per the AGENTS.md documentation gate,
  and one against `unbound-force/website` per the org
  constitution's Cross-Repo Documentation rule, referencing
  slice 1's issue unbound-force/website#358 for the feature
  docs this deployment slice extends. Record both issue
  numbers for the PR body.

- [ ] 6.2 Run the full CI-parity gate (MP-023):
  `make check` — lint, `go test -race -count=1 ./...`,
  build, the scope-based coverage gate, and the TypeScript
  plugin suite (including its 94% statement coverage floor).
  Fix any pack-enumeration ripples the new custom template
  causes in the plugin tests. All gates MUST pass without
  modification — any gate conflict is a stop-and-report per
  the AGENTS.md gatekeeping rule.

- [ ] 6.3 Verify constitution alignment against the proposal's
  assessments: the MP-001 through MP-024 rules are unchanged
  (task 1.1 was comment-only), no quality or governance gate
  was weakened (task 6.2 evidence), the classifier subagent
  ships its deny-all permission set, and nothing
  credential-shaped is deployed. Record the verification
  outcome in the change's wrap-up notes.
