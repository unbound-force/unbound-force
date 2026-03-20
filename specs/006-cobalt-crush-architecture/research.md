# Research: Cobalt-Crush Architecture

**Spec**: 006-cobalt-crush-architecture
**Date**: 2026-03-20

## R1: Convention Pack Relocation Strategy

**Decision**: Move packs from `.opencode/divisor/packs/` to
`.opencode/unbound/packs/`. Update all references in Go
source, tests, agents, embedded assets, and documentation.

**Rationale**: Convention packs are now shared between
Cobalt-Crush and The Divisor. A neutral, org-level location
(`unbound`) makes clear these are framework resources, not
hero-specific artifacts. This eliminates SC-002's need for
content comparison — both heroes read the same files.

**Alternatives considered**:
- Keep at `.opencode/divisor/packs/`: Simpler (no refactor)
  but implies Divisor ownership. Cobalt-Crush reading
  "divisor" packs is confusing.
- `.opencode/packs/`: Too generic, could conflict with
  other OpenCode tooling.
- `.opencode/cobalt-crush/packs/`: Creates a second copy,
  introducing drift risk.

**Impact**: 97 references across 20 files. 12 files
physically moved. ~35 path string changes in Go tests.
Low risk — all changes are mechanical find-and-replace
with drift detection tests to catch errors.

## R2: isDivisorAsset() Refactor for Shared Packs

**Decision**: Add `isConventionPack(relPath)` call to
`isDivisorAsset()` so convention packs at the new
`opencode/unbound/packs/` location are still included
in the `--divisor` subset deployment.

**Rationale**: Convention packs are essential for The
Divisor to function. Even though they moved to a neutral
location, `--divisor` mode must still deploy them.
Calling `isConventionPack()` inside `isDivisorAsset()`
decouples the pack location from the Divisor identity.

**Alternatives considered**:
- Add `"opencode/unbound/"` prefix to `isDivisorAsset()`:
  Too broad — would include future non-Divisor assets.
- Rename `isDivisorAsset` to `isReviewAsset`: Scope creep.
  The function name can stay since `--divisor` is the flag.

## R3: Agent File Design — cobalt-crush-dev.md

**Decision**: Single Markdown file with these sections:
1. YAML frontmatter (description, mode, model, temperature,
   tools)
2. H1: Role — engineering philosophy, approach
3. H2: Source Documents — AGENTS.md, constitution, specs,
   convention packs at `.opencode/unbound/packs/`, artifacts
   at `.unbound-force/artifacts/`, graphthulhu MCP (if
   available)
4. H2: Engineering Philosophy — clean code, SOLID, TDD
   awareness, CI/CD focus, spec-driven development
5. H2: Code Implementation Checklist — convention pack
   adherence, test hook generation, documentation, error
   handling
6. H2: Gaze Feedback Loop — how to read and address Gaze
   reports
7. H2: Divisor Review Preparation — how to prepare code
   for review, address findings
8. H2: Speckit Integration — how to work with tasks.md,
   phase checkpoints, dependency ordering
9. H2: Decision Framework — how to resolve ambiguity,
   trade-offs, pattern selection
10. H2: Output Standards — code quality expectations

**Rationale**: Mirrors the Divisor agent structure (Source
Documents, checklists, output format) but with a
development focus instead of a review focus. The sections
cover all 6 user stories.

## R4: Artifact Directory Convention

**Decision**: `.unbound-force/artifacts/` as the formal
directory for inter-hero artifacts (Gaze reports, Divisor
verdicts).

**Rationale**: This is the path specified in Spec 008
(Swarm Orchestration) for artifact storage. Using it now
establishes the convention before the formal schema (Spec
009) is defined. The directory structure would be:
```
.unbound-force/artifacts/
├── quality-report/     # Gaze output
├── review-verdict/     # Divisor output
└── ...                 # Future artifact types
```

**Alternatives considered**:
- No formal directory (read stdout/files): Fragile,
  non-discoverable, varies by tool invocation.
- `.gaze/reports/` and `.divisor/verdicts/`: Hero-specific
  paths, harder to discover.

**Note**: Gaze does not currently write to this directory.
Cobalt-Crush's instructions will reference it as the
expected location, but the agent should also read Gaze
output from common locations (`coverage.out`, stdout).
Full artifact directory support is deferred to Spec 008/009.

## R5: graphthulhu MCP Integration

**Decision**: The agent file includes a conditional
instruction: "If graphthulhu MCP tools are available
(knowledge-graph_search, knowledge-graph_get_page, etc.),
use them to search for related specs, past review patterns,
and architectural decisions. Otherwise, read project files
directly."

**Rationale**: graphthulhu is already configured in
`opencode.json` for this repo (Spec 010). The agent
should leverage it when available for richer context
(cross-spec search, link traversal). But it's a SHOULD,
not a MUST — the agent must work without it.

## R6: Spec 005 Documentation Update Strategy

**Decision**: Update Spec 005 artifacts to reflect the
new pack location. These are living documents, not
historical records.

**Rationale**: Spec 005's artifacts (plan.md, tasks.md,
data-model.md, contracts/, quickstart.md) are referenced
by future implementers and the roadmap. Stale paths cause
confusion. The spec status remains "Complete" — the
refactor is an improvement, not a re-implementation.

**Scope**: ~50 references across 7 files. Mechanical
find-and-replace of `.opencode/divisor/packs/` →
`.opencode/unbound/packs/` and `opencode/divisor/packs/`
→ `opencode/unbound/packs/`.

## R7: Embedded Asset Count

**Decision**: Add 1 new embedded file (`cobalt-crush-dev.md`),
bringing the total from 45 to 46. Convention pack files
move but count stays the same (6 packs, new location).

**Impact on tests**: `expectedAssetPaths` gains 1 entry.
`cmd/unbound/main_test.go` file count assertion changes
from 45 to 46.
<!-- scaffolded by unbound vdev -->
