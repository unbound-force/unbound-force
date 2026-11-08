---
tier: validated
compiled_at: 2026-10-25T16:28:46Z
compiled_by: rits/zai-org/glm-5-3
sources:
  - scaffold-1
  - uf-init-20260622T203631-jay-flowers
  - uf-init-20260622T203637-jay-flowers
  - uf-init-20260622T203645-jay-flowers
  - scaffold-testing-20260812T175847-jay-flowers
  - scaffold-testing-20260812T175854-jay-flowers
  - scaffold-20260814T213248-jay-flowers
  - scaffold-20260814T213253-jay-flowers
  - scaffold-20260816T181501-jay-flowers
  - scaffold-20260816T181506-jay-flowers
  - scaffold-20260816T184148-jay-flowers
  - add-todowrite-to-pipelines-20260829T193047-jay-flowers
  - add-todowrite-to-pipelines-20260829T193101-jay-flowers
  - add-todowrite-to-pipelines-20260829T193107-jay-flowers
  - opsxuf-init-protect-block-20260929T211418-jay-flowers
  - opsxuf-init-stealth-mode-20261001T170427-jay-flowers
  - opsxuf-init-stealth-mode-20261001T170441-jay-flowers
  - dual-copy-sync-20261010T172859-jay-flowers
  - scaffold-cleanup-20261011T013729-jay-flowers
  - scaffold-drift-20261012T005353-jay-flowers
  - scaffold-template-state-20261012T011816-jay-flowers
  - convention-packs-20261014T120909-jay-flowers
  - convention-packs-20261014T120911-jay-flowers
  - convention-packs-20261014T120916-jay-flowers
  - scaffold-20261014T151515-jay-flowers
  - scaffold-drift-20261014T165537-jay-flowers
  - scaffold-drift-20261014T181320-jay-flowers
  - convention-packs-20261014T220256-jay-flowers
  - scaffold-mirrors-20261016T151128-jay-flowers
  - scaffold-parity-20261016T151133-jay-flowers
  - scaffold-staging-20261017T033610-jay-flowers
  - scaffold-parity-20261019T164211-jay-flowers
  - scaffold-parity-20261020T113858-jay-flowers
  - scaffold-governance-20261022T180325-jay-flowers
topic: scaffold
---

# Scaffold

## Current State (October 2026)

### Dual-copy architecture and drift enforcement

- Dual-copy system: canonical source tree `internal/scaffold/assets/` (embedded via `//go:embed assets` / `var assets embed.FS`, scaffold.go:32); active copies in
  `.opencode/commands/`, `.opencode/agents/`, `.opencode/skills/`, `.opencode/uf/packs/` (plus `plugins/` and `lib/`).
- `TestEmbeddedAssets_MatchSource` (scaffold_test.go:93 — underscore in the name) enforces byte-identity between embedded assets and active copies; its failure message prints
  the exact `cp` remediation for each drifted file.
- `TestCanonicalSources_AreEmbedded` walks the canonical dirs; any repo-local file not in `expectedAssetPaths` or `knownNonEmbeddedFiles` fails the test.
- Drift detection is Go tests under `make test` / `make check`. `make plugin-test` is the vitest TypeScript suite only and contains no drift tests — spec artifacts must attach
  drift verification to `make test` / `make check`.
- Review shortcut: identical `index <old>..<new>` blob-hash lines on canonical and mirror hunks of a unified diff prove byte parity; mismatched pairs are immediate, localized
  drift evidence. Quick check only — the Go drift suite is the CI gate.
- Most-flagged spec-review finding: modifying `.opencode/` files without syncing the embedded twin (all 6 reviewers flagged it in add-todowrite-to-pipelines). Proposals touching
  live plugin/command/lib files MUST include lockstep scaffold-sync tasks naming the twin paths from the start.

### Tool-owned vs custom assets

- `isToolOwned` (scaffold.go:638-689) gates auto-update on content diff. Tool-owned: `openspec/schemas/`, `uf/schemas/`, `opencode/commands/`, `opencode/skills/`,
  `opencode/package.json`, activation-gated assets, specific `lib/*.ts` files, `uf/reviewer-capabilities.yaml`, canonical convention packs.
- User-owned: agent files (including Divisor personas), `uf/review-matrix.yaml`, `uf/sibling-repos.yaml`, and `*-custom.md` packs — exempt initial deployment templates; `uf init`
  seeds the embedded copy once and never overwrites the canonical file thereafter.
- Orthogonal predicate `isNeverOverwrite` (constitution-class artifacts): user-owned AND survives `--force`; its guard clause sits before the `opts.Force` block to intercept all
  overwrite paths. `TestRun_ForceOverwrites` subtracts never-overwrite files from `expectedAssetPaths` dynamically, so new protected files adapt.

### The edit workflow

- Edit the canonical source first, `cp` to the active copy, then run `go test -run TestEmbeddedAssets_MatchSource ./internal/scaffold/`; run `make check` before review. One `cp`
  per drifted file is the fix — any forgotten file blocks the pipeline. Re-sync after every modification round, including code-review fix loops.
- Session-state hygiene: execution-checklist markers (`[x]`) in pipeline command templates mutate during runs and MUST be reset to `[ ]` before committing — in both copies,
  since baked-in state propagates to every newly scaffolded project.
- Cherry-picking scaffold syncs across branches duplicates scaffold markers (both branches add markers independently); dedup keeps only the first marker per file, applied to
  live files AND embedded copies, or `TestEmbeddedAssets_MatchSource` and `TestEmbeddedAssets_SingleMarker` fail.
- Template conventions: instruction blocks adapt terminology to each command's local structure (steps vs checklist items vs phases); insertion content uses explicit staging
  (`git add <paths>`), never `git add -A`; `<protect>` blocks open after `## Instructions` and close before final sections, with idempotency checks for existing tags.
- `/uf.init` idempotency checks must be split into independent variant checks with distinct structural markers (branch check, dirty tree, commit-before-archive, branch-switch
  confirmation) so partial application is never mistaken for complete application.

### Cross-repo canonical sources and knownNonEmbeddedFiles

- `knownNonEmbeddedFiles` (scaffold_test.go:2704) is the sanctioned registry for repo-local files not produced by `uf init`: speckit commands/templates/scripts, openspec config
  and skills, gaze agents/commands (gaze agents' canonical source is `gaze/internal/scaffold/assets/agents/`), legacy reviewer agents (Spec 019), replicator-scaffolded
  agents/commands/skills, dewey commands, DCP config (`dcp.jsonc`), determinism pack/skill.
- Cleanup gotcha: grep scaffold_test.go for removed hero/feature names — dead `knownNonScaffolded`/`knownExternalFiles` entries are a frequently missed deletion target during
  large removals.
- Governance rule: scaffold-embedded templates, plugin remediation text, and agent checklists must only reference artifacts that are themselves scaffold-embedded (scaffolded
  skills: always-on-guidance, dispatch-advisor, pre-flight, review-context, speckit-workflow, tier-routing). Repo-local artifacts cited by scaffolded sources (determinism pack,
  determinism-first skill) dangle in hero repos after `uf init`.
- Adding a convention pack triggers two gates: `TestCanonicalSources_AreEmbedded` (knownNonEmbeddedFiles entry) and `TestValidateConventionPack_AllPacksValid`
  (internal/schemas/packvalidator_test.go — non-coding packs need a skip-list entry with rationale: severity.md, content.md, ci.md, determinism.md). Cover both or `make test` fails.
- Rule-ID prefixes form one global namespace across all packs co-loaded into Divisor review prompts (readdir auto-discovery injects every pack into each prompt): CS/AP/SC/TC/DR
  (default + language packs), VB/BA/PR/TD/FA/FT (content), CI (ci). Grep `.opencode/uf/packs/` and its embedded mirror for collisions before assigning NN-NNN identifiers.
- Citation accuracy: severity.md defines only the LOW/MEDIUM spec-review auto-fix boundary; the enumerated fix-scope list is the review council's hybrid fix policy
  (uf.review-council.md). Paraphrases must cite the council command, not severity.md.

### Test plumbing and counts

- The CLI "N files processed" assertion (cmd/unbound-force/main_test.go:47-66) is a RANGE check [70,74], parsed via line-anchored regexp `(\d+) files processed$` — a plain
  substring test would let "167 files processed" satisfy the "67" case. The exact asset inventory is pinned by scaffold-level drift tests.
- Programmatically-generated config pattern (dcp.jsonc): `configureFoo(opts *Options) []subToolResult` following `configureOpencodeJSON()`, call site in `initSubTools()` via
  `collect()`, a knownNonEmbeddedFiles entry, and updates to all 5+ initSubTools result-count assertions. Raw-string constants (not embed.FS) suit small fixed-content files.
- JSONC handling: strip full-line `//` comments before `json.Unmarshal` (no parser dependency); fail safe with a malformed-JSON error result. Trivial helpers (<15 lines) may be
  inlined per package (stripDCPComments in doctor vs stripJSONCComments in scaffold); a third consumer triggers extraction to a shared package.
- Directory-walk callbacks need a hidden-file skip (`strings.HasPrefix(info.Name(), ".")`) for macOS `.DS_Store`.
- `uf init --stealth` (local-only, leaves tree git-clean): parse `git ls-files -s -z` and skip gitlink entries (stage mode 160000) — a submodule looks like a directory via
  os.Lstat but must never be os.RemoveAll'd during restore. Rollback is a first-class, baseline-diffing, fail-fast component: pre-init baseline of canonical paths/untracked/
  staged sets; restore `.git/info/exclude` before the untracked sweep; deletion restricted to the canonical scaffolded set; un-staging scoped to staged-now minus staged-at-start;
  error naming residual dirty paths.

### Mirror-sync constraints

- Mirror-synced plugin files multiply refactor surfaces: the canonical tree, the byte-identical mirror, `expectedAssetPaths`, `task42CanonicalAssets` (scaffold_test.go:1470),
  barrel-entry constraints, and closure-test import grammar.
- Dual-copy set: agent-executor.ts, invoke-agent/index.ts, review-dispatch/index.ts, uf-workflow/index.ts, uf-finale/index.ts, and command templates (uf.unleash.md, uf-init.md).
- Import-closure regexes in staging tests (plugin_activation_test.go) must cover every relative-prefix form the module layout can produce: a character class admitting `/` while
  requiring a literal `./` lets parent-relative (`../`) imports from nested modules silently escape the closure — staging gaps fail open. Decide the supported layout grammar
  first; regex and doc comment must state the same layout.

## History

- Apr 2026 (earliest learning): cherry-pick marker duplication — dual-copy coordination predates the test-enforced workflow.
- Jun 2026: /uf-init edits caught in spec review for missing embedded-copy sync; idempotency checks split into variant checks; explicit-staging rule applied to insertion content.
- Aug 2026: skills added to the reverse-drift canonicalDirs walk (forcing knownNonEmbeddedFiles registration plus a hidden-file skip); Spec 022 hivemind guard; isNeverOverwrite
  separated from isToolOwned; dcp.jsonc + JSONC parsing; TodoWrite blocks in pipeline templates. The August compile carried 5 scaffold-prefixed learnings; 20 exist as of Oct 2026.
- Hivemind migration (Spec 022): `hivemind_*` renamed to `replicator_org_*` / `replicator_comms_*` / `replicator_forge_*` / `dewey_*`, guarded by
  `TestScaffoldOutput_NoHivemindReferences` (scaffold_test.go:3348). Stale references surfaced in always-on-guidance/SKILL.md during scaffolding — update tool names before
  scaffolding files with legacy references.
- Count assertions evolved from hardcoded values ("39 files processed") to the [70,74] range with regexp parsing, as the asset inventory grew and exact lists moved to
  scaffold-level drift tests.
- knownNonEmbeddedFiles grew steadily: speckit files, then openspec config/skills, replicator agents/commands/skills, dewey commands, gaze agents (cross-repo canonical), legacy
  reviewers, DCP config, determinism pack/skill.
- Convention packs grew from default/custom to severity, content, ci, determinism — acquiring the two-gate requirement, the validator skip list, and the global rule-ID prefix
  namespace along the way.
- Oct 2026: plugin mirror-sync discipline hardened (review-dispatch/index.ts synced twice during fetch-pr-review-context), blob-hash parity review technique, staging closure
  regex fail-open fix (PR 745 round 2), scaffold-embedded-references-only governance rule.
