<!-- originating_issue: 889 is recorded in .openspec.yaml;
     /uf.finale emits `Closes #889` in the PR body. The repo
     documentation issue filed per the AGENTS.md documentation
     gate (task 5.4) and the cross-repo `unbound-force/website`
     documentation issue per the org constitution's Cross-Repo
     Documentation rule are filed during implementation; their
     numbers are added to the PR body at that time. Slice 1's
     website issue unbound-force/website#358 tracks the
     model-matrix feature docs this deployment slice extends. -->

## Why

Slice 1 (#887, `model-matrix-delegation`) deployed the
model-matrix configuration surface to hero repos but could not
deploy everything that surface references. The tracked starter
matrix maps the `meta-prompt` convention pack, and the matrix
loader fails closed at load time when a `packs:` key has no pack
file under `.opencode/uf/packs/` (MM-FR-001). The full scaffold
deploys every embedded pack, and `meta-prompt` is the one pack
the starter matrix names that is not in the embedded set — so a
hero repo scaffolded from current `main` receives a
`.uf/model-matrix.yaml` whose only unmet reference is
`meta-prompt.md`, and every `uf_delegate_command` invocation in
that repo fails with the non-retryable `MATRIX_INVALID` error.
Deploying the pack is load-bearing, not cosmetic.

Two further deferrals land here. Slice 1's SC-FR-007 and the
`knownNonEmbeddedFiles` entries it left behind state that the
`meta-prompt` pack and the hidden `model-matrix-classifier`
subagent are canonical repo content with scaffold deployment
deferred to slice 2, issue #889 — this change. And slice 1's
design D10 anticipated that a deployed repo's tracked
`.uf/model-matrix.yaml` is team-owned config once `uf init`
writes it, diverging from the template by design; the re-init
ownership semantics for the deployed matrix are defined here.

Done means (issue #889): a fresh hero repo scaffolded by
`uf init` receives the tracked matrix, the
`.uf/schemas/model-matrix/` registry, `meta-prompt.md`, and a
`meta-prompt-custom.md` template; re-running `uf init` preserves
user and team edits to custom and project-owned files while
refreshing tool-owned ones; drift and scaffold tests are green
and `make check` passes end to end; the deployment behavior is
documented in `docs/configuration.md` with a CHANGELOG entry.

## What Changes

1. **Meta-prompt pack deployment** — the canonical
   `.opencode/uf/packs/meta-prompt.md` gains a byte-identical
   embedded mirror under
   `internal/scaffold/assets/opencode/uf/packs/` and
   `alwaysDeployedPacks` entries, so every full hero repo (and
   every DivisorOnly repo, which deploys all convention packs)
   receives it. Its two self-describing comments — the header
   note calling it a repo-local pack that is not a scaffold
   asset, and the Custom Rules note saying it ships no custom
   companion — are revised to match its new deployed status;
   the MP-001 through MP-024 rules themselves are unchanged.

2. **Meta-prompt custom template** — a new
   `.opencode/uf/packs/meta-prompt-custom.md`
   initial-deployment template follows the existing
   `*-custom.md` pattern (frontmatter, CR-NNN custom-rule
   prefix, sentinel-terminated Custom Rules stub, no version
   marker in the canonical file). It is user-owned and exempt
   from drift checks by the existing custom-pack skip rule;
   `uf init` seeds it on first run and never overwrites it.

3. **Classifier subagent deployment** — the hidden
   `.opencode/agents/model-matrix-classifier.md` subagent
   (generic, least-privilege, no meta-repo-specific content)
   gains an embedded mirror, so hero repos get the agent-engine
   half of the classifier cascade instead of silently degrading
   to static defaults. Agent files stay user-owned on re-init
   per the existing agent precedent.

4. **Deployed-matrix ownership semantics** — the deployed
   `.uf/model-matrix.yaml` becomes a project-owned extension
   point: `isToolOwned("uf/model-matrix.yaml")` flips to false,
   so a plain `uf init` re-run preserves a team-edited tracked
   matrix while `--force` still replaces it — the
   `uf/review-matrix.yaml` and `uf/reviewer-capabilities.yaml`
   init-survival precedent (#884/#886) that issue #889 cites.
   Fresh deploys keep seeding the template. The
   `.uf/schemas/model-matrix/` registry and canonical packs
   (including `meta-prompt.md`) remain tool-owned and refresh
   when content differs. Issue #889's re-run acceptance
   criterion is satisfied under this classification: custom
   files and the team-owned tracked matrix are preserved, and
   the tool-owned matrix files — the schema registry, whose
   prefix enum entry is in place — refresh correctly.

5. **AGENTS.md pack listing** — `collectDeployedPacks` gains
   the two meta-prompt entries alongside the
   `alwaysDeployedPacks` map (both lists are extended; the
   pre-existing duplication between them is noted in the design
   and left unrefactored to keep this change minimal), so the
   scaffolded AGENTS.md Convention Packs section tells agents
   to read the new pack.

6. **Deployment completeness guard** — a new scaffold test
   parses the embedded `uf/model-matrix.yaml` template and
   fails, naming the offending pack, when any `packs:` key has
   no embedded pack file. This is the regression test for the
   slice-1 gap: it fails on current `main` (meta-prompt
   missing) and passes once this change lands, and it keeps any
   future template edit from shipping the same break.

7. **Test-registry updates** — `expectedAssetPaths`,
   `task42CanonicalAssets`, `TestIsToolOwned`,
   `TestIsDivisorAsset`, `TestShouldDeployPack`, the
   `TestCollectDeployedPacks_*` family, and the
   `TestEnsureAGENTSmdPackSection_*` expectations gain the new
   assets and the flipped ownership entry; the two
   `knownNonEmbeddedFiles` entries slice 1 marked as deferred
   to this slice are removed. A new
   `TestRun_ModelMatrixPreservedOnRerun` follows the
   `TestRun_ReviewerCapabilitiesPreservedOnRerun` precedent.

8. **Docs and bookkeeping** — `docs/configuration.md` gains a
   scaffold-deployment subsection (what `uf init` deploys, the
   re-init ownership table, the meta-prompt and classifier
   deployment); `CHANGELOG.md` gains an entry; the repo
   `AGENTS.md` convention-pack list gains
   `meta-prompt-custom.md`; a repo documentation issue and a
   cross-repo `unbound-force/website` documentation issue are
   filed per the AGENTS.md documentation gate and the org
   constitution's Cross-Repo Documentation rule.

Out of scope, per the issue's deployment-only constraint: the
matrix loader, selection precedence, override whitelist, and
schema semantics (#887 owns them); deploying the `determinism`
pack or its skill (not in #889's list — the meta-prompt pack's
by-identifier citations of it remain provenance references, a
documented, accepted degradation); and the gitignore block,
which slice 1 already extended with the override entry.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `scaffold`: the deployed asset surface gains the meta-prompt
  pack (canonical + custom template), the model-matrix
  classifier subagent, and the AGENTS.md pack-section entries;
  the deployed tracked matrix is reclassified as a
  project-owned extension point preserved across re-init; a
  deployment completeness guard ties every embedded template
  `packs:` key to an embedded pack file.

### Removed Capabilities

- None. DivisorOnly deployments keep excluding the matrix and
  its schema registry (slice-1 design D10); the new packs
  deploy there because DivisorOnly deploys all convention
  packs.

## Impact

- **New files**:
  `.opencode/uf/packs/meta-prompt-custom.md`;
  `internal/scaffold/assets/opencode/uf/packs/meta-prompt.md`;
  `internal/scaffold/assets/opencode/uf/packs/meta-prompt-custom.md`;
  `internal/scaffold/assets/opencode/agents/model-matrix-classifier.md`.
- **Modified**: `.opencode/uf/packs/meta-prompt.md`
  (comment-only revision, then mirrored byte-identically);
  `internal/scaffold/scaffold.go` (`alwaysDeployedPacks`,
  `collectDeployedPacks`, `isToolOwned` flip with rewritten
  comment); `internal/scaffold/scaffold_test.go` (asset
  registries, ownership and deployment tables, new guard and
  re-run tests, `knownNonEmbeddedFiles` removals);
  `docs/configuration.md`; `CHANGELOG.md`; repo `AGENTS.md`
  (convention-pack list).
- **Unchanged**: the matrix loader and selection chain
  (`.opencode/lib/model-matrix.ts`), the uf-workflow plugin,
  the schema registry contents, the gitignore block, and every
  existing ownership classification other than the deployed
  tracked matrix.
- **Behavior**: hero repos scaffolded or refreshed by
  `uf init` after this change load a valid matrix out of the
  box (the `meta-prompt` reference resolves), get the full
  classifier cascade, and keep team edits to the tracked
  matrix across re-init. Repos scaffolded during the slice-1
  window are healed by a re-run: the pack file deploys as a new
  file regardless of ownership, making their existing matrix
  valid.
- **Security**: no credential-shaped content is deployed; the
  classifier subagent is read-only (edit, bash, and webfetch
  denied) and hidden from direct invocation.
- **Gates**: no coverage floor, review limit, CI flag, or
  pinned dependency changes; the `.opencode/` suite must hold
  its governance-controlled 94% statement coverage floor, and
  `make check` (lint, test, build, coverage gate, plugin tests)
  must pass end to end.
- **Documentation gate**: user-facing change — the repo
  documentation issue and the cross-repo website issue are
  filed before PR merge (task 5.4).

## Constitution Alignment

Assessed against the Unbound Force org constitution (v1.4.0,
amended 2026-10-14; six active principles).

### I. Autonomous Collaboration

**Assessment**: PASS

The deployment ships self-describing artifacts — a convention
pack agents are told to read via the scaffolded AGENTS.md
section, a matrix template whose comments document the
tracked/override split — with no runtime coupling between
heroes. Ownership classification is expressed in the scaffold's
existing artifact-driven re-init contract (create, skip,
update), not in any cross-repo coordination.

### II. Composability First

**Assessment**: PASS

Every deployed piece remains independently removable: an absent
matrix is a valid no-op (MM-FR-004), an absent classifier agent
degrades the cascade to static defaults without blocking, and
the meta-prompt custom template is an optional extension point.
The meta-prompt pack is deployed as-is with no new mandatory
dependencies; its by-identifier citations of the determinism
pack (not deployed here) remain readable provenance.

### III. Observable Quality

**Assessment**: PASS

Deployment is machine-verified: drift tests name mismatched
mirrors, the new completeness guard names missing pack files,
and the re-run survival tests assert the create/skip/update
classification per file. The scaffold's existing result summary
surfaces each deployed asset, so what a hero repo received is
inspectable without session reconstruction.

### IV. Testability

**Assessment**: PASS

Every requirement is verifiable in isolation with the existing
scaffold test seams (`Options.ReadFile`/`WriteFile` injection,
`assetContent`, embed-FS walks): the guard test parses the
embedded template, the re-run tests exercise real `Run()`
invocations against temp directories, and no external service
is involved. The ownership flip is a pure predicate change
covered by a table-driven test.

### V. Security by Default

**Assessment**: PASS

No secrets or model identities are deployed — the tracked
template carries structure only, the override stays gitignored
and unscaffolded beyond the gitignore line slice 1 added, and
the classifier subagent ships a deny-all permission set (edit,
bash, webfetch denied) with `hidden: true`.

### VI. Deterministic Execution

**Assessment**: PASS

Deployment is a deterministic file walk with explicit
ownership classification — identical inputs produce identical
deployed trees, and the classification itself is pinned by
table-driven tests so behavior cannot drift silently. The
completeness guard turns an implicit cross-file assumption
(the template's pack references resolve) into a checked
invariant.
