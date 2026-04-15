# Research: AGENTS.md Behavioral Guidance Injection

**Branch**: `030-agents-md-guidance` | **Date**: 2026-04-15
**Phase**: 0 (Research)

## R1: Canonical Text Sources

The 8 guidance blocks are derived from two existing
`AGENTS.md` files in the Unbound Force ecosystem. This
section documents the exact source locations and the
generalization decisions for each block.

### Source 1: Meta Repo AGENTS.md

**File**: `unbound-force/unbound-force/AGENTS.md`

| Block | Location | Generalization Needed |
|-------|----------|----------------------|
| Core Mission | Lines 12-16 (`## Core Mission`) | None — already generic |
| Gatekeeping Value Protection | Lines 25-38 (`### Gatekeeping Value Protection`) | Minimal — remove meta-repo-specific examples |
| Workflow Phase Boundaries | Lines 40-54 (`### Workflow Phase Boundaries`) | Minimal — `specs/NNN-*/` is universal |
| Website Documentation Sync Gate | Lines 397-418 (`### Website Documentation Gate`) | None — already generic with `gh` template |
| Knowledge Retrieval | Lines 533-592 (`## Knowledge Retrieval`) | Minimal — tool matrix is universal |

### Source 2: Gaze AGENTS.md

**File**: `unbound-force/gaze/AGENTS.md`

| Block | Location | Generalization Needed |
|-------|----------|----------------------|
| CI Parity Gate | Line 27 (`- **CI Parity Gate**:`) | None — already generic |
| Review Council PR Prerequisite | Lines 38-54 (`### Review Council as PR Prerequisite`) | None — already generic |
| Spec-First Development | Lines 56-82 (`## Spec-First Development`) | Minimal — remove Gaze-specific spec examples |
| Core Mission | Lines 12-16 (`## Core Mission`) | Identical to meta repo |

## R2: Current `/uf-init` Structure

The `/uf-init` command currently has 9 steps:

| Step | Name | Target Files | Action |
|------|------|-------------|--------|
| 1 | Check Prerequisites | `.opencode/`, skills, commands | Verify existence |
| 2 | Apply Branch Enforcement | 6 OpenSpec files | Insert branch management |
| 3 | Apply Dewey Context | 7 OpenSpec files | Insert Dewey queries |
| 4 | Apply 3-Tier Degradation | 3 skill files | Insert fallback tiers |
| 5 | Speckit Custom Commands | 4 command files | Create if missing |
| 6 | Speckit Command Guardrails | 9 command files | Append guardrails |
| 7 | Speckit UF Customizations | 5 command files | Verify (read-only) |
| 8 | OpenSpec Command Guardrails | 1 command file | Append guardrails |
| 9 | Report Results | N/A | Display summary |

The new AGENTS.md guidance injection fits naturally as
**Step 9** (before the report), with the existing Step 9
renumbered to **Step 10**.

### Pattern Analysis

All existing steps follow the same pattern:
1. Read target file
2. Check if customization is already present (idempotent)
3. If not present, insert at correct location
4. Report status with `✅`/`⊘`/`❌` indicators

The AGENTS.md guidance step follows this identical
pattern, ensuring consistency with the rest of `/uf-init`.

## R3: Idempotency Detection Strategy

### Existing Patterns in `/uf-init`

| Step | Detection Method |
|------|-----------------|
| Branch Enforcement | Semantic: look for `opsx/<name>`, `git checkout -b opsx/` |
| Dewey Context | Tool invocation: `dewey_semantic_search` or `dewey_search` |
| 3-Tier Degradation | Keywords: "Tier 1", "Tier 2", "Tier 3", "graceful degradation" |
| Speckit Commands | File existence: check if `.opencode/command/speckit.*.md` exists |
| Speckit Guardrails | Heading: `## Guardrails` |
| OpenSpec Guardrails | Heading: `## Guardrails` |

### Recommended Detection for Guidance Blocks

Use **heading-based detection** (matching the Guardrails
pattern) as the primary strategy, with **semantic
fallback** for blocks that might exist under different
headings.

| Block | Primary Detection | Semantic Fallback |
|-------|-------------------|-------------------|
| Core Mission | `## Core Mission` heading | "Strategic Architecture" + "Outcome Orientation" |
| Gatekeeping Value Protection | `Gatekeeping Value Protection` in heading | "MUST NOT modify values that serve as quality" |
| Workflow Phase Boundaries | `Workflow Phase Boundaries` in heading | "MUST NOT cross workflow phase boundaries" |
| CI Parity Gate | `CI Parity Gate` in heading or bold text | "replicate the CI checks locally" |
| Review Council PR Prerequisite | `Review Council` in heading | "/review-council" + "APPROVE" |
| Website Documentation Sync Gate | `Website Documentation` in heading | "gh issue create --repo" + "website" |
| Spec-First Development | `Spec-First Development` in heading | "preceded by a spec workflow" |
| Knowledge Retrieval | `## Knowledge Retrieval` heading | "dewey_semantic_search" tool reference |

## R4: Insertion Point Strategy

Each guidance block has a natural placement within a
typical `AGENTS.md` structure:

| Block | Placement | Rationale |
|-------|-----------|-----------|
| Core Mission | After `## Project Overview`, before `## Behavioral Constraints` | Sets strategic context early |
| Gatekeeping Value Protection | Inside `## Behavioral Constraints` (as subsection) | Behavioral rule |
| Workflow Phase Boundaries | Inside `## Behavioral Constraints` (as subsection) | Behavioral rule |
| CI Parity Gate | Inside `## Technical Guardrails` or `## Behavioral Constraints` | Technical behavioral rule |
| Review Council PR Prerequisite | After behavioral constraints, before build commands | Process rule |
| Website Documentation Sync Gate | Near documentation validation gate or spec commit gate | Documentation process |
| Spec-First Development | After behavioral constraints, before build commands | Development process |
| Knowledge Retrieval | After coding conventions, before testing conventions | Tool usage guidance |

**Key insight**: The AI agent executing `/uf-init` uses
LLM reasoning to find insertion points. The instructions
should describe the *concept* of where to insert (e.g.,
"after the project overview section") rather than exact
line numbers, because each repo's `AGENTS.md` has a
different structure.

## R5: Generalization Analysis

### Blocks Requiring No Generalization

These blocks are already repo-agnostic:

1. **Core Mission** — 3 bullets about strategic framing.
   No repo-specific content.
2. **CI Parity Gate** — "Read `.github/workflows/`" is
   universal. No repo-specific commands.
3. **Review Council PR Prerequisite** — References
   `/review-council` command (available in all repos
   via `uf init`). No repo-specific content.

### Blocks Requiring Minimal Generalization

4. **Gatekeeping Value Protection** — The 8 protected
   categories are universal. The "What to do instead"
   instruction is universal. Remove any meta-repo-specific
   examples.
5. **Workflow Phase Boundaries** — `specs/NNN-*/` is the
   universal convention. Phase-to-output mapping is
   universal.
6. **Website Documentation Sync Gate** — The `gh issue
   create` template is universal. Exemption list is
   universal.
7. **Spec-First Development** — "What requires a spec"
   list needs generalization: remove Gaze-specific items
   (e.g., "embedded assets under `internal/scaffold/
   assets/`") and use generic descriptions.
8. **Knowledge Retrieval** — Tool selection matrix is
   universal. 3-tier degradation is universal.

## R6: Cross-Repo Audit

Current state of the 8 guidance sections across Unbound
Force repos:

| Block | meta | gaze | website | dewey | homebrew-tap | replicator |
|-------|:----:|:----:|:-------:|:-----:|:------------:|:----------:|
| Core Mission | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ |
| Gatekeeping Protection | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| Phase Boundaries | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| CI Parity Gate | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ |
| Review Council | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ |
| Website Doc Sync | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| Spec-First Dev | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ |
| Knowledge Retrieval | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |

**Observation**: No single repo has all 8 sections. The
meta repo has 4, Gaze has 3 (different set), and the
remaining 4 repos have none. This validates the need for
standardized injection.

## R7: File Modification Scope

### Files Modified

| File | Change Type | Lines Added |
|------|------------|-------------|
| `.opencode/command/uf-init.md` | Add Step 9 (AGENTS.md guidance), renumber Step 9→10 | ~200-250 |
| `internal/scaffold/assets/opencode/command/uf-init.md` | Sync copy | Same |

### Files NOT Modified

- No Go source files (`*.go`)
- No test files (`*_test.go`)
- No agent files (`*.md` under `.opencode/agents/`)
- No convention packs (`*.md` under `.opencode/uf/packs/`)
- No CI workflows (`.github/workflows/`)
- No schema files (`schemas/`)
- No constitution (`.specify/memory/constitution.md`)

### Test Impact

- `TestScaffoldAssetDrift` — will verify the asset copy
  stays in sync (existing test, no changes needed)
- `expectedAssetPaths` — no change (no new embedded
  assets added, just modifying existing `uf-init.md`)
- All existing tests pass without modification (FR-008)
