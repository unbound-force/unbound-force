# Quickstart: AGENTS.md Behavioral Guidance Injection

**Branch**: `030-agents-md-guidance` | **Date**: 2026-04-15
**Phase**: 1 (Design)

## What This Feature Does

Adds a new step to the `/uf-init` command that
automatically injects 8 standardized behavioral guidance
sections into a repo's `AGENTS.md`. This ensures every
repo in the Unbound Force ecosystem enforces the same
quality gates, workflow boundaries, and documentation
requirements.

## Implementation Summary

### Files to Modify

| # | File | Change |
|---|------|--------|
| 1 | `.opencode/command/uf-init.md` | Add Step 9 (AGENTS.md guidance injection), renumber existing Step 9 → Step 10, extend report |
| 2 | `internal/scaffold/assets/opencode/command/uf-init.md` | Sync copy of file #1 |

### What NOT to Modify

- No Go source files
- No test files
- No agent files
- No convention packs
- No CI workflows
- No constitution

## Step-by-Step Implementation Guide

### Step 1: Add Step 9 to `/uf-init`

Open `.opencode/command/uf-init.md`. Insert a new
`### Step 9: AGENTS.md Behavioral Guidance` section
between the current Step 8 (OpenSpec Command Guardrails)
and Step 9 (Report Results).

The new step should:

1. **Check** if `AGENTS.md` exists at the repo root
2. **If not**: Report `⊘ AGENTS.md: not found (skipped)`
   and skip the entire step
3. **If yes**: Read `AGENTS.md` and check for each of the
   8 guidance blocks
4. For each block:
   - Search for the detection heading/phrases
   - If found: Report `⊘ <block>: already present`
   - If not found: Find the correct insertion point and
     inject the block text. Report `✅ <block>: injected`

### Step 2: Renumber Existing Step 9

The current "Step 9: Report Results" becomes
"Step 10: Report Results".

### Step 3: Extend the Report

Add an "AGENTS.md Guidance" section to the report
template in the new Step 10:

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

### Step 4: Sync Scaffold Asset Copy

Copy the modified `.opencode/command/uf-init.md` to
`internal/scaffold/assets/opencode/command/uf-init.md`.
The files must be byte-identical.

### Step 5: Verify

1. Run `go test -race -count=1 ./...` — all tests must
   pass (especially `TestScaffoldAssetDrift`)
2. Run `git diff` to review changes
3. Verify only 2 files are modified

## Verification Checklist

- [ ] New Step 9 is present in `/uf-init`
- [ ] All 8 guidance block texts are defined inline
- [ ] Each block has detection phrases for idempotency
- [ ] `AGENTS.md` not found case is handled gracefully
- [ ] Existing Step 9 is renumbered to Step 10
- [ ] Report template includes "AGENTS.md Guidance"
- [ ] Scaffold asset copy is synchronized
- [ ] `go test -race -count=1 ./...` passes
- [ ] Only 2 files modified (live + asset copy)

## Key Design Decisions

1. **Inline text, not file references**: The guidance
   block text is defined directly in the `/uf-init`
   command instructions, not in separate template files.
   This keeps the implementation self-contained and
   avoids adding new embedded assets.

2. **Heading-based idempotency**: Detection uses Markdown
   heading text (with semantic fallback) rather than
   marker comments. This matches the existing pattern in
   Steps 2-4 and works with manually-added sections.

3. **Generalized text**: The injected text is
   repo-agnostic. Repo-specific customization is done by
   the user editing the injected text after `/uf-init`
   runs.

4. **Step 9 placement**: The AGENTS.md guidance step is
   placed after all OpenSpec/Speckit customizations
   (Steps 2-8) and before the report (Step 10). This
   groups all file-modification steps together, with the
   report as the final action.
