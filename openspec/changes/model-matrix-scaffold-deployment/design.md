## Context

Slice 1 (`model-matrix-delegation`, issue #887) built the
model-matrix selection layer and registered — but did not fully
deploy — its scaffold surface. Three facts define this slice's
starting state:

1. The embedded starter template
   `internal/scaffold/assets/uf/model-matrix.yaml` (deployed by
   every full `uf init` run since slice 1) maps five packs:
   `meta-prompt`, `go`, `typescript`, `python`, `content`. The
   matrix loader (`.opencode/lib/model-matrix.ts`,
   `loadModelMatrix`) enforces pack existence at load time
   (MM-FR-001): a `packs:` key without a `.opencode/uf/packs/
   <key>.md` file resolves to `status: "invalid"`, and the
   mirrored uf-workflow plugin fails the delegation
   non-retryably (`MATRIX_INVALID`). The full scaffold deploys
   every embedded pack (the language filter in the deploy walk
   is DivisorOnly-only), and `meta-prompt` is the one mapped
   pack absent from the embedded set — so hero repos
   scaffolded from current `main` receive a matrix that cannot
   load. `internal/schemas/packvalidator_test.go` and
   `knownNonEmbeddedFiles` both carry slice-1 entries marking
   `meta-prompt.md` and `model-matrix-classifier.md` as
   "deferred to slice 2 (#889), per SC-FR-007".

2. `isToolOwned` currently classifies `uf/model-matrix.yaml` as
   tool-owned, with a slice-1 comment stating that "a deployed
   repo's tracked copy diverging from the template is slice-2
   semantics (#889)". Slice-1 design D10 anticipated the
   resolution: "a deployed repo's tracked `.uf/model-matrix.yaml`
   is team-owned config once `uf init` writes it (slice 2), at
   which point it diverges from the template by design."

3. The ownership mechanics are the established review-council
   pattern: `uf/review-matrix.yaml` and
   `uf/reviewer-capabilities.yaml` are project-owned extension
   points preserved across plain re-init (issues #884/#886),
   `uf/schemas/<name>/` registries are tool-owned and refresh,
   and `*-custom.md` packs are user-owned, drift-exempt
   initial-deployment templates.

The meta-prompt pack (`.opencode/uf/packs/meta-prompt.md`,
MP-001 through MP-024, issue #888) is a language-agnostic
spec-authoring governance pack; `alwaysDeployedPacks` does not
list it, and its self-describing comments call it repo-local
and companion-free — both written before this slice existed.
The classifier subagent
(`.opencode/agents/model-matrix-classifier.md`) is a hidden,
least-privilege (edit/bash/webfetch denied), fully generic
subagent with no meta-repo-specific content.

Constitution alignment is assessed in the proposal (all six
principles PASS); the decisions below note the specific
principles they exercise.

## Goals / Non-Goals

### Goals

- A fresh full `uf init` produces a hero repo whose model
  matrix loads valid: every `packs:` key resolves, the
  classifier subagent exists, and `uf_delegate_command` works
  without matrix errors.
- Every hero repo receives the meta-prompt pack and its custom
  template, language-independent, listed in the scaffolded
  AGENTS.md Convention Packs section.
- Re-init behavior matches the ownership classification:
  team-edited tracked matrices, custom packs, and the override
  survive; schema registries and canonical packs refresh.
- The self-consistency of the deployed surface becomes a
  checked invariant (a guard test), not an implicit
  assumption.
- Drift and scaffold tests stay authoritative and green;
  `make check` passes end to end.

### Non-Goals

- Changing matrix semantics: loader, precedence chain,
  override whitelist, schema contents (issue #887 owns them).
- Deploying the `determinism` pack or its skill — issue #889's
  list does not include them.
- Refactoring the duplicated pack lists (`alwaysDeployedPacks`
  map vs `collectDeployedPacks` candidates) beyond extending
  both.
- Touching the gitignore block — slice 1 already added the
  override entry.
- DivisorOnly matrix deployment — the matrix serves
  content-agent delegation, not review dispatch.

## Decisions

### D1: Meta-prompt deploys as an always-deployed,
### drift-enforced pack

The canonical `meta-prompt.md` gains a byte-identical embedded
mirror (drift-enforced by `TestEmbeddedAssets_MatchSource` and
the `task42CanonicalAssets` table) and entries in
`alwaysDeployedPacks`. Rationale: the deployed starter matrix
references the pack, so without it the deployed surface is
broken, not merely incomplete — this is the slice-1 gap this
change exists to close. DivisorOnly mode also receives the
pack: DivisorOnly deploys all convention packs because Divisor
personas load every pack at review time, and the meta-prompt
pack's own intro states its rules apply at review time.
(Composability First: the pack is optional content for any
consumer; removing it degrades guidance, not function.)

### D2: A meta-prompt-custom.md template ships, and the pack's
### self-description is revised to match

Issue #889's acceptance criteria name a
`meta-prompt-custom.md` template explicitly, so the
`*-custom.md` pattern applies: canonical stub (frontmatter,
CR-NNN custom-rule prefix guidance, sentinel-terminated Custom
Rules section, no version marker — the `default-custom.md`
shape), embedded copy, user-owned, drift-exempt under the
existing custom-pack skip rule. The pack's two
self-describing comments — the header note claiming it is a
repo-local pack that is not a scaffold asset, and the Custom
Rules note claiming it ships no custom companion — become
false once deployed, so both are revised; the MP rules
themselves are untouched (comment-only revision, then
byte-identical mirror sync). The `content.md`/`content-custom.md`
pair establishes that a governance-styled pack may carry a
custom companion; project-specific meta-prompt rules in hero
repos belong in the sidecar, while this repo governs its own
pack through revision. The pack validator needs no ripple:
custom packs are already skipped there.

### D3: The classifier subagent deploys as a user-owned agent
### asset

`.opencode/agents/model-matrix-classifier.md` gains an embedded
mirror; the full deploy walk writes it as a new file, and it
stays user-owned on re-init per the agent-file precedent
("Agent files ... are user-owned"). Rationale: slice-1 SC-FR-007
and the `knownNonEmbeddedFiles` comment explicitly deferred this
file's deployment to this slice, and issue #889's outcome —
"the same content-agent model-selection capability this meta
repo has" — includes the classifier cascade's agent engine.
Without the file, hero repos silently degrade to static
defaults; with it, deleting or editing it degrades gracefully
(the cascade never blocks selection), so user ownership is
safe. The `knownNonEmbeddedFiles` entry is removed once the
mirror exists.

### D4: The deployed tracked matrix is team-owned

`isToolOwned("uf/model-matrix.yaml")` flips to false, making
the deployed matrix a project-owned extension point preserved
across plain re-init; `--force` remains the explicit replace
path; fresh deploys keep seeding the template through the
new-file creation path.

Evidence, in order of weight:

- Slice-1 design D10 (reviewed, approved): "a deployed repo's
  tracked `.uf/model-matrix.yaml` is team-owned config once
  `uf init` writes it (slice 2), at which point it diverges
  from the template by design." The scaffold.go comment slice 1
  left at the classification says the same: divergence
  semantics are "slice-2 semantics (#889)".
- The tracked file is the only home for team-shared structure
  config. The override whitelist admits only
  `profiles.*.model`, `profiles.*.variant`, and
  `classifier.llm.*` — model identities — so a team's
  profiles, `packs:` mapping, `defaults:`, and `agents:` pins
  live solely in the tracked file. A tool-owned classification
  clobbers that config on every refresh.
- The direct precedent: `uf/review-matrix.yaml` is user-owned
  for exactly this reason ("generated repositories may
  customize model selection without --force replacing their
  policy"), and issue #889 cites the #884/#886 init-survival
  work as "useful precedent for init-survival behavior".

Issue #889's re-run criterion — "preserves user edits to
custom files and correctly refreshes tool-owned matrix files
(enum entries in place)" — is satisfied under this
classification: the preserved class covers custom packs and
the team-owned tracked matrix; the refreshed class covers the
tool-owned matrix files, i.e. the `.uf/schemas/model-matrix/`
registry (whose `uf/schemas/` prefix enum entry is in place)
and the canonical packs including `meta-prompt.md`. The
alternative reading — keeping the tracked matrix tool-owned so
re-init resets team profiles — was rejected: it contradicts
D10, defeats the cited precedent, and destroys team config the
data model gives no other home. (Observable Quality: the
ownership table in `docs/configuration.md` makes the
classification inspectable; Autonomous Collaboration: teams
express policy through a tracked artifact, not runtime
coordination.)

Blast radius: drift tests are unaffected (they compare embedded
vs canonical within this repo); `TestIsToolOwned`'s entry flips
with a rewritten comment; fresh deploys are unchanged; repos
scaffolded during the slice-1 window are healed by re-init
(the pack deploys as a new file regardless of ownership).

### D5: Both pack lists are extended, not refactored

`alwaysDeployedPacks` (the deploy filter) and
`collectDeployedPacks` (the AGENTS.md listing source) duplicate
the always-deployed set. Both gain the meta-prompt entries;
the duplication itself is left as-is — a refactor would touch
map-iteration determinism and test fixtures for no behavioral
gain inside a deployment-only slice. Tests assert the meta-
prompt entries in both lists, which pins the divergence risk.

### D6: A guard test owns template/pack self-consistency

A new scaffold test parses the embedded `uf/model-matrix.yaml`
(YAML parse of the embed-FS content), iterates `packs:` keys,
and fails naming each `internal/scaffold/assets/opencode/uf/
packs/<key>.md` that is not embedded. Embedded presence is the
correct predicate because the full walk deploys every embedded
pack; DivisorOnly excludes the matrix itself, so no constraint
applies there. The test fails on current `main` (meta-prompt
missing) and passes after this change — the regression guard
for the gap this slice closes, and for any future template
edit that maps a non-embedded pack. (Observable Quality: the
invariant is machine-checked, with the offending path named.)

### D7: DivisorOnly keeps excluding the matrix and registry

Slice-1 D10 excludes `uf/model-matrix.yaml` and
`.uf/schemas/model-matrix/` from DivisorOnly deployments — the
matrix drives content-agent delegation, which DivisorOnly repos
do not run — and the mirrored lib treats an absent matrix as a
valid no-op. This slice changes nothing there; the
`TestIsDivisorAsset` entries stay false. The meta-prompt packs
deploy in DivisorOnly (D1).

## Risks / Trade-offs

- **Slice-1-window repos carry an invalid matrix** until they
  re-run `uf init`, which deploys `meta-prompt.md` as a new
  file and heals them. Teams that edited their tracked matrix
  during the window and re-ran init under the transient
  tool-owned classification may have lost those edits to the
  template refresh; the D4 flip ends that behavior. Documented
  in the CHANGELOG entry so the healing re-run is discoverable.
- **Preserved matrices drift from template improvements** — a
  hero repo that never re-inits keeps its seeded starter
  matrix forever. Accepted: identical to the review-matrix
  trade-off; `--force` is the deliberate reset path.
- **The meta-prompt pack cites `determinism.md` by identifier**
  (MP-022 cites DE-001/DE-005/DE-008), and hero repos do not
  receive that pack. Accepted degradation: the citations are
  provenance, the MUST lives in MP-022's own text, and hero
  repos that want the deeper rules can adopt the pack later.
  Deploying it here would exceed issue #889's explicit list.
- **Dual pack lists can diverge** (D5). Mitigated by tests
  asserting meta-prompt appears in both; the structural fix is
  out of scope.
- **Comment-only pack revision** (D2) must not alter rule
  semantics. Guarded by review and by the pack's own rule
  inventory (MP-001 through MP-024 unchanged); any rule edit
  would be a scope violation of the deployment-only
  constraint.
