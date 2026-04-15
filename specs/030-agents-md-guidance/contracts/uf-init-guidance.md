# Contract: /uf-init AGENTS.md Guidance Injection

**Branch**: `030-agents-md-guidance` | **Date**: 2026-04-15
**Phase**: 1 (Design)

## Overview

This contract defines the behavioral interface for the
AGENTS.md guidance injection step in `/uf-init`. Since
this feature is Markdown-only (no Go code), the contract
describes the **observable behavior** of the AI agent
executing the command, not function signatures.

## Contract: Step 9 — AGENTS.md Behavioral Guidance

### Preconditions

1. Steps 1-8 of `/uf-init` have completed (prerequisites
   verified, customizations applied)
2. The AI agent has access to the Read and Edit tools

### Inputs

- `AGENTS.md` at the repository root (may or may not
  exist)

### Behavior

#### Case 1: AGENTS.md Does Not Exist

**When** `AGENTS.md` is not found at the repo root,
**Then** the step is skipped with the message:
> "No AGENTS.md found — skipping behavioral guidance
> injection."

No files are modified. The report shows:
```
⊘ AGENTS.md: not found (skipped)
```

#### Case 2: AGENTS.md Exists, No Guidance Sections

**When** `AGENTS.md` exists and contains none of the 8
guidance sections,
**Then** all 8 sections are injected at appropriate
locations within the document.

The report shows 8 `✅` entries:
```
✅ Core Mission: injected
✅ Gatekeeping Value Protection: injected
✅ Workflow Phase Boundaries: injected
✅ CI Parity Gate: injected
✅ Review Council PR Prerequisite: injected
✅ Website Documentation Sync Gate: injected
✅ Spec-First Development: injected
✅ Knowledge Retrieval: injected
```

#### Case 3: AGENTS.md Exists, Some Sections Present

**When** `AGENTS.md` exists and contains some (but not
all) of the 8 guidance sections,
**Then** only the missing sections are injected. Existing
sections are preserved unchanged.

The report shows a mix of `✅` and `⊘` entries.

#### Case 4: AGENTS.md Exists, All Sections Present

**When** `AGENTS.md` exists and contains all 8 guidance
sections,
**Then** no modifications are made to `AGENTS.md`.

The report shows 8 `⊘` entries:
```
⊘ Core Mission: already present (skipped)
⊘ Gatekeeping Value Protection: already present (skipped)
⊘ Workflow Phase Boundaries: already present (skipped)
⊘ CI Parity Gate: already present (skipped)
⊘ Review Council PR Prerequisite: already present (skipped)
⊘ Website Documentation Sync Gate: already present (skipped)
⊘ Spec-First Development: already present (skipped)
⊘ Knowledge Retrieval: already present (skipped)
```

### Postconditions

1. `AGENTS.md` contains all 8 guidance sections (or was
   not modified if they were already present, or does not
   exist)
2. No other files were modified by this step
3. The report summary includes an "AGENTS.md Guidance"
   section with status for each block

### Idempotency Guarantee

Running Step 9 twice on the same `AGENTS.md` produces
identical file content. The second run detects all 8
sections as already present and makes no modifications.

Formally: `f(f(x)) = f(x)` where `f` is the injection
function and `x` is the `AGENTS.md` content.

## Contract: Report Extension (Step 10)

### Preconditions

1. Step 9 has completed (guidance injection or skip)
2. All previous steps (1-8) have completed

### Behavior

The report template gains a new section between
"OpenSpec Command Guardrails" and "Summary":

```
### AGENTS.md Guidance
  [status] Core Mission: [action]
  [status] Gatekeeping Value Protection: [action]
  [status] Workflow Phase Boundaries: [action]
  [status] CI Parity Gate: [action]
  [status] Review Council PR Prerequisite: [action]
  [status] Website Documentation Sync Gate: [action]
  [status] Spec-First Development: [action]
  [status] Knowledge Retrieval: [action]
```

The "Summary" line's counters (Applied, Already present,
Errors) include the AGENTS.md guidance results.

## Contract: Scaffold Asset Synchronization

### Preconditions

1. `.opencode/command/uf-init.md` has been modified

### Behavior

`internal/scaffold/assets/opencode/command/uf-init.md`
MUST be byte-identical to `.opencode/command/uf-init.md`.

### Verification

`TestScaffoldAssetDrift` (existing test) verifies this
automatically. No new test code is needed.

## Non-Functional Requirements

- **No new Go code**: This contract is enforced by AI
  agent behavior, not compiled code
- **No new embedded assets**: The `expectedAssetPaths`
  count does not change
- **No new dependencies**: No new imports, packages, or
  external tools
- **Backward compatible**: Existing `/uf-init` behavior
  for Steps 1-8 is unchanged
