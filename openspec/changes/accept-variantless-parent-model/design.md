## Context

`validateRunModel` in review-dispatch finalization rejected
successful host and host-resolved advisor runs unless both
`resolved_parent_model` and `resolved_parent_variant` were non-null,
while the writer contract (`resolveHostModel()`, commit 716f2036)
and the persistence layer (`NullableVariantSchema`,
`host.variant ?? null`) sanction a null variant. Variantless
providers such as `enmaas/rits/zai-org/glm-5-3` therefore produced
dispatches that finalization fail-closed on (observed on
unbound-force/dewey PR #135). The code fix is shipped, tested, and
approved by all four council reviewers; this change is the spec
record that keeps the unarchived requirement stack from
reintroducing the bug.

No synced `openspec/specs/review-dispatch/` exists — the unarchived
delta stack is the only requirement record. Two older unarchived
changes carry over-strict both-non-null text:
`advisor-host-model-fallback` under its MODIFIED 'Finalization
Model Validation' requirement, and
`review-council-multi-model-fanout` under its ADDED 'Review
Dispatch Artifact' requirement (RD-FR-008, the "after successful
lookup, both MUST be non-null" host-lookup clause). The
iteration-2 council review requested three artifact remediations:
a mechanically sound supersession strategy, both
documentation-gate issue citations, and the convention-pack
artifact set (design.md; the REMOVED Requirements section). The
iteration-3 review found the FMV half of that supersession
strategy unsound — repo-wide, no delta ever ADDEDs 'Finalization
Model Validation' (fanout ADDEDs RD-FR-001 through RD-FR-008
only), so no archive ordering can ever materialize the
requirement a MODIFY needs — and demanded the ADDED reshape
recorded here.

## Goals / Non-Goals

### Goals

- Record the corrected finalization contract so the unarchived
  requirement record matches the shipped validator
- Land the corrected contract over BOTH over-strict predecessor
  texts mechanically: establish 'Finalization Model Validation'
  outright via ADDED (no predecessor text can ever exist to
  modify), supersede the fanout 'Review Dispatch Artifact' text
  via same-name MODIFIED, plus an explicit archive-ordering
  constraint
- Satisfy the meta-prompt pack artifact set: proposal.md,
  design.md, tasks.md, and one capability delta organized under
  all three delta section headings (MP-012, MP-013)
- Cite both documentation-gate tracking issues (in-repo
  unbound-force/unbound-force#891 and unbound-force/website#360)
  in every artifact claiming gate completeness

### Non-Goals

- Changing code, tests, or the scaffold mirror — approved by all
  four reviewers and untouched by this change
- Editing the two older unarchived changes in place — the
  council's chosen remedy is this change's delta plus the
  ordering constraint, not amending pending changes. The orphan
  FMV MODIFIED in `advisor-host-model-fallback` stays that
  change's own defect to repair
- Allocating new FR identifiers — the ADDED 'Finalization
  Model Validation' entry reuses the requirement name the
  unarchived stack already carries (no FR id was ever allocated
  to it anywhere, and none is invented here); RD-FR-008 is the
  fanout change's allocation and is reused, not renumbered
  (MP-002)

## Decisions

### D1: Relax resolved_parent_variant only

**Decision**: The recorded contract rejects a successful host or
host-resolved advisor run only when `resolved_parent_model` is
null; `resolved_parent_variant` conforms to the variant-or-null
schema contract and MAY be null when the provider reports no
variant. The failure-message expectation is "requires resolved
parent model".

**Rationale**: Validator, persistence, and schema parity.
`resolveHostModel()` in `.opencode/lib/agent-executor.ts` (commit
716f2036, absent-variant fallback) legitimately succeeds with no
variant when the provider does not populate one; the persistence
layer writes `host.variant ?? null` under `NullableVariantSchema`;
and the shipped `validateRunModel` enforces exactly the
model-non-null rule. Identity enforcement is unchanged —
`resolved_parent_model` stays required non-null and is
cross-checked against `reported_model` (a reported-model mismatch
makes the run failed), so the essential identity control survives
the relaxation.

**Constitution alignment**: Observable Quality — the requirement
states exactly what the machine-parseable run records carry.
Security by Default — the identity control is untouched; only the
redundant variant half of a both-non-null check is relaxed, on
tool-owned local artifacts.

### D2: ADDED FMV establishment plus RDA supersession and ordering

**Decision**: One ADDED entry and one MODIFIED entry in one
delta. The ADDED 'Finalization Model Validation' entry
establishes the corrected requirement outright. The MODIFIED
'Review Dispatch Artifact' (RD-FR-008) supersedes the
over-strict host-lookup text that
`review-council-multi-model-fanout` carries under its ADDED
requirement of that name. Plus an explicit ordering constraint,
stated in the delta preamble comment, the 'Finalization Model
Validation' establishment paragraph, the proposal (What Changes
and Impact), and this design: this change MUST archive after
`review-council-multi-model-fanout`.

**Rationale**: Requirement-name matching is the only
supersession mechanism, and requirement existence is a
precondition for it — archive applies a MODIFIED delta by
replacing the same-named requirement at that change's own
archive and hard-aborts when the requirement does not exist in
the synced spec (verified mechanically in a scratch project).
The iteration-2 premise — "fanout's ADDED requirements create
the capability spec on archive, after which this change's
MODIFY operations can apply" — conflated spec-existence with
requirement-existence. It is true for the RDA half: fanout
ADDEDs RD-FR-008, so the MODIFIED restatement applies once this
change archives after fanout. It is false for FMV, which no
delta ever ADDs — fanout ADDEDs RD-FR-001 through RD-FR-008
only, the fallback change carries FMV only as a MODIFIED entry
(itself an orphan), and no canonical
`openspec/specs/review-dispatch/` exists. No archive ordering
can ever materialize a predecessor FMV text, so the ADDED form
is the only mechanically sound way to land the corrected FMV
contract; it also retires the MP-013 Previously-form question
for this entry outright (ADDED entries carry no Previously
paragraph) and drops the fallback change from the ordering
constraint — only fanout must precede, for spec creation plus
the RDA base. The fallback change's orphan FMV MODIFIED is a
pre-existing stack-wide defect in `advisor-host-model-fallback`,
out of scope here: this change neither edits it in place nor
depends on it.

The FMV entry's shape also obeys a mechanical budget:
`openspec validate --strict` fails on an overlong description
(>500 characters) for ADDED requirements (MODIFIED is exempt so
existing text stays whole), so the description states only the
three-branch contract and the rejection rule, and the
writer-contract rationale plus the establishment record sit in
prose after the scenarios — the same placement the MODIFIED
'Review Dispatch Artifact' entry uses for its Previously
paragraph, outside the measured description. The scenario set
is the complete union of both lineages: the predecessor
scenario names from the fallback delta ('Host-resolved advisor
passes validation', 'Host-resolved advisor without resolved
parent fails') with corrected bodies, plus the iteration-2
variantless scenarios (host and host-resolved-advisor
variantless passes; null-model rejection). The null-model
rejection appears under both names because the fallback delta
and the iteration-2 fix loop each named it independently;
carrying both keeps name continuity with each record. Every
scenario maps to an existing green test in
`.opencode/test/finalize-error-boundaries.test.ts` or
`.opencode/test/finalize-review-dispatch.test.ts`.

### D3: Full restatement of 'Review Dispatch Artifact' (RD-FR-008)

**Decision**: The MODIFIED 'Review Dispatch Artifact' entry
restates the requirement the fanout delta ADDED, in full —
envelope, closed payload schema, provenance, statuses, verdict
semantics, publication, and all seven existing scenarios —
changing only the host-lookup provenance clause (and the advisor
clause it corrects), and adding two variantless-success scenarios.
The requirement ends with a Previously paragraph stating the
contract it replaces (MP-013).

**Rationale**: A MODIFIED requirement replaces the whole block at
archive. Verified mechanically in a scratch OpenSpec project:
`openspec archive` replaced the target requirement with the
MODIFIED delta's text verbatim — an untouched paragraph of the
original requirement was deleted — and validation refuses a
MODIFIED block that omits scenarios the current spec carries ("a
MODIFIED requirement replaces the whole block, so archive refuses
to drop them"). A provenance-only entry would therefore delete
RD-FR-008's schema and publication text at archive. The
`review-dispatch-diff-manifest` change set the precedent: its
MODIFIED 'Per-Run Timeout Derivation' (RD-FR-028) restates the
requirement `derive-per-run-timeout` ADDED, in full.

## Risks / Trade-offs

### Risk: cross-change ordering has no mechanical guard

Archive applies each unarchived delta at that change's own archive
with no cross-change ordering guard, so the corrected RDA text
survives only if this change archives after fanout (learning
`spec-governance-20261031T221006-jay-flowers`).

**Mitigation**: The ordering constraint is stated in four
discoverable places — the delta preamble comment, the
'Finalization Model Validation' establishment paragraph, the
proposal (What Changes and Impact), and this design — and the
openspec-archive-change workflow surfaces delta sync state before
archiving, so the operator meets an explicit refusal rather than
silently regressing the canonical spec.

### Risk: a supersession claim must name the exact requirement

A same-name MODIFIED delta cannot supersede text under a different
requirement name; a blanket "supersedes the over-strict text in
change X" claim fails mechanically when the target text lives
under another heading (learning
`openspec-20261031T221010-jay-flowers`).

**Mitigation**: The RDA target was verified by grepping its
requirement heading, the MODIFIED entry in this delta carries the
same name, and the establishment paragraph names the target
requirement explicitly. The FMV half makes no supersession claim
at all — it establishes, because no target text can ever exist.

### Risk: the fallback change can no longer archive cleanly

The orphan FMV MODIFIED in `advisor-host-model-fallback` fails
closed in every archive order relative to this change. Before
this change archives, no FMV requirement exists for its MODIFY to
apply against (orphan hard abort). After this change archives,
the ADDED requirement exists, so its MODIFY would apply — but its
block carries only its two scenario names and would drop this
change's variantless scenarios, which archive refuses
(scenario-discontinuity abort). Either way that change must
repair its own delta before archiving.

**Mitigation**: The defect is pre-existing in
`advisor-host-model-fallback` and out of scope here; this change
documents it (D2, the delta preamble, the FMV establishment
paragraph) rather than editing the older change in place. The
fail-closed abort means the canonical spec can never silently
regress to the over-strict text.

### Risk: restatement drift until archive

The full RD-FR-008 restatement duplicates the fanout delta's
ADDED text until this change archives; if fanout were amended
before either change archives, the copies could diverge.

**Mitigation**: The restatement was extracted verbatim from the
committed fanout delta with only the provenance clause rewritten,
and archive's scenario-continuity check refuses a MODIFIED block
that drops scenarios the current spec carries, so silent drift
cannot pass archive.

### Trade-off: FMV description budget and dual-lineage names

ADDED requirements carry a 500-character description budget
(`openspec validate --strict` fails on the overlong-description
warning for ADDED entries; MODIFIED is exempt so existing text
stays whole), so the FMV entry states the three-branch contract
and the rejection rule in its description and moves the
writer-contract rationale and establishment record to prose after
the scenarios — the placement the MODIFIED 'Review Dispatch
Artifact' entry uses for its Previously paragraph, outside the
measured description. The null-model rejection appears under two
scenario names (the fallback delta's and the fix loop's) with
equivalent bodies because each lineage named it independently;
dropping either name would break continuity with one record while
the union keeps both. The ADDED form also retires the iteration-2
Previously-placement deviation outright — ADDED entries carry no
Previously paragraph (MP-013 governs MODIFIED entries only), and
the MODIFIED 'Review Dispatch Artifact' entry follows MP-013's
letter.
