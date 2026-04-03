# Research: Hivemind-to-Dewey Memory Migration

**Branch**: `022-hivemind-dewey-migration`
**Date**: 2026-04-03

## R1: Dewey Tool Equivalents for Hivemind Operations

**Decision**: Replace `hivemind_store` with
`dewey_store_learning` and `hivemind_find` with
`dewey_semantic_search`.

**Rationale**: The Spec 021 data-model.md (Agent Migration
Map) defines the canonical tool mapping. Dewey's
`dewey_store_learning` is the new MCP tool for persisting
learnings (implemented by dewey#25). Dewey's
`dewey_semantic_search` already exists and provides
equivalent functionality to `hivemind_find` for
retrieving semantically relevant learnings.

**Alternatives considered**:
- Keep `hivemind_find` alongside `dewey_semantic_search`
  (dual-read). Rejected: creates confusion about which
  system is authoritative, and Hivemind's index will
  become stale as new learnings go to Dewey.
- Create a wrapper tool that delegates to both. Rejected:
  unnecessary complexity for a clean migration.

## R2: Graceful Degradation Pattern

**Decision**: Follow the existing 3-tier degradation
pattern documented in AGENTS.md, adapted for the
Dewey-unified world.

**Rationale**: The 3-tier pattern (Full Dewey, Graph-only,
No Dewey) is already established across the codebase.
This migration adds a sub-tier for the case where Dewey
is available but `dewey_store_learning` is not yet
implemented (dewey#25 dependency).

**Degradation tiers for learning storage**:
1. **Full Dewey**: `dewey_store_learning` available →
   store learning normally.
2. **Dewey without learning storage**: Dewey is available
   but `dewey_store_learning` tool does not exist →
   warn and skip storage. Display learnings in output.
3. **No Dewey**: Dewey not installed → warn and skip
   storage. Display learnings in output.

**Degradation tiers for learning retrieval**:
1. **Full Dewey**: `dewey_semantic_search` available →
   query for prior learnings normally.
2. **No Dewey**: Dewey not installed → skip prior
   learnings with informational note. Proceed with
   review.

**Alternatives considered**:
- Fall back to Hivemind when Dewey is unavailable.
  Rejected: perpetuates the dual-system problem. Clean
  break is better — if Dewey is unavailable, skip
  gracefully.

## R3: Scaffold Asset Synchronization Strategy

**Decision**: Update canonical files first, then copy
each modified file to its corresponding scaffold asset
path. Rely on `TestEmbeddedAssets_MatchSource` to enforce
synchronization.

**Rationale**: The scaffold engine uses `embed.FS` to
bundle asset copies. The existing drift detection test
(`TestEmbeddedAssets_MatchSource`) compares every embedded
asset against its canonical source file. Updating both
in the same commit ensures the test passes.

**Asset mapping** (canonical → scaffold):
- `.opencode/command/unleash.md` →
  `internal/scaffold/assets/opencode/command/unleash.md`
- `.opencode/agents/divisor-adversary.md` →
  `internal/scaffold/assets/opencode/agents/divisor-adversary.md`
- `.opencode/agents/divisor-architect.md` →
  `internal/scaffold/assets/opencode/agents/divisor-architect.md`
- `.opencode/agents/divisor-guard.md` →
  `internal/scaffold/assets/opencode/agents/divisor-guard.md`
- `.opencode/agents/divisor-sre.md` →
  `internal/scaffold/assets/opencode/agents/divisor-sre.md`
- `.opencode/agents/divisor-testing.md` →
  `internal/scaffold/assets/opencode/agents/divisor-testing.md`

**Alternatives considered**:
- Automate the copy with a script. Rejected: manual copy
  is fine for 6 files, and the drift test catches any
  mistakes.

## R4: Regression Test Pattern

**Decision**: Add `TestScaffoldOutput_NoHivemindReferences`
following the exact pattern of
`TestScaffoldOutput_NoGraphthulhuReferences`.

**Rationale**: The graphthulhu regression test (Spec 015)
established a proven pattern: scaffold all files to a
temp directory, walk the output, search for stale tool
reference patterns. This test prevented graphthulhu
references from reappearing after the Dewey migration.
The same pattern prevents Hivemind tool references from
reappearing.

**Stale patterns to detect**:
- `hivemind_store` — the old learning storage tool
- `hivemind_find` — the old learning retrieval tool

**Allowed patterns** (not flagged):
- `Hivemind` as a word in historical prose (e.g., Recent
  Changes entries in AGENTS.md). However, scaffold assets
  should not contain even prose Hivemind references since
  they are deployed to new repositories.

**Alternatives considered**:
- Grep-based CI check instead of Go test. Rejected: the
  Go test runs in the existing test suite, is
  discoverable via `go test`, and follows the established
  pattern.

## R5: Documentation Update Scope

**Decision**: Update AGENTS.md to describe Dewey as the
unified memory layer. Update the "Embedding Model
Alignment" section to remove the Hivemind framing.
Update the Spec 020 Recent Changes entry to remove the
"Dewey complements Hivemind" language.

**Rationale**: FR-005 requires documentation to describe
Dewey as the unified memory system, superseding the
"complements Hivemind" framing from Spec 020. The
AGENTS.md "Embedding Model Alignment" section currently
says "To ensure Swarm's Hivemind uses the same model" —
this should be updated to reflect that Dewey is the
primary consumer.

**Files requiring documentation updates**:
- `AGENTS.md`: "Embedding Model Alignment" section,
  Recent Changes entry for Spec 020, new Recent Changes
  entry for this spec.
- `internal/setup/setup.go`: Comment on line 112 says
  "Setting these env vars aligns Swarm's Hivemind" —
  update to reference Dewey. Comment on line 128 says
  "Set Ollama env vars so Swarm's Hivemind uses the
  same" — update similarly.

**Alternatives considered**:
- Leave setup.go comments unchanged (they're internal).
  Rejected: comments should be accurate. Stale comments
  create confusion for future contributors.

## R6: Dewey Learning Storage Dependency (dewey#25)

**Decision**: Implement the migration now with graceful
degradation for the case where `dewey_store_learning`
does not exist yet.

**Rationale**: The spec explicitly accounts for this
dependency (US-3, Acceptance Scenario 3). The
`/unleash` retrospective step checks for tool
availability before attempting to store. If
`dewey_store_learning` is not available, the step warns
and displays learnings in the output. This means the
migration can land before dewey#25 ships — retrieval
(US-2) works immediately since `dewey_semantic_search`
already exists, and storage (US-1) activates
automatically once dewey#25 lands.

**Alternatives considered**:
- Wait for dewey#25 to land before migrating. Rejected:
  blocks the retrieval migration (US-2) unnecessarily.
  The graceful degradation pattern handles the gap.
